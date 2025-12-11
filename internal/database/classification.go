package database

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ClassificationService provides intelligent classification functionality
type ClassificationService struct {
	db          *sql.DB
	logger      *log.Logger
	config      ClassificationConfig
	classifier  *ClassificationEngine
}

// ClassificationConfig represents configuration for classification
type ClassificationConfig struct {
	AutoClassificationThreshold float64 `json:"auto_classification_threshold"` // Minimum confidence for auto-classification
	BatchSize                   int     `json:"batch_size"`                   // Number of issues to process in one batch
	EnableParallel              bool    `json:"enable_parallel"`              // Enable parallel processing
	MaxWorkers                  int     `json:"max_workers"`                  // Maximum number of parallel workers
	MinConfidence               float64 `json:"min_confidence"`               // Minimum confidence score
	EnableTechStackDetection    bool    `json:"enable_tech_stack_detection"`  // Enable technology stack detection
}

// DefaultClassificationConfig returns default configuration
func DefaultClassificationConfig() ClassificationConfig {
	return ClassificationConfig{
		AutoClassificationThreshold: 0.8,
		BatchSize:                   100,
		EnableParallel:              true,
		MaxWorkers:                  4,
		MinConfidence:               0.6,
		EnableTechStackDetection:    true,
	}
}

// ClassificationEngine performs the actual classification logic
type ClassificationEngine struct {
	categoryRules   map[string][]ClassificationRule
	priorityRules   map[string][]ClassificationRule
	techStackRules  map[string][]ClassificationRule
	keywordWeights  map[string]map[string]float64 // category -> keyword -> weight
	patternRules    []PatternRule
}

// ClassificationRule represents a rule for classification
type ClassificationRule struct {
	Name        string    `json:"name"`
	Pattern     string    `json:"pattern"`
	Weight      float64   `json:"weight"`
	Category    string    `json:"category"`
	Priority    string    `json:"priority"`
	TechStack   []string  `json:"tech_stack"`
	Keywords    []string  `json:"keywords"`
	Regex       *regexp.Regexp
	CreatedAt   time.Time `json:"created_at"`
}

// PatternRule represents a regex pattern for classification
type PatternRule struct {
	Name        string            `json:"name"`
	Pattern     string            `json:"pattern"`
	Category    string            `json:"category"`
	Priority    string            `json:"priority"`
	TechStack   []string          `json:"tech_stack"`
	Confidence  float64           `json:"confidence"`
	Regex       *regexp.Regexp
	Groups      map[string]string `json:"groups"` // Named groups for extraction
}

// ClassificationResult represents the result of classification
type ClassificationResult struct {
	Issue         *Issue       `json:"issue"`
	PredictedCategory string   `json:"predicted_category"`
	PredictedPriority string   `json:"predicted_priority"`
	PredictedTechStack []string `json:"predicted_tech_stack"`
	Confidence     float64     `json:"confidence"`
	Reasons       []string     `json:"reasons"`
	RuleMatches   []RuleMatch  `json:"rule_matches"`
	ProcessedAt   time.Time    `json:"processed_at"`
}

// RuleMatch represents a matched classification rule
type RuleMatch struct {
	RuleName      string  `json:"rule_name"`
	Pattern       string  `json:"pattern"`
	Weight        float64 `json:"weight"`
	MatchedText   string  `json:"matched_text"`
	Category      string  `json:"category"`
	Priority      string  `json:"priority"`
	TechStack     []string `json:"tech_stack"`
	Confidence    float64 `json:"confidence"`
}

// NewClassificationEngine creates a new classification engine
func NewClassificationEngine() *ClassificationEngine {
	engine := &ClassificationEngine{
		categoryRules:  make(map[string][]ClassificationRule),
		priorityRules:  make(map[string][]ClassificationRule),
		techStackRules: make(map[string][]ClassificationRule),
		keywordWeights: make(map[string]map[string]float64),
		patternRules:   make([]PatternRule, 0),
	}
	
	// Initialize default rules
	engine.initializeDefaultRules()
	
	return engine
}

