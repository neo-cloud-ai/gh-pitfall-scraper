-- gh-pitfall-scraper 数据库初始化脚本
-- 创建数据库表和索引

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建Issues表
CREATE TABLE IF NOT EXISTS issues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_owner VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    issue_number INTEGER NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    labels TEXT, -- JSON格式存储标签
    state VARCHAR(50) NOT NULL DEFAULT 'open',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    html_url TEXT NOT NULL,
    score DECIMAL(5,2) DEFAULT 0.0,
    severity DECIMAL(5,2) DEFAULT 0.0,
    created_at_db TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at_db TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 唯一约束
    CONSTRAINT unique_repo_issue UNIQUE (repo_owner, repo_name, issue_number)
);

-- 创建Indexes表
CREATE TABLE IF NOT EXISTS indexes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    index_name VARCHAR(255) NOT NULL,
    index_type VARCHAR(100),
    column_names TEXT, -- JSON格式存储列名
    is_unique BOOLEAN DEFAULT FALSE,
    is_primary BOOLEAN DEFAULT FALSE,
    size_bytes BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建ScrapingLogs表
CREATE TABLE IF NOT EXISTS scraping_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_owner VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL, -- success, failed, partial
    total_issues INTEGER DEFAULT 0,
    new_issues INTEGER DEFAULT 0,
    updated_issues INTEGER DEFAULT 0,
    error_message TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建Keywords表
CREATE TABLE IF NOT EXISTS keywords (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    keyword VARCHAR(255) NOT NULL UNIQUE,
    category VARCHAR(100),
    weight DECIMAL(3,2) DEFAULT 1.0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建RepositoryStats表
CREATE TABLE IF NOT EXISTS repository_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_owner VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    total_issues INTEGER DEFAULT 0,
    open_issues INTEGER DEFAULT 0,
    closed_issues INTEGER DEFAULT 0,
    last_scraped_at TIMESTAMP WITH TIME ZONE,
    avg_score DECIMAL(5,2) DEFAULT 0.0,
    avg_severity DECIMAL(5,2) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 唯一约束
    CONSTRAINT unique_repo_stats UNIQUE (repo_owner, repo_name)
);

