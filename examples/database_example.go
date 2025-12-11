package main

import (
	"fmt"
	"log"
	"time"

	"github.com/your-repo/gh-pitfall-scraper/internal/database"
)

func main() {
	// 1. 配置数据库
	config := database.DefaultDatabaseConfig()
	config.Path = "./data/issues.db"
	config.MaxConnections = 10
	
	// 2. 创建数据库实例
	db, err := database.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// 3. 初始化数据库
	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 4. 演示基本操作
	fmt.Println("Database initialized successfully!")
	
	// 5. 演示数据库统计
	stats, err := db.GetDatabaseStats()
	if err != nil {
		log.Printf("Failed to get database stats: %v", err)
	} else {
		fmt.Printf("Database Stats: %+v\n", stats)
	}
}