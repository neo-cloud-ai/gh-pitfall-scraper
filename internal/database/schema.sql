-- GitHub Pitfall Scraper Database Schema
-- 设计目标：支持 GitHub Issues 的去重、分类、时间序列分析
-- 版本：1.0.0

-- 设置 SQLite 优化选项
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 10000;
PRAGMA temp_store = memory;

-- 仓库表：存储被取的 Git爬Hub 仓库信息
CREATE TABLE IF NOT EXISTS repositories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner TEXT NOT NULL,                          -- 仓库所有者
    name TEXT NOT NULL,                           -- 仓库名称
    full_name TEXT UNIQUE NOT NULL,               -- 完整仓库名称 (owner/repo)
    description TEXT,                             -- 仓库描述
    url TEXT NOT NULL,                            -- 仓库URL
    stars INTEGER DEFAULT 0,                      -- Star数量
    forks INTEGER DEFAULT 0,                      -- Fork数量
    issues_count INTEGER DEFAULT 0,               -- Issues总数
    language TEXT,                                -- 主要编程语言
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 更新时间
    last_scraped_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 最后爬取时间
    is_active BOOLEAN DEFAULT 1,                  -- 是否活跃
    metadata TEXT,                                -- JSON格式的额外元数据
    CONSTRAINT unique_repository UNIQUE (owner, name)
);

-- 分类表：管理问题的分类和标签
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,                    -- 分类名称
    description TEXT,                             -- 分类描述
    color TEXT DEFAULT '#000000',                 -- 分类颜色标识
    is_active BOOLEAN DEFAULT 1,                  -- 是否启用
    priority INTEGER DEFAULT 0,                   -- 优先级（数值越大优先级越高）
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Issues 表：存储 GitHub Issues 信息
CREATE TABLE IF NOT EXISTS issues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    issue_id BIGINT UNIQUE NOT NULL,              -- GitHub原始Issue ID
    repository_id INTEGER NOT NULL,               -- 关联仓库ID
    number INTEGER NOT NULL,                      -- Issue编号
    title TEXT NOT NULL,                          -- Issue标题
    body TEXT,                                    -- Issue内容
    state TEXT NOT NULL,                          -- Issue状态 (open/closed)
    author_login TEXT NOT NULL,                   -- 作者用户名
    author_type TEXT DEFAULT 'User',              -- 作者类型 (User/Organization)
    labels TEXT,                                  -- JSON格式标签数组
    assignees TEXT,                               -- JSON格式指派人数组
    milestone TEXT,                               -- Milestone信息
    reactions TEXT,                               -- JSON格式反应统计
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- Issue创建时间
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- Issue更新时间
    closed_at DATETIME,                           -- Issue关闭时间
    first_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP, -- 首次发现时间
    last_seen_at DATETIME DEFAULT CURRENT_TIMESTAMP,  -- 最后更新时间
    is_pitfall BOOLEAN DEFAULT 0,                 -- 是否识别为坑点
    severity_score REAL DEFAULT 0,                -- 严重程度评分
    category_id INTEGER,                          -- 分类ID
    score REAL DEFAULT 0,                         -- 综合评分
    url TEXT NOT NULL,                            -- Issue URL
    html_url TEXT NOT NULL,                       -- Issue HTML URL
    comments_count INTEGER DEFAULT 0,             -- 评论数
    is_duplicate BOOLEAN DEFAULT 0,               -- 是否为重复Issue
    duplicate_of INTEGER,                         -- 关联的原始Issue ID
    metadata TEXT,                                -- JSON格式额外元数据
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    FOREIGN KEY (duplicate_of) REFERENCES issues(issue_id) ON DELETE SET NULL
);

-- 时间序列表：存储按时间维度聚合的数据
CREATE TABLE IF NOT EXISTS time_series (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repository_id INTEGER NOT NULL,               -- 关联仓库ID
    date DATE NOT NULL,                           -- 日期
    year INTEGER NOT NULL,                        -- 年份
    month INTEGER NOT NULL,                       -- 月份
    day INTEGER NOT NULL,                         -- 日期
    week_of_year INTEGER NOT NULL,                -- 年内周数
    day_of_week INTEGER NOT NULL,                 -- 星期几 (0-6)
    new_issues_count INTEGER DEFAULT 0,           -- 新增Issues数
    closed_issues_count INTEGER DEFAULT 0,        -- 关闭Issues数
    active_issues_count INTEGER DEFAULT 0,        -- 活跃Issues数
    pitfall_issues_count INTEGER DEFAULT 0,       -- 坑点Issues数
    avg_severity_score REAL DEFAULT 0,            -- 平均严重程度
    total_comments INTEGER DEFAULT 0,             -- 总评论数
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(repository_id, date),
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE
);

