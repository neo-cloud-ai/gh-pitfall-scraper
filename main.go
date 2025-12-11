// gh-pitfall-scraper main.go 完整源代码
// 来源: https://github.com/neo-cloud-ai/gh-pitfall-scraper/blob/main/main.go
// 文件信息: 59 行代码 (48 行有效代码) · 1.17 KB

package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "gh-pitfall-scraper/internal/scraper"
    "gopkg.in/yaml.v3"
)

type Config struct {
    GithubToken string `yaml:"github_token"`
    Repos []struct {
        Owner string `yaml:"owner"`
        Name string `yaml:"name"`
    } `yaml:"repos"`
    Keywords []string `yaml:"keywords"`
}

func main() {
    // load config
    cfgFile, err := os.ReadFile("config.yaml")
    if err != nil {
        log.Fatalf("failed to read config: %v", err)
    }
    var cfg Config
    if err := yaml.Unmarshal(cfgFile, &cfg); err != nil {
        log.Fatalf("failed to parse config: %v", err)
    }

    client := scraper.NewGithubClient(cfg.GithubToken)
    var results []scraper.PitfallIssue
    for _, repo := range cfg.Repos {
        fmt.Printf("📦 Scraping: %s/%s...\n", repo.Owner, repo.Name)
        issues, err := scraper.ScrapeRepo(
            client,
            repo.Owner,
            repo.Name,
            cfg.Keywords,
        )
        if err != nil {
            log.Printf("Error scraping %s: %v\n", repo.Name, err)
            continue
        }
        results = append(results, issues...)
    }

    // output result
    output, _ := json.MarshalIndent(results, "", " ")
    os.WriteFile("output/issues.json", output, 0644)
    fmt.Println("🎉 Done. Output saved to output/issues.json")
}