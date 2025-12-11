package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// 数据库迁移配置
type MigrationConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	
	MigrationsDir string
	BackupDir     string
}

// Migration 数据库迁移项
type Migration struct {
	ID          string
	Name        string
	Description string
	UpSQL       string
	DownSQL     string
	AppliedAt   time.Time
}

// MigrationManager 迁移管理器
type MigrationManager struct {
	db     *sql.DB
	config *MigrationConfig
	logger *log.Logger
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(config *MigrationConfig, logger *log.Logger) (*MigrationManager, error) {
	if logger == nil {
		logger = log.New(os.Stdout, "[Migration] ", log.LstdFlags|log.Lmsgprefix)
	}
	
	// 构建数据库连接字符串
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
	
	// 打开数据库连接
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}
	
	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}
	
	return &MigrationManager{
		db:     db,
		config: config,
		logger: logger,
	}, nil
}

// InitSchema 初始化架构
func (m *MigrationManager) InitSchema() error {
	m.logger.Println("初始化数据库架构...")
	
	// 创建迁移表
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			up_sql TEXT NOT NULL,
			down_sql TEXT,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	
	if err != nil {
		return fmt.Errorf("创建迁移表失败: %w", err)
	}
	
	m.logger.Println("数据库架构初始化完成")
	return nil
}

// CreateMigration 创建新的迁移
func (m *MigrationManager) CreateMigration(name, description string) error {
	if name == "" {
		return fmt.Errorf("迁移名称不能为空")
	}
	
	// 生成版本号
	timestamp := time.Now().Format("20060102150405")
	version := fmt.Sprintf("%s_%s", timestamp, strings.ReplaceAll(name, " ", "_"))
	
	// 创建迁移文件
	migration := Migration{
		ID:          version,
		Name:        name,
		Description: description,
		UpSQL:       "-- TODO: 添加迁移SQL\n",
		DownSQL:     "-- TODO: 添加回滚SQL\n",
	}
	
	// 写入迁移文件
	migrationsDir := m.config.MigrationsDir
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}
	
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("创建迁移目录失败: %w", err)
	}
	
	upFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_up.sql", version))
	downFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_down.sql", version))
	
	if err := os.WriteFile(upFile, []byte(migration.UpSQL), 0644); err != nil {
		return fmt.Errorf("创建向上迁移文件失败: %w", err)
	}
	
	if err := os.WriteFile(downFile, []byte(migration.DownSQL), 0644); err != nil {
		return fmt.Errorf("创建向下迁移文件失败: %w", err)
	}
	
	m.logger.Printf("创建迁移文件: %s, %s", upFile, downFile)
	return nil
}

// MigrateUp 执行向上迁移
func (m *MigrationManager) MigrateUp() error {
	m.logger.Println("执行数据库迁移...")
	
	// 初始化架构
	if err := m.InitSchema(); err != nil {
		return err
	}
	
	// 获取已应用的迁移
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("获取已应用迁移失败: %w", err)
	}
	
	// 扫描迁移文件
	migrations, err := m.scanMigrations()
	if err != nil {
		return fmt.Errorf("扫描迁移文件失败: %w", err)
	}
	
	// 执行未应用的迁移
	appliedCount := 0
	for _, migration := range migrations {
		if !contains(applied, migration.ID) {
			m.logger.Printf("应用迁移: %s - %s", migration.ID, migration.Name)
			
			// 开始事务
			tx, err := m.db.Begin()
			if err != nil {
				return fmt.Errorf("开始事务失败: %w", err)
			}
			
			// 执行迁移SQL
			if _, err := tx.Exec(migration.UpSQL); err != nil {
				tx.Rollback()
				return fmt.Errorf("执行迁移 %s 失败: %w", migration.ID, err)
			}
			
			// 记录迁移
			if _, err := tx.Exec(`
				INSERT INTO schema_migrations (version, name, description, up_sql, down_sql)
				VALUES ($1, $2, $3, $4, $5)
			`, migration.ID, migration.Name, migration.Description, migration.UpSQL, migration.DownSQL); err != nil {
				tx.Rollback()
				return fmt.Errorf("记录迁移失败: %w", err)
			}
			
			// 提交事务
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("提交事务失败: %w", err)
			}
			
			appliedCount++
			m.logger.Printf("迁移 %s 应用成功", migration.ID)
		}
	}
	
	m.logger.Printf("迁移完成，共应用 %d 个迁移", appliedCount)
	return nil
}