-- 创建数据库配置表
CREATE TABLE IF NOT EXISTS database_config (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
-- Issues表索引
CREATE INDEX IF NOT EXISTS idx_issues_repo ON issues (repo_owner, repo_name);
CREATE INDEX IF NOT EXISTS idx_issues_state ON issues (state);
CREATE INDEX IF NOT EXISTS idx_issues_created_at ON issues (created_at);
CREATE INDEX IF NOT EXISTS idx_issues_updated_at ON issues (updated_at);
CREATE INDEX IF NOT EXISTS idx_issues_score ON issues (score);
CREATE INDEX IF NOT EXISTS idx_issues_severity ON issues (severity);
CREATE INDEX IF NOT EXISTS idx_issues_created_at_db ON issues (created_at_db);
CREATE INDEX IF NOT EXISTS idx_issues_updated_at_db ON issues (updated_at_db);

-- Indexes表索引
CREATE INDEX IF NOT EXISTS idx_indexes_issue_id ON indexes (issue_id);
CREATE INDEX IF NOT EXISTS idx_indexes_index_name ON indexes (index_name);

-- ScrapingLogs表索引
CREATE INDEX IF NOT EXISTS idx_scraping_logs_repo ON scraping_logs (repo_owner, repo_name);
CREATE INDEX IF NOT EXISTS idx_scraping_logs_status ON scraping_logs (status);
CREATE INDEX IF NOT EXISTS idx_scraping_logs_start_time ON scraping_logs (start_time);
CREATE INDEX IF NOT EXISTS idx_scraping_logs_created_at ON scraping_logs (created_at);

-- Keywords表索引
CREATE INDEX IF NOT EXISTS idx_keywords_category ON keywords (category);
CREATE INDEX IF NOT EXISTS idx_keywords_active ON keywords (is_active);

-- RepositoryStats表索引
CREATE INDEX IF NOT EXISTS idx_repo_stats_repo ON repository_stats (repo_owner, repo_name);
CREATE INDEX IF NOT EXISTS idx_repo_stats_last_scraped ON repository_stats (last_scraped_at);

-- 创建触发器函数：更新updated_at字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建触发器：自动更新updated_at字段
CREATE TRIGGER update_issues_updated_at 
    BEFORE UPDATE ON issues 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_indexes_updated_at 
    BEFORE UPDATE ON indexes 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_keywords_updated_at 
    BEFORE UPDATE ON keywords 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_repository_stats_updated_at 
    BEFORE UPDATE ON repository_stats 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_database_config_updated_at 
    BEFORE UPDATE ON database_config 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入初始配置数据
INSERT INTO database_config (key, value, description) VALUES 
('app_version', '2.0.0', '应用程序版本'),
('db_version', '1.0.0', '数据库架构版本'),
('init_comtrue', '数据库pleted', '初始化完成标识')
ON CONFLICT (key) DO NOTHING;

-- 插入示例关键词
INSERT INTO keywords (keyword, category, weight) VALUES 
('performance', '性能', 1.0),
('regression', '回归', 1.0),
('latency', '延迟', 1.0),
('throughput', '吞吐量', 1.0),
('OOM', '内存溢出', 1.0),
('memory leak', '内存泄漏', 1.0),
('CUDA', 'CUDA', 1.0),
('kernel', '内核', 1.0),
('NCCL', 'NCCL', 1.0),
('hang', '挂起', 1.0),
('deadlock', '死锁', 1.0),
('kv cache', 'KV缓存', 1.0)
ON CONFLICT (keyword) DO NOTHING;

-- 创建视图：Issues统计视图
CREATE OR REPLACE VIEW issues_stats AS
SELECT 
    repo_owner,
    repo_name,
    COUNT(*) as total_issues,
    COUNT(CASE WHEN state = 'open' THEN 1 END) as open_issues,
    COUNT(CASE WHEN state = 'closed' THEN 1 END) as closed_issues,
    ROUND(AVG(score), 2) as avg_score,
    ROUND(AVG(severity), 2) as avg_severity,
    MAX(updated_at) as last_updated
FROM issues 
GROUP BY repo_owner, repo_name;

-- 创建函数：清理过期数据
CREATE OR REPLACE FUNCTION cleanup_old_data(retention_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER := 0;
BEGIN
    -- 清理过期的scraping_logs
    DELETE FROM scraping_logs 
    WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '1 day' * retention_days;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- 记录清理操作
    INSERT INTO database_config (key, value, description) 
    VALUES ('last_cleanup', CURRENT_TIMESTAMP::text, '最后清理时间')
    ON CONFLICT (key) 
    DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 创建函数：更新仓库统计信息
CREATE OR REPLACE FUNCTION update_repository_stats()
RETURNS INTEGER AS $$
DECLARE
    repo_record RECORD;
    updated_count INTEGER := 0;
BEGIN
    FOR repo_record IN 
        SELECT repo_owner, repo_name 
        FROM issues 
        GROUP BY repo_owner, repo_name
    LOOP
        INSERT INTO repository_stats (
            repo_owner, repo_name, total_issues, open_issues, closed_issues,
            avg_score, avg_severity, last_scraped_at
        )
        SELECT 
            repo_record.repo_owner,
            repo_record.repo_name,
            COUNT(*),
            COUNT(CASE WHEN state = 'open' THEN 1 END),
            COUNT(CASE WHEN state = 'closed' THEN 1 END),
            ROUND(AVG(score), 2),
            ROUND(AVG(severity), 2),
            MAX(updated_at)
        FROM issues 
        WHERE repo_owner = repo_record.repo_owner AND repo_name = repo_record.repo_name
        ON CONFLICT (repo_owner, repo_name) 
        DO UPDATE SET
            total_issues = EXCLUDED.total_issues,
            open_issues = EXCLUDED.open_issues,
            closed_issues = EXCLUDED.closed_issues,
            avg_score = EXCLUDED.avg_score,
            avg_severity = EXCLUDED.avg_severity,
            last_scraped_at = EXCLUDED.last_scraped_at,
            updated_at = CURRENT_TIMESTAMP;
            
        updated_count := updated_count + 1;
    END LOOP;
    
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;

-- 注释
COMMENT ON TABLE issues IS '存储GitHub Issues信息';
COMMENT ON TABLE indexes IS '存储Issues的索引信息';
COMMENT ON TABLE scraping_logs IS '存储爬取日志';
COMMENT ON TABLE keywords IS '存储关键词配置';
COMMENT ON TABLE repository_stats IS '存储仓库统计信息';
COMMENT ON TABLE database_config IS '存储数据库配置信息';

COMMENT ON COLUMN issues.labels IS 'JSON格式存储的标签数组';
COMMENT ON COLUMN indexes.column_names IS 'JSON格式存储的列名数组';
COMMENT ON COLUMN issues.score IS '问题严重程度评分(0-10)';
COMMENT ON COLUMN issues.severity IS '问题严重性评分(0-10)';