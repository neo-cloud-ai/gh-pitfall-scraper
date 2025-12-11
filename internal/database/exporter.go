package database

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

// ExportFormat 导出格式类型
type ExportFormat string

const (
	FormatJSON    ExportFormat = "json"
	FormatCSV     ExportFormat = "csv"
	FormatMarkdown ExportFormat = "md"
)

// ExportFilter 导出过滤器
type ExportFilter struct {
	DateFrom        *time.Time    `json:"date_from,omitempty"`
	DateTo          *time.Time    `json:"date_to,omitempty"`
	Categories      []string      `json:"categories,omitempty"`
	Tags            []string      `json:"tags,omitempty"`
	Repositories    []string      `json:"repositories,omitempty"`
	States          []string      `json:"states,omitempty"`
	MinScore        *float64      `json:"min_score,omitempty"`
	MaxScore        *float64      `json:"max_score,omitempty"`
	IsPitfall       *bool         `json:"is_pitfall,omitempty"`
	IsDuplicate     *bool         `json:"is_duplicate,omitempty"`
	IncludeMetadata bool          `json:"include_metadata,omitempty"`
}

// ExportResult 导出结果
type ExportResult struct {
	TotalRecords   int           `json:"total_records"`
	ExportedRecords int          `json:"exported_records"`
	Format         ExportFormat  `json:"format"`
	OutputPath     string        `json:"output_path"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	Duration       time.Duration `json:"duration"`
	Filter         ExportFilter  `json:"filter"`
}

// Exporter 数据导出器
type Exporter struct {
	db      *sql.DB
	logger  *log.Logger
	outputs map[ExportFormat]OutputWriter
}

// OutputWriter 输出写入器接口
type OutputWriter interface {
	WriteHeader() error
	WriteRecord(issue *Issue) error
	WriteFooter() error
	Close() error
}

// JSONWriter JSON格式输出写入器
type JSONWriter struct {
	file     *os.File
	encoder  *json.Encoder
	first    bool
}

// CSVWriter CSV格式输出写入器
type CSVWriter struct {
	writer   *csv.Writer
	file     *os.File
	headers  []string
}

// MarkdownWriter Markdown格式输出写入器
type MarkdownWriter struct {
	file     *os.File
	tmpl     *template.Template
	issueTmpl string
}

// NewExporter 创建新的导出器
func NewExporter(db *sql.DB) *Exporter {
	exporter := &Exporter{
		db:     db,
		logger: log.New(log.Writer(), "[Exporter] ", log.LstdFlags),
		outputs: make(map[ExportFormat]OutputWriter),
	}
	
	return exporter
}

// ExportIssues 导出Issues数据
func (e *Exporter) ExportIssues(filter ExportFilter, format ExportFormat, outputPath string) (*ExportResult, error) {
	startTime := time.Now()
	
	// 构建查询
	query, params := e.buildExportQuery(filter)
	
	// 执行查询
	rows, err := e.db.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query issues for export")
	}
	defer rows.Close()
	
	// 创建输出写入器
	writer, err := e.createOutputWriter(format, outputPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create output writer")
	}
	
	// 写入数据
	exportedCount, err := e.writeIssues(writer, rows)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write issues")
	}
	
	// 关闭写入器
	if err := writer.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close output writer")
	}
	
	endTime := time.Now()
	result := &ExportResult{
		TotalRecords:   e.getTotalCount(filter),
		ExportedRecords: exportedCount,
		Format:         format,
		OutputPath:     outputPath,
		StartTime:      startTime,
		EndTime:        endTime,
		Duration:       endTime.Sub(startTime),
		Filter:         filter,
	}
	
	e.logger.Printf("导出完成: %d 条记录导出到 %s", exportedCount, outputPath)
	return result, nil
}

// buildExportQuery 构建导出查询
func (e *Exporter) buildExportQuery(filter ExportFilter) (string, []interface{}) {
	query := `
		SELECT i.id, i.number, i.title, i.body, i.url, i.state, 
			   i.author_login, i.labels, i.assignees, i.milestone, 
			   i.reactions, i.created_at, i.updated_at, i.closed_at,
			   i.first_seen_at, i.last_seen_at, i.is_pitfall, i.severity_score,
			   i.category_id, i.score, i.html_url, i.comments_count,
			   i.is_duplicate, i.duplicate_of, i.metadata,
			   r.owner, r.name, r.full_name, r.description, r.url as repo_url,
			   r.stars, r.forks, r.language, r.created_at as repo_created_at,
			   c.name as category_name, c.description as category_description
		FROM issues i
		LEFT JOIN repositories r ON i.repository_id = r.id
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE 1=1
	`
	
	var params []interface{}
	paramCount := 0
	
	// 时间范围过滤
	if filter.DateFrom != nil {
		paramCount++
		query += fmt.Sprintf(" AND i.created_at >= $%d", paramCount)
		params = append(params, *filter.DateFrom)
	}
	
	if filter.DateTo != nil {
		paramCount++
		query += fmt.Sprintf(" AND i.created_at <= $%d", paramCount)
		params = append(params, *filter.DateTo)
	}
	
	// 分类过滤
	if len(filter.Categories) > 0 {
		paramCount++
		query += fmt.Sprintf(" AND i.category_id IN (SELECT id FROM categories WHERE name = ANY($%d))", paramCount)
		params = append(params, filter.Categories)
	}
	
	// 仓库过滤
	if len(filter.Repositories) > 0 {
		repoConditions := make([]string, len(filter.Repositories))
		for i, repo := range filter.Repositories {
			paramCount++
			parts := strings.Split(repo, "/")
			if len(parts) == 2 {
				repoConditions[i] = fmt.Sprintf("(r.owner = $%d AND r.name = $%d)", paramCount, paramCount+1)
				params = append(params, parts[0], parts[1])
				paramCount++
			}
		}
		if len(repoConditions) > 0 {
			query += fmt.Sprintf(" AND (%s)", strings.Join(repoConditions, " OR "))
		}
	}
	
	// 状态过滤
	if len(filter.States) > 0 {
		paramCount++
		query += fmt.Sprintf(" AND i.state = ANY($%d)", paramCount)
		params = append(params, filter.States)
	}
	
	// 分数范围过滤
	if filter.MinScore != nil {
		paramCount++
		query += fmt.Sprintf(" AND i.score >= $%d", paramCount)
		params = append(params, *filter.MinScore)
	}
	
	if filter.MaxScore != nil {
		paramCount++
		query += fmt.Sprintf(" AND i.score <= $%d", paramCount)
		params = append(params, *filter.MaxScore)
	}
	
	// 特殊标记过滤
	if filter.IsPitfall != nil {
		paramCount++
		query += fmt.Sprintf(" AND i.is_pitfall = $%d", paramCount)
		params = append(params, *filter.IsPitfall)
	}
	
	if filter.IsDuplicate != nil {
		paramCount++
		query += fmt.Sprintf(" AND i.is_duplicate = $%d", paramCount)
		params = append(params, *filter.IsDuplicate)
	}
	
	query += " ORDER BY i.created_at DESC"
	
	return query, params
}

// createOutputWriter 创建输出写入器
func (e *Exporter) createOutputWriter(format ExportFormat, outputPath string) (OutputWriter, error) {
	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create output directory")
	}
	
	switch format {
	case FormatJSON:
		return e.createJSONWriter(outputPath)
	case FormatCSV:
		return e.createCSVWriter(outputPath)
	case FormatMarkdown:
		return e.createMarkdownWriter(outputPath)
	default:
		return nil, errors.Errorf("unsupported export format: %s", format)
	}
}

// writeIssues 写入Issues数据
func (e *Exporter) writeIssues(writer OutputWriter, rows *sql.Rows) (int, error) {
	if err := writer.WriteHeader(); err != nil {
		return 0, errors.Wrap(err, "failed to write header")
	}
	
	count := 0
	for rows.Next() {
		issue, err := e.scanIssueRow(rows)
		if err != nil {
			e.logger.Printf("failed to scan issue row: %v", err)
			continue
		}
		
		if err := writer.WriteRecord(issue); err != nil {
			return count, errors.Wrap(err, "failed to write issue record")
		}
		count++
	}
	
	if err := writer.WriteFooter(); err != nil {
		return count, errors.Wrap(err, "failed to write footer")
	}
	
	return count, nil
}

// scanIssueRow 扫描Issue行数据
func (e *Exporter) scanIssueRow(rows *sql.Rows) (*Issue, error) {
	var issue Issue
	var labels, assignees JSONSlice
	var reactions ReactionCount
	var metadata JSONMap
	var repo Repository
	var categoryName, categoryDescription string
	
	err := rows.Scan(
		&issue.ID, &issue.Number, &issue.Title, &issue.Body, &issue.URL, &issue.State,
		&issue.AuthorLogin, &labels, &assignees, &issue.Milestone, &reactions,
		&issue.CreatedAt, &issue.UpdatedAt, &issue.ClosedAt, &issue.FirstSeenAt,
		&issue.LastSeenAt, &issue.IsPitfall, &issue.SeverityScore, &issue.CategoryID,
		&issue.Score, &issue.HTMLURL, &issue.CommentsCount, &issue.IsDuplicate,
		&issue.DuplicateOf, &metadata, &repo.Owner, &repo.Name, &repo.FullName,
		&repo.Description, &repo.URL, &repo.Stars, &repo.Forks, &repo.Language,
		&repo.CreatedAt, &categoryName, &categoryDescription,
	)
	
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan issue row")
	}
	
	issue.Labels = labels
	issue.Assignees = assignees
	issue.Reactions = reactions
	issue.Metadata = metadata
	
	// 添加关联数据
	issue.Repository = &repo
	if categoryName != "" {
		issue.Category = &Category{
			ID:          *issue.CategoryID,
			Name:        categoryName,
			Description: categoryDescription,
		}
	}
	
	return &issue, nil
}

// getTotalCount 获取总数
func (e *Exporter) getTotalCount(filter ExportFilter) int {
	query, params := e.buildExportQuery(filter)
	query = "SELECT COUNT(*) FROM (" + query + ") as subquery"
	
	var total int
	err := e.db.QueryRow(query, params...).Scan(&total)
	if err != nil {
		e.logger.Printf("failed to get total count: %v", err)
		return 0
	}
	
	return total
}

// JSON输出写入器实现
func (e *Exporter) createJSONWriter(outputPath string) (OutputWriter, error) {
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create JSON file")
	}
	
	return &JSONWriter{
		file:    file,
		encoder: json.NewEncoder(file),
		first:   true,
	}, nil
}

func (w *JSONWriter) WriteHeader() error {
	if _, err := w.file.WriteString("{\n  \"issues\": ["); err != nil {
		return err
	}
	return nil
}

func (w *JSONWriter) WriteRecord(issue *Issue) error {
	if !w.first {
		if _, err := w.file.WriteString(","); err != nil {
			return err
		}
	}
	w.first = false
	
	// 格式化JSON输出
	encoder := json.NewEncoder(w.file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(issue); err != nil {
		return err
	}
	return nil
}

func (w *JSONWriter) WriteFooter() error {
	_, err := w.file.WriteString("\n  ]\n}")
	return err
}

func (w *JSONWriter) Close() error {
	return w.file.Close()
}

// CSV输出写入器实现
func (e *Exporter) createCSVWriter(outputPath string) (OutputWriter, error) {
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CSV file")
	}
	
	writer := csv.NewWriter(file)
	
	// 定义CSV表头
	headers := []string{
		"ID", "Number", "Title", "Body", "State", "Author", "Labels",
		"Assignees", "Milestone", "Score", "Severity", "CreatedAt", "UpdatedAt",
		"Repository", "Category", "IsPitfall", "IsDuplicate", "URL",
	}
	
	if err := writer.Write(headers); err != nil {
		return nil, errors.Wrap(err, "failed to write CSV headers")
	}
	
	return &CSVWriter{
		writer:  writer,
		file:    file,
		headers: headers,
	}, nil
}

func (w *CSVWriter) WriteRecord(issue *Issue) error {
	record := []string{
		fmt.Sprintf("%d", issue.ID),
		fmt.Sprintf("%d", issue.Number),
		issue.Title,
		issue.Body,
		issue.State,
		issue.AuthorLogin,
		strings.Join(issue.Labels, ";"),
		strings.Join(issue.Assignees, ";"),
		issue.Milestone,
		fmt.Sprintf("%.2f", issue.Score),
		fmt.Sprintf("%.2f", issue.SeverityScore),
		issue.CreatedAt.Time.Format(time.RFC3339),
		issue.UpdatedAt.Time.Format(time.RFC3339),
		issue.Repository.FullName,
		getCategoryName(issue),
		fmt.Sprintf("%t", issue.IsPitfall),
		fmt.Sprintf("%t", issue.IsDuplicate),
		issue.URL,
	}
	
	return w.writer.Write(record)
}

func (w *CSVWriter) WriteFooter() error {
	w.writer.Flush()
	return w.writer.Error()
}

func (w *CSVWriter) Close() error {
	return w.file.Close()
}

// Markdown输出写入器实现
func (e *Exporter) createMarkdownWriter(outputPath string) (OutputWriter, error) {
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Markdown file")
	}
	
	// Markdown模板
	tmplStr := `
# GitHub Issues Export Report

Generated at: {{ .GeneratedAt }}

## Summary

- **Total Issues**: {{ .TotalCount }}
- **Export Date**: {{ .ExportDate }}
- **Format**: Markdown

## Issues

{{ range .Issues }}
### Issue #{{ .Number }}: {{ .Title }}

- **Repository**: [{{ .Repository.FullName }}]({{ .Repository.URL }})
- **State**: {{ .State }}
- **Author**: {{ .AuthorLogin }}
- **Score**: {{ .Score }}
- **Severity**: {{ .SeverityScore }}
- **Created**: {{ .CreatedAt.Time.Format "2006-01-02 15:04:05" }}
- **Updated**: {{ .UpdatedAt.Time.Format "2006-01-02 15:04:05" }}
- **Labels**: {{ range .Labels }}` + "`{{ . }}`" + `{{ end }}
- **Category**: {{ .Category }}
- **Is Pitfall**: {{ .IsPitfall }}
- **Is Duplicate**: {{ .IsDuplicate }}

#### Description

{{ .Body }}

---

{{ end }}
`
	
	tmpl, err := template.New("export").Parse(tmplStr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse Markdown template")
	}
	
	return &MarkdownWriter{
		file:     file,
		tmpl:     tmpl,
	}, nil
}

func (w *MarkdownWriter) WriteHeader() error {
	// Markdown不需要特殊的header
	return nil
}

func (w *MarkdownWriter) WriteRecord(issue *Issue) error {
	// Markdown格式一次写入所有数据
	return nil
}

func (w *MarkdownWriter) WriteFooter() error {
	// Markdown格式在所有数据写入后添加footer
	return nil
}

func (w *MarkdownWriter) Close() error {
	return w.file.Close()
}

// Helper functions
func getCategoryName(issue *Issue) string {
	if issue.Category != nil {
		return issue.Category.Name
	}
	if issue.CategoryID != nil {
		return fmt.Sprintf("Category ID: %d", *issue.CategoryID)
	}
	return ""
}

// ExportToBuffer 导出到内存缓冲区
func (e *Exporter) ExportToBuffer(filter ExportFilter, format ExportFormat) (*ExportResult, *bytes.Buffer, error) {
	var buffer bytes.Buffer
	
	// 创建临时文件路径
	tempDir := os.TempDir()
	outputPath := filepath.Join(tempDir, fmt.Sprintf("export_%d.%s", time.Now().Unix(), format))
	
	result, err := e.ExportIssues(filter, format, outputPath)
	if err != nil {
		return nil, nil, err
	}
	
	// 读取文件内容到缓冲区
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, nil, err
	}
	
	buffer.Write(data)
	
	// 清理临时文件
	os.Remove(outputPath)
	
	return result, &buffer, nil
}