-- 索引创建：优化查询性能

-- Repositories 表索引
CREATE INDEX IF NOT EXISTS idx_repositories_owner ON repositories(owner);
CREATE INDEX IF NOT EXISTS idx_repositories_full_name ON repositories(full_name);
CREATE INDEX IF NOT EXISTS idx_repositories_stars ON repositories(stars DESC);
CREATE INDEX IF NOT EXISTS idx_repositories_last_scraped ON repositories(last_scraped_at DESC);
CREATE INDEX IF NOT EXISTS idx_repositories_is_active ON repositories(is_active);

-- Issues 表索引（最重要的查询优化）
CREATE UNIQUE INDEX IF NOT EXISTS idx_issues_github_id ON issues(issue_id);
CREATE INDEX IF NOT EXISTS idx_issues_repository ON issues(repository_id);
CREATE INDEX IF NOT EXISTS idx_issues_number ON issues(repository_id, number);
CREATE INDEX IF NOT EXISTS idx_issues_state ON issues(state);
CREATE INDEX IF NOT EXISTS idx_issues_created_at ON issues(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_issues_updated_at ON issues(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_issues_author ON issues(author_login);
CREATE INDEX IF NOT EXISTS idx_issues_category ON issues(category_id);
CREATE INDEX IF NOT EXISTS idx_issues_is_pitfall ON issues(is_pitfall);
CREATE INDEX IF NOT EXISTS idx_issues_severity_score ON issues(severity_score DESC);
CREATE INDEX IF NOT EXISTS idx_issues_score ON issues(score DESC);
CREATE INDEX IF NOT EXISTS idx_issues_comments_count ON issues(comments_count DESC);
CREATE INDEX IF NOT EXISTS idx_issues_first_seen ON issues(first_seen_at DESC);
CREATE INDEX IF NOT EXISTS idx_issues_is_duplicate ON issues(is_duplicate);

-- 复合索引：优化常用查询组合
CREATE INDEX IF NOT EXISTS idx_issues_repo_state ON issues(repository_id, state);
CREATE INDEX IF NOT EXISTS idx_issues_repo_category ON issues(repository_id, category_id);
CREATE INDEX IF NOT EXISTS idx_issues_pitfall_severity ON issues(is_pitfall, severity_score DESC);
CREATE INDEX IF NOT EXISTS idx_issues_duplicate_status ON issues(is_duplicate, duplicate_of);

-- Time Series 表索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_time_series_repo_date ON time_series(repository_id, date);
CREATE INDEX IF NOT EXISTS idx_time_series_date ON time_series(date DESC);
CREATE INDEX IF NOT EXISTS idx_time_series_repository ON time_series(repository_id);
CREATE INDEX IF NOT EXISTS idx_time_series_year_month ON time_series(year, month);
CREATE INDEX IF NOT EXISTS idx_time_series_new_issues ON time_series(new_issues_count DESC);
CREATE INDEX IF NOT EXISTS idx_time_series_pitfall_count ON time_series(pitfall_issues_count DESC);

-- Categories 表索引
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);
CREATE INDEX IF NOT EXISTS idx_categories_active ON categories(is_active);
CREATE INDEX IF NOT EXISTS idx_categories_priority ON categories(priority DESC);

-- 视图：简化常用查询

-- 活跃Issues视图
CREATE VIEW IF NOT EXISTS active_issues AS
SELECT 
    i.*,
    r.owner,
    r.name as repo_name,
    c.name as category_name
FROM issues i
JOIN repositories r ON i.repository_id = r.id
LEFT JOIN categories c ON i.category_id = c.id
WHERE i.state = 'open';

-- 坑点Issues视图
CREATE VIEW IF NOT EXISTS pitfall_issues AS
SELECT 
    i.*,
    r.owner,
    r.name as repo_name,
    c.name as category_name
FROM issues i
JOIN repositories r ON i.repository_id = r.id
LEFT JOIN categories c ON i.category_id = c.id
WHERE i.is_pitfall = 1;

-- 重复Issues视图
CREATE VIEW IF NOT EXISTS duplicate_issues AS
SELECT 
    i.*,
    r.owner,
    r.name as repo_name,
    c.name as category_name
FROM issues i
JOIN repositories r ON i.repository_id = r.id
LEFT JOIN categories c ON i.category_id = c.id
WHERE i.is_duplicate = 1;

-- 仓库统计视图
CREATE VIEW IF NOT EXISTS repository_stats AS
SELECT 
    r.*,
    COUNT(i.id) as total_issues,
    COUNT(CASE WHEN i.state = 'open' THEN 1 END) as open_issues,
    COUNT(CASE WHEN i.state = 'closed' THEN 1 END) as closed_issues,
    COUNT(CASE WHEN i.is_pitfall = 1 THEN 1 END) as pitfall_issues,
    COUNT(CASE WHEN i.is_duplicate = 1 THEN 1 END) as duplicate_issues,
    AVG(i.severity_score) as avg_severity_score,
    AVG(i.score) as avg_score,
    MAX(i.last_seen_at) as last_issue_update
FROM repositories r
LEFT JOIN issues i ON r.id = i.repository_id
GROUP BY r.id;

-- 初始化默认分类数据
INSERT OR IGNORE INTO categories (name, description, color, priority) VALUES
('Bug', 'Bug类问题', '#ff0000', 100),
('Documentation', '文档相关问题', '#0066cc', 80),
('Performance', '性能相关问题', '#ff6600', 90),
('Security', '安全问题', '#cc0000', 100),
('Usability', '可用性问题', '#00cc66', 70),
('Compatibility', '兼容性问题', '#9933ff', 75),
('API', 'API相关问题', '#0099cc', 85),
('UI/UX', '界面用户体验问题', '#ff9900', 65),
('Configuration', '配置相关问题', '#66cc66', 60),
('Integration', '集成相关问题', '#cc6600', 80),
('Testing', '测试相关问题', '#99cc00', 70),
('Deployment', '部署相关问题', '#ff3366', 85),
('Other', '其他问题', '#999999', 10);

-- 创建触发器：自动更新时间戳
CREATE TRIGGER IF NOT EXISTS update_repositories_timestamp 
AFTER UPDATE ON repositories
BEGIN
    UPDATE repositories SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_issues_timestamp 
AFTER UPDATE ON issues
BEGIN
    UPDATE issues SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_categories_timestamp 
AFTER UPDATE ON categories
BEGIN
    UPDATE categories SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_time_series_timestamp 
AFTER UPDATE ON time_series
BEGIN
    UPDATE time_series SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- 创建触发器：自动更新时间序列数据
CREATE TRIGGER IF NOT EXISTS insert_issue_time_series
AFTER INSERT ON issues
WHEN NEW.created_at IS NOT NULL
BEGIN
    INSERT OR IGNORE INTO time_series (repository_id, date, year, month, day, week_of_year, day_of_week)
    VALUES (
        NEW.repository_id,
        date(NEW.created_at),
        strftime('%Y', NEW.created_at),
        strftime('%m', NEW.created_at),
        strftime('%d', NEW.created_at),
        strftime('%W', NEW.created_at),
        CAST(strftime('%w', NEW.created_at) AS INTEGER)
    );
    
    UPDATE time_series 
    SET new_issues_count = new_issues_count + 1,
        active_issues_count = CASE WHEN NEW.state = 'open' THEN active_issues_count + 1 ELSE active_issues_count END,
        pitfall_issues_count = CASE WHEN NEW.is_pitfall = 1 THEN pitfall_issues_count + 1 ELSE pitfall_issues_count END
    WHERE repository_id = NEW.repository_id 
    AND date = date(NEW.created_at);
END;

CREATE TRIGGER IF NOT EXISTS update_issue_time_series
AFTER UPDATE ON issues
WHEN NEW.state != OLD.state OR NEW.is_pitfall != OLD.is_pitfall
BEGIN
    -- 更新关闭Issue的统计
    CASE 
        WHEN NEW.state = 'closed' AND OLD.state = 'open' THEN
            UPDATE time_series 
            SET closed_issues_count = closed_issues_count + 1,
                active_issues_count = active_issues_count - 1
            WHERE repository_id = NEW.repository_id 
            AND date = date(NEW.closed_at);
        WHEN NEW.state = 'open' AND OLD.state = 'closed' THEN
            UPDATE time_series 
            SET closed_issues_count = closed_issues_count - 1,
                active_issues_count = active_issues_count + 1
            WHERE repository_id = NEW.repository_id 
            AND date = date(NEW.updated_at);
    END;
    
    -- 更新坑点Issue的统计
    CASE 
        WHEN NEW.is_pitfall = 1 AND OLD.is_pitfall = 0 THEN
            UPDATE time_series 
            SET pitfall_issues_count = pitfall_issues_count + 1
            WHERE repository_id = NEW.repository_id 
            AND date = date(NEW.updated_at);
        WHEN NEW.is_pitfall = 0 AND OLD.is_pitfall = 1 THEN
            UPDATE time_series 
            SET pitfall_issues_count = pitfall_issues_count - 1
            WHERE repository_id = NEW.repository_id 
            AND date = date(NEW.updated_at);
    END;
END;