// initializeDefaultRules initializes default classification rules
func (ce *ClassificationEngine) initializeDefaultRules() {
	// Category rules
	ce.categoryRules["bug"] = []ClassificationRule{
		{
			Name:     "bug_keyword",
			Pattern:  "(?i)(bug|error|issue|wrong|crash|fail|broken)",
			Weight:   0.8,
			Category: "bug",
		},
		{
			Name:     "bug_specific",
			Pattern:  "(?i)(segfault|nullpointerexception|stack overflow|memory leak)",
			Weight:   0.9,
			Category: "bug",
		},
	}
	
	ce.categoryRules["performance"] = []ClassificationRule{
		{
			Name:     "performance_keyword",
			Pattern:  "(?i)(slow|performance|latency|speed|optimize|timeout)",
			Weight:   0.8,
			Category: "performance",
		},
		{
			Name:     "performance_specific",
			Pattern:  "(?i)(cpu usage|memory usage|response time|deadlock|race condition)",
			Weight:   0.9,
			Category: "performance",
		},
	}
	
	ce.categoryRules["security"] = []ClassificationRule{
		{
			Name:     "security_keyword",
			Pattern:  "(?i)(security|vulnerability|xss|csrf|injection|exploit)",
			Weight:   0.9,
			Category: "security",
		},
		{
			Name:     "security_specific",
			Pattern:  "(?i)(sql injection|xss|csrf|buffer overflow|integer overflow)",
			Weight:   1.0,
			Category: "security",
		},
	}
	
	ce.categoryRules["feature"] = []ClassificationRule{
		{
			Name:     "feature_keyword",
			Pattern:  "(?i)(feature|enhancement|improvement|add|implement|new)",
			Weight:   0.8,
			Category: "feature",
		},
	}
	
	ce.categoryRules["documentation"] = []ClassificationRule{
		{
			Name:     "documentation_keyword",
			Pattern:  "(?i)(docs|documentation|readme|guide|tutorial)",
			Weight:   0.9,
			Category: "documentation",
		},
	}
	
	ce.categoryRules["api"] = []ClassificationRule{
		{
			Name:     "api_keyword",
			Pattern:  "(?i)(api|endpoint|rest|graphql|webhook|integration)",
			Weight:   0.8,
			Category: "api",
		},
	}
	
	ce.categoryRules["ui"] = []ClassificationRule{
		{
			Name:     "ui_keyword",
			Pattern:  "(?i)(ui|interface|design|layout|css|html|responsive)",
			Weight:   0.8,
			Category: "ui",
		},
	}
	
	ce.categoryRules["database"] = []ClassificationRule{
		{
			Name:     "database_keyword",
			Pattern:  "(?i)(database|sql|mysql|postgresql|mongodb|redis|query)",
			Weight:   0.9,
			Category: "database",
		},
	}
	
	ce.categoryRules["testing"] = []ClassificationRule{
		{
			Name:     "testing_keyword",
			Pattern:  "(?i)(test|testing|unit test|integration test|jest|junit)",
			Weight:   0.9,
			Category: "testing",
		},
	}
	
	ce.categoryRules["deployment"] = []ClassificationRule{
		{
			Name:     "deployment_keyword",
			Pattern:  "(?i)(deploy|docker|kubernetes|ci|cd|pipeline)",
			Weight:   0.8,
			Category: "deployment",
		},
	}
	
	// Priority rules
	ce.priorityRules["critical"] = []ClassificationRule{
		{
			Name:     "critical_keyword",
			Pattern:  "(?i)(critical|urgent|emergency|security|production down)",
			Weight:   1.0,
			Priority: "critical",
		},
	}
	
	ce.priorityRules["high"] = []ClassificationRule{
		{
			Name:     "high_keyword",
			Pattern:  "(?i)(important|high|severe|major|blocking)",
			Weight:   0.8,
			Priority: "high",
		},
	}
	
	ce.priorityRules["medium"] = []ClassificationRule{
		{
			Name:     "medium_keyword",
			Pattern:  "(?i)(normal|medium|moderate|standard)",
			Weight:   0.6,
			Priority: "medium",
		},
	}
	
	ce.priorityRules["low"] = []ClassificationRule{
		{
			Name:     "low_keyword",
			Pattern:  "(?i)(minor|low|enhancement|feature request|nice to have)",
			Weight:   0.4,
			Priority: "low",
		},
	}
	
	// Technology stack rules
	ce.techStackRules["javascript"] = []ClassificationRule{
		{
			Name:      "javascript_keyword",
			Pattern:   "(?i)(javascript|js|node|npm|yarn|webpack|react|vue|angular)",
			Weight:    0.9,
			TechStack: []string{"javascript"},
		},
	}
	
	ce.techStackRules["python"] = []ClassificationRule{
		{
			Name:      "python_keyword",
			Pattern:   "(?i)(python|py|pip|django|flask|fastapi|pandas|numpy)",
			Weight:    0.9,
			TechStack: []string{"python"},
		},
	}
	
	ce.techStackRules["go"] = []ClassificationRule{
		{
			Name:      "go_keyword",
			Pattern:   "(?i)(golang|go|goroutine|gorilla|gin|echo|fiber)",
			Weight:    0.9,
			TechStack: []string{"go"},
		},
	}
	
	ce.techStackRules["java"] = []ClassificationRule{
		{
			Name:      "java_keyword",
			Pattern:   "(?i)(java|jvm|spring|hibernate|maven|gradle|junit)",
			Weight:    0.9,
			TechStack: []string{"java"},
		},
	}
	
	ce.techStackRules["csharp"] = []ClassificationRule{
		{
			Name:      "csharp_keyword",
			Pattern:   "(?i)(c#|.net|dotnet|microsoft|asp.net|entity framework)",
			Weight:    0.9,
			TechStack: []string{"csharp"},
		},
	}
	
	ce.techStackRules["php"] = []ClassificationRule{
		{
			Name:      "php_keyword",
			Pattern:   "(?i)(php|laravel|symfony|composer|drupal|wordpress)",
			Weight:    0.9,
			TechStack: []string{"php"},
		},
	}
	
	ce.techStackRules["rust"] = []ClassificationRule{
		{
			Name:      "rust_keyword",
			Pattern:   "(?i)(rust|cargo|rustc|tokio|serde|actix)",
			Weight:    0.9,
			TechStack: []string{"rust"},
		},
	}
	
	ce.techStackRules["database"] = []ClassificationRule{
		{
			Name:      "database_keyword",
			Pattern:   "(?i)(mysql|postgresql|sqlite|mongodb|redis|cassandra|oracle)",
			Weight:    1.0,
			TechStack: []string{"database"},
		},
	}
	
	// Compile regex patterns
	ce.compileRegexPatterns()
	
	// Initialize keyword weights
	ce.initializeKeywordWeights()
}