// MigrateDown 执行向下迁移
func (m *MigrationManager) MigrateDown(steps int) error {
	m.logger.Printf("执行向下迁移 %d 步...", steps)
	
	// 获取已应用的迁移
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("获取已应用迁移失败: %w", err)
	}
	
	// 按时间倒序排列
	for i := len(applied) - 1; i >= 0 && steps > 0; i-- {
		migration := applied[i]
		
		if migration.DownSQL == "" {
			m.logger.Printf("跳过迁移 %s (无向下SQL)", migration.ID)
			continue
		}
		
		m.logger.Printf("回滚迁移: %s - %s", migration.ID, migration.Name)
		
		// 开始事务
		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("开始事务失败: %w", err)
		}
		
		// 执行回滚SQL
		if _, err := tx.Exec(migration.DownSQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("执行回滚 %s 失败: %w", migration.ID, err)
		}
		
		// 删除迁移记录
		if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", migration.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("删除迁移记录失败: %w", err)
		}
		
		// 提交事务
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败: %w", err)
		}
		
		steps--
		m.logger.Printf("迁移 %s 回滚成功", migration.ID)
	}
	
	m.logger.Println("向下迁移完成")
	return nil
}

// Status 显示迁移状态
func (m *MigrationManager) Status() error {
	m.logger.Println("迁移状态:")
	
	// 初始化架构
	if err := m.InitSchema(); err != nil {
		return err
	}
	
	// 获取已应用的迁移
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("获取已应用迁移失败: %w", err)
	}
	
	// 扫描迁移文件
	migrations, err := m.scanMigrations()
	if err != nil {
		return fmt.Errorf("扫描迁移文件失败: %w", err)
	}
	
	fmt.Println("\n已应用的迁移:")
	for _, migration := range applied {
		fmt.Printf("  ✓ %s - %s (%s)\n", migration.ID, migration.Name, migration.AppliedAt.Format("2006-01-02 15:04:05"))
	}
	
	fmt.Println("\n待应用的迁移:")
	for _, migration := range migrations {
		if !contains(applied, migration.ID) {
			fmt.Printf("  ○ %s - %s\n", migration.ID, migration.Name)
		}
	}
	
	return nil
}