// compileRegexPatterns compiles regex patterns for all rules
func (ce *ClassificationEngine) compileRegexPatterns() {
	// Compile category rules
	for category, rules := range ce.categoryRules {
		for i := range rules {
			regex, err := regexp.Compile(rules[i].Pattern)
			if err != nil {
				log.Printf("Failed to compile regex for category rule %s: %v", rules[i].Name, err)
				continue
			}
			ce.categoryRules[category][i].Regex = regex
		}
	}
	
	// Compile priority rules
	for priority, rules := range ce.priorityRules {
		for i := range rules {
			regex, err := regexp.Compile(rules[i].Pattern)
			if err != nil {
				log.Printf("Failed to compile regex for priority rule %s: %v", rules[i].Name, err)
				continue
			}
			ce.priorityRules[priority][i].Regex = regex
		}
	}
	
	// Compile tech stack rules
	for tech, rules := range ce.techStackRules {
		for i := range rules {
			regex, err := regexp.Compile(rules[i].Pattern)
			if err != nil {
				log.Printf("Failed to compile regex for tech stack rule %s: %v", rules[i].Name, err)
				continue
			}
			ce.techStackRules[tech][i].Regex = regex
		}
	}
}

// initializeKeywordWeights initializes keyword weights for different categories
func (ce *ClassificationEngine) initializeKeywordWeights() {
	ce.keywordWeights = map[string]map[string]float64{
		"bug": {
			"bug":        1.0,
			"error":      0.9,
			"crash":      0.8,
			"fail":       0.7,
			"broken":     0.8,
			"wrong":      0.6,
			"issue":      0.5,
		},
		"performance": {
			"performance": 1.0,
			"slow":        0.9,
			"speed":       0.8,
			"optimize":    0.9,
			"timeout":     0.7,
			"latency":     0.8,
		},
		"security": {
			"security":     1.0,
			"vulnerability": 1.0,
			"xss":          0.9,
			"injection":    0.9,
			"exploit":      0.8,
		},
		"feature": {
			"feature":      1.0,
			"enhancement":  0.9,
			"improvement":  0.8,
			"add":          0.7,
			"implement":    0.8,
			"new":          0.6,
		},
		"api": {
			"api":       1.0,
			"endpoint":  0.9,
			"rest":      0.8,
			"graphql":   0.9,
			"webhook":   0.8,
			"integration": 0.7,
		},
	}
}

// ClassifyIssue classifies a single issue
func (ce *ClassificationEngine) ClassifyIssue(issue *Issue) (*ClassificationResult, error) {
	if issue == nil {
		return nil, fmt.Errorf("issue cannot be nil")
	}
	
	result := &ClassificationResult{
		Issue:         issue,
		ProcessedAt:   time.Now(),
		RuleMatches:   make([]RuleMatch, 0),
		Reasons:       make([]string, 0),
	}
	
	// Combine title and body for analysis
	fullText := strings.Join([]string{issue.Title, issue.Body}, " ")
	
	// Classify category
	category, categoryConfidence, categoryMatches := ce.classifyCategory(fullText, issue.Keywords)
	result.PredictedCategory = category
	
	// Classify priority
	priority, priorityConfidence, priorityMatches := ce.classifyPriority(fullText, issue.Keywords, issue.Score)
	result.PredictedPriority = priority
	
	// Classify tech stack
	techStack, techStackConfidence, techStackMatches := ce.classifyTechStack(fullText, issue.Keywords)
	result.PredictedTechStack = techStack
	
	// Calculate overall confidence
	confidence := (categoryConfidence + priorityConfidence + techStackConfidence) / 3.0
	result.Confidence = confidence
	
	// Collect rule matches
	result.RuleMatches = append(categoryMatches, priorityMatches...)
	result.RuleMatches = append(result.RuleMatches, techStackMatches...)
	
	// Generate reasons
	result.Reasons = ce.generateReasons(fullText, category, priority, techStack, result.RuleMatches)
	
	return result, nil
}

// classifyCategory classifies the category of an issue
func (ce *ClassificationEngine) classifyCategory(text string, keywords StringArray) (string, float64, []RuleMatch) {
	var matches []RuleMatch
	scores := make(map[string]float64)
	
	// Score based on regex rules
	for category, rules := range ce.categoryRules {
		for _, rule := range rules {
			if rule.Regex != nil && rule.Regex.MatchString(text) {
				matches = append(matches, RuleMatch{
					RuleName:   rule.Name,
					Pattern:    rule.Pattern,
					Weight:     rule.Weight,
					Category:   category,
					Confidence: rule.Weight,
				})
				scores[category] += rule.Weight
			}
		}
	}
	
	// Score based on keywords
	for _, keyword := range keywords {
		for category, weights := range ce.keywordWeights {
			if weight, exists := weights[strings.ToLower(keyword)]; exists {
				scores[category] += weight * 0.5 // Lower weight for keywords
				matches = append(matches, RuleMatch{
					RuleName:   fmt.Sprintf("keyword_%s", keyword),
					Pattern:    keyword,
					Weight:     weight,
					Category:   category,
					Confidence: weight * 0.5,
				})
			}
		}
	}
	
	// Find the category with highest score
	bestCategory := "uncategorized"
	bestScore := 0.0
	
	for category, score := range scores {
		if score > bestScore {
			bestScore = score
			bestCategory = category
		}
	}
	
	// Normalize confidence
	confidence := math.Min(1.0, bestScore/2.0) // Assuming max score of 2.0
	
	return bestCategory, confidence, matches
}

// classifyPriority classifies the priority of an issue
func (ce *ClassificationEngine) classifyPriority(text string, keywords StringArray, score float64) (string, float64, []RuleMatch) {
	var matches []RuleMatch
	scores := make(map[string]float64)
	
	// Score based on regex rules
	for priority, rules := range ce.priorityRules {
		for _, rule := range rules {
			if rule.Regex != nil && rule.Regex.MatchString(text) {
				matches = append(matches, RuleMatch{
					RuleName:   rule.Name,
					Pattern:    rule.Pattern,
					Weight:     rule.Weight,
					Priority:   priority,
					Confidence: rule.Weight,
				})
				scores[priority] += rule.Weight
			}
		}
	}
	
	// Score based on issue score
	if score >= 25.0 {
		scores["critical"] += 0.8
	} else if score >= 20.0 {
		scores["high"] += 0.7
	} else if score >= 15.0 {
		scores["medium"] += 0.6
	} else {
		scores["low"] += 0.5
	}
	
	// Score based on keywords
	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)
		for priority, rules := range ce.priorityRules {
			for _, rule := range rules {
				if strings.Contains(keywordLower, rule.Name) {
					scores[priority] += 0.3
				}
			}
		}
	}
	
	// Find the priority with highest score
	bestPriority := "medium" // Default priority
	bestScore := 0.0
	
	for priority, score := range scores {
		if score > bestScore {
			bestScore = score
			bestPriority = priority
		}
	}
	
	// Normalize confidence
	confidence := math.Min(1.0, bestScore/2.0)
	
	return bestPriority, confidence, matches
}