// getAppliedMigrations 获取已应用的迁移
func (m *MigrationManager) getAppliedMigrations() ([]Migration, error) {
	rows, err := m.db.Query(`
		SELECT version, name, description, up_sql, down_sql, applied_at
		FROM schema_migrations
		ORDER BY applied_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var migrations []Migration
	for rows.Next() {
		var migration Migration
		err := rows.Scan(&migration.ID, &migration.Name, &migration.Description,
			&migration.UpSQL, &migration.DownSQL, &migration.AppliedAt)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}
	
	return migrations, nil
}

// scanMigrations 扫描迁移文件
func (m *MigrationManager) scanMigrations() ([]Migration, error) {
	migrationsDir := m.config.MigrationsDir
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}
	
	// 创建迁移目录
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建迁移目录失败: %w", err)
	}
	
	// 读取迁移文件
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("读取迁移目录失败: %w", err)
	}
	
	var migrations []Migration
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), "_up.sql") {
			continue
		}
		
		// 解析版本号
		version := strings.TrimSuffix(file.Name(), "_up.sql")
		
		// 读取向上迁移SQL
		upPath := filepath.Join(migrationsDir, file.Name())
		upSQL, err := os.ReadFile(upPath)
		if err != nil {
			m.logger.Printf("读取迁移文件失败: %s", upPath)
			continue
		}
		
		// 读取向下迁移SQL
		downPath := filepath.Join(migrationsDir, fmt.Sprintf("%s_down.sql", version))
		downSQL, err := os.ReadFile(downPath)
		if err != nil {
			downSQL = nil // 向下迁移是可选的
		}
		
		migration := Migration{
			ID:      version,
			UpSQL:   string(upSQL),
			DownSQL: string(downSQL),
		}
		
		migrations = append(migrations, migration)
	}
	
	return migrations, nil
}

// Close 关闭数据库连接
func (m *MigrationManager) Close() error {
	return m.db.Close()
}

// contains 检查切片是否包含指定元素
func contains(migrations []Migration, version string) bool {
	for _, migration := range migrations {
		if migration.ID == version {
			return true
		}
	}
	return false
}

// 解析命令行参数
func parseArgs() (*MigrationConfig, string, []string) {
	config := &MigrationConfig{
		Host:          getEnv("DB_HOST", "localhost"),
		Port:          getEnvAsInt("DB_PORT", 5432),
		User:          getEnv("DB_USER", "postgres"),
		Password:      getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", "gh_pitfall_scraper"),
		SSLMode:       getEnv("DB_SSLMODE", "disable"),
		MigrationsDir: getEnv("MIGRATIONS_DIR", "./migrations"),
		BackupDir:     getEnv("BACKUP_DIR", "./backups"),
	}
	
	command := ""
	args := make([]string, 0)
	
	flag.StringVar(&config.Host, "host", config.Host, "数据库主机")
	flag.IntVar(&config.Port, "port", config.Port, "数据库端口")
	flag.StringVar(&config.User, "user", config.User, "数据库用户")
	flag.StringVar(&config.Password, "password", config.Password, "数据库密码")
	flag.StringVar(&config.DBName, "dbname", config.DBName, "数据库名称")
	flag.StringVar(&config.SSLMode, "sslmode", config.SSLMode, "SSL模式")
	flag.StringVar(&config.MigrationsDir, "migrations", config.MigrationsDir, "迁移文件目录")
	flag.StringVar(&config.BackupDir, "backup", config.BackupDir, "备份目录")
	
	flag.Parse()
	
	// 获取命令
	remainingArgs := flag.Args()
	if len(remainingArgs) > 0 {
		command = remainingArgs[0]
		args = remainingArgs[1:]
	}
	
	return config, command, args
}

// 获取环境变量或默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 获取环境变量作为整数
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 主函数
func main() {
	config, command, args := parseArgs()
	
	if command == "" {
		printUsage()
		return
	}
	
	// 创建迁移管理器
	mgr, err := NewMigrationManager(config, nil)
	if err != nil {
		log.Fatalf("创建迁移管理器失败: %v", err)
	}
	defer mgr.Close()
	
	// 执行命令
	switch command {
	case "init":
		err = mgr.InitSchema()
	case "create":
		if len(args) < 1 {
			log.Fatal("创建迁移需要指定名称")
		}
		name := args[0]
		description := ""
		if len(args) > 1 {
			description = args[1]
		}
		err = mgr.CreateMigration(name, description)
	case "migrate":
		err = mgr.MigrateUp()
	case "rollback":
		steps := 1
		if len(args) > 0 {
			if s, err := strconv.Atoi(args[0]); err == nil {
				steps = s
			}
		}
		err = mgr.MigrateDown(steps)
	case "status":
		err = mgr.Status()
	default:
		log.Fatalf("未知命令: %s", command)
	}
	
	if err != nil {
		log.Fatalf("命令执行失败: %v", err)
	}
}

// 显示用法
func printUsage() {
	fmt.Println("数据库迁移工具")
	fmt.Println()
	fmt.Println("用法: migration [选项] <命令> [参数]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  init                 初始化数据库架构")
	fmt.Println("  create <name> [desc] 创建新的迁移")
	fmt.Println("  migrate              执行向上迁移")
	fmt.Println("  rollback [steps]     执行向下迁移")
	fmt.Println("  status               显示迁移状态")
	fmt.Println()
	fmt.Println("选项:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("环境变量:")
	fmt.Println("  DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME")
	fmt.Println("  DB_SSLMODE, MIGRATIONS_DIR, BACKUP_DIR")
}