// classifyTechStack classifies the technology stack of an issue
func (ce *ClassificationEngine) classifyTechStack(text string, keywords StringArray) ([]string, float64, []RuleMatch) {
	var matches []RuleMatch
	scores := make(map[string]float64)
	var detectedTechStacks []string
	
	// Score based on regex rules
	for tech, rules := range ce.techStackRules {
		for _, rule := range rules {
			if rule.Regex != nil && rule.Regex.MatchString(text) {
				matches = append(matches, RuleMatch{
					RuleName:    rule.Name,
					Pattern:     rule.Pattern,
					Weight:      rule.Weight,
					TechStack:   rule.TechStack,
					Confidence:  rule.Weight,
				})
				scores[tech] += rule.Weight
			}
		}
	}
	
	// Score based on keywords
	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)
		for tech, rules := range ce.techStackRules {
			for _, rule := range rules {
				for _, techName := range rule.TechStack {
					if strings.Contains(keywordLower, techName) {
						scores[tech] += 0.5
					}
				}
			}
		}
	}
	
	// Collect tech stacks with scores above threshold
	threshold := 0.3
	for tech, score := range scores {
		if score >= threshold {
			detectedTechStacks = append(detectedTechStacks, tech)
		}
	}
	
	// If no tech stack detected, try to infer from repository language or keywords
	if len(detectedTechStacks) == 0 {
		// This is a simplified inference - in a real implementation, 
		// you might want to use more sophisticated heuristics
		if strings.Contains(strings.ToLower(text), "python") {
			detectedTechStacks = append(detectedTechStacks, "python")
		} else if strings.Contains(strings.ToLower(text), "javascript") {
			detectedTechStacks = append(detectedTechStacks, "javascript")
		}
	}
	
	// Calculate confidence
	confidence := 0.0
	if len(detectedTechStacks) > 0 {
		avgScore := 0.0
		for _, tech := range detectedTechStacks {
			avgScore += scores[tech]
		}
		confidence = math.Min(1.0, avgScore/float64(len(detectedTechStacks)))
	}
	
	return detectedTechStacks, confidence, matches
}

// generateReasons generates human-readable reasons for the classification
func (ce *ClassificationEngine) generateReasons(text, category, priority string, techStack []string, matches []RuleMatch) []string {
	var reasons []string
	
	// Add category reason
	if category != "uncategorized" {
		reasons = append(reasons, fmt.Sprintf("Classified as '%s' based on content analysis", category))
	}
	
	// Add priority reason
	reasons = append(reasons, fmt.Sprintf("Priority set to '%s'", priority))
	
	// Add tech stack reasons
	if len(techStack) > 0 {
		reasons = append(reasons, fmt.Sprintf("Technology stack detected: %s", strings.Join(techStack, ", ")))
	}
	
	// Add top rule matches
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Confidence > matches[j].Confidence
	})
	
	topMatches := matches
	if len(matches) > 3 {
		topMatches = matches[:3]
	}
	
	for _, match := range topMatches {
		if match.Category != "" {
			reasons = append(reasons, fmt.Sprintf("Matched pattern '%s' for category '%s'", match.Pattern, match.Category))
		}
		if match.Priority != "" {
			reasons = append(reasons, fmt.Sprintf("Matched pattern '%s' for priority '%s'", match.Pattern, match.Priority))
		}
	}
	
	return reasons
}

// NewClassificationService creates a new classification service
func NewClassificationService(db *sql.DB, config ClassificationConfig) *ClassificationService {
	if config.BatchSize == 0 {
		config.BatchSize = DefaultClassificationConfig().BatchSize
	}
	if config.MaxWorkers == 0 {
		config.MaxWorkers = DefaultClassificationConfig().MaxWorkers
	}
	
	return &ClassificationService{
		db:         db,
		logger:     log.New(log.Writer(), "[Classification] ", log.LstdFlags),
		config:     config,
		classifier: NewClassificationEngine(),
	}
}

// ClassifyIssues classifies multiple issues
func (cs *ClassificationService) ClassifyIssues(issues []*Issue) (*ClassificationResult, error) {
	if len(issues) == 0 {
		return &ClassificationResult{
			TotalProcessed:    0,
			Classified:        0,
			AutoClassified:    0,
			ManuallyReviewed:  0,
			Confidence:        0.0,
			ByCategory:        make(map[string]int),
			ByPriority:        make(map[string]int),
			ByTechStack:       make(map[string]int),
		}, nil
	}
	
	cs.logger.Printf("Starting classification of %d issues", len(issues))
	
	startTime := time.Now()
	
	var results []*ClassificationResult
	var mutex sync.Mutex
	var wg sync.WaitGroup
	
	// Process issues in batches
	if cs.config.EnableParallel {
		batchSize := cs.config.BatchSize
		numBatches := (len(issues) + batchSize - 1) / batchSize
		
		for i := 0; i < numBatches; i++ {
			start := i * batchSize
			end := min(start+batchSize, len(issues))
			batch := issues[start:end]
			
			wg.Add(1)
			go func(b []*Issue) {
				defer wg.Done()
				batchResults := cs.processBatch(b)
				mutex.Lock()
				results = append(results, batchResults...)
				mutex.Unlock()
			}(batch)
		}
		
		wg.Wait()
	} else {
		results = cs.processBatch(issues)
	}
	
	// Update database with classification results
	stats, err := cs.updateClassificationResults(results)
	if err != nil {
		return nil, fmt.Errorf("failed to update classification results: %w", err)
	}
	
	processingTime := time.Since(startTime)
	cs.logger.Printf("Classification completed in %v", processingTime)
	
	return stats, nil
}

// processBatch processes a batch of issues for classification
func (cs *ClassificationService) processBatch(issues []*Issue) []*ClassificationResult {
	var results []*ClassificationResult
	
	for _, issue := range issues {
		result, err := cs.classifier.ClassifyIssue(issue)
		if err != nil {
			cs.logger.Printf("Failed to classify issue %d: %v", issue.ID, err)
			continue
		}
		
		// Auto-classify if confidence is high enough
		if result.Confidence >= cs.config.AutoClassificationThreshold {
			issue.Category = result.PredictedCategory
			issue.Priority = result.PredictedPriority
			issue.TechStack = result.PredictedTechStack
		}
		
		results = append(results, result)
	}
	
	return results
}

// updateClassificationResults updates the database with classification results
func (cs *ClassificationService) updateClassificationResults(results []*ClassificationResult) (*ClassificationResult, error) {
	tx, err := cs.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	stats := &ClassificationResult{
		TotalProcessed:    len(results),
		Classified:        0,
		AutoClassified:    0,
		ManuallyReviewed:  0,
		Confidence:        0.0,
		ByCategory:        make(map[string]int),
		ByPriority:        make(map[string]int),
		ByTechStack:       make(map[string]int),
	}
	
	stmt, err := tx.Prepare(`
		UPDATE issues SET category = ?, priority = ?, tech_stack = ?, updated_at_db = ?
		WHERE id = ?
	`)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	avgConfidence := 0.0
	
	for _, result := range results {
		if result.PredictedCategory != "" || result.PredictedPriority != "" {
			stats.Classified++
			avgConfidence += result.Confidence
		}
		
		// Check if this was auto-classified
		if result.Confidence >= cs.config.AutoClassificationThreshold {
			stats.AutoClassified++
			
			// Update database
			_, err := stmt.Exec(
				result.PredictedCategory,
				result.PredictedPriority,
				result.PredictedTechStack,
				time.Now(),
				result.Issue.ID,
			)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update issue classification: %w", err)
			}
		} else {
			stats.ManuallyReviewed++
		}
		
		// Update statistics
		if result.PredictedCategory != "" {
			stats.ByCategory[result.PredictedCategory]++
		}
		if result.PredictedPriority != "" {
			stats.ByPriority[result.PredictedPriority]++
		}
		for _, tech := range result.PredictedTechStack {
			stats.ByTechStack[tech]++
		}
	}
	
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// Calculate average confidence
	if len(results) > 0 {
		stats.Confidence = avgConfidence / float64(len(results))
	}
	
	cs.logger.Printf("Classification stats: %d total, %d classified, %d auto-classified", 
		stats.TotalProcessed, stats.Classified, stats.AutoClassified)
	
	return stats, nil
}

// ClassifySingleIssue classifies a single issue and returns the result
func (cs *ClassificationService) ClassifySingleIssue(issue *Issue) (*ClassificationResult, error) {
	if issue == nil {
		return nil, fmt.Errorf("issue cannot be nil")
	}
	
	result, err := cs.classifier.ClassifyIssue(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to classify issue: %w", err)
	}
	
	// Auto-classify if confidence is high enough
	if result.Confidence >= cs.config.AutoClassificationThreshold {
		issue.Category = result.PredictedCategory
		issue.Priority = result.PredictedPriority
		issue.TechStack = result.PredictedTechStack
		
		// Update database
		_, err := cs.db.Exec(`
			UPDATE issues SET category = ?, priority = ?, tech_stack = ?, updated_at_db = ?
			WHERE id = ?
		`, result.PredictedCategory, result.PredictedPriority, result.PredictedTechStack, time.Now(), issue.ID)
		
		if err != nil {
			cs.logger.Printf("Failed to update database with classification result: %v", err)
		}
	}
	
	return result, nil
}

// GetClassificationStats returns statistics about classification
func (cs *ClassificationService) GetClassificationStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get counts by category
	rows, err := cs.db.Query("SELECT category, COUNT(*) FROM issues WHERE category IS NOT NULL AND category != '' GROUP BY category ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	defer rows.Close()
	
	var categories []map[string]interface{}
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			continue
		}
		categories = append(categories, map[string]interface{}{
			"category": category,
			"count":    count,
		})
	}
	stats["categories"] = categories
	
	// Get counts by priority
	rows, err = cs.db.Query("SELECT priority, COUNT(*) FROM issues WHERE priority IS NOT NULL AND priority != '' GROUP BY priority ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to get priority stats: %w", err)
	}
	defer rows.Close()
	
	var priorities []map[string]interface{}
	for rows.Next() {
		var priority string
		var count int
		if err := rows.Scan(&priority, &count); err != nil {
			continue
		}
		priorities = append(priorities, map[string]interface{}{
			"priority": priority,
			"count":    count,
		})
	}
	stats["priorities"] = priorities
	
	// Get counts by tech stack
	rows, err = cs.db.Query("SELECT tech_stack FROM issues WHERE tech_stack IS NOT NULL AND tech_stack != '[]'")
	if err != nil {
		return nil, fmt.Errorf("failed to get tech stack stats: %w", err)
	}
	defer rows.Close()
	
	techStackCounts := make(map[string]int)
	for rows.Next() {
		var techStackStr string
		if err := rows.Scan(&techStackStr); err != nil {
			continue
		}
		
		// Parse JSON array (simplified - in real implementation, use proper JSON parsing)
		techStacks := parseTechStackString(techStackStr)
		for _, tech := range techStacks {
			if tech != "" {
				techStackCounts[tech]++
			}
		}
	}
	
	// Sort tech stacks by count
	var techStacks []map[string]interface{}
	for tech, count := range techStackCounts {
		techStacks = append(techStacks, map[string]interface{}{
			"tech_stack": tech,
			"count":      count,
		})
	}
	
	sort.Slice(techStacks, func(i, j int) bool {
		return techStacks[i]["count"].(int) > techStacks[j]["count"].(int)
	})
	
	stats["tech_stacks"] = techStacks
	
	return stats, nil
}

// parseTechStackString parses a tech stack string (simplified JSON parsing)
func parseTechStackString(techStackStr string) []string {
	// This is a simplified parser - in a real implementation, use proper JSON parsing
	if techStackStr == "[]" || techStackStr == "" {
		return []string{}
	}
	
	// Remove brackets and quotes, split by comma
	cleaned := strings.TrimPrefix(techStackStr, "[")
	cleaned = strings.TrimSuffix(cleaned, "]")
	cleaned = strings.ReplaceAll(cleaned, `"`, "")
	
	return strings.Split(cleaned, ",")
}