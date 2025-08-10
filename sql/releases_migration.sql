-- Release配置表
CREATE TABLE IF NOT EXISTS release_configs (
    id SERIAL PRIMARY KEY,
    project VARCHAR(255) NOT NULL UNIQUE,
    repo_owner VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    access_token TEXT,
    auto_sync BOOLEAN DEFAULT false,
    sync_interval INTEGER DEFAULT 60, -- 分钟
    last_sync_at TIMESTAMP,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Releases表
CREATE TABLE IF NOT EXISTS releases (
    id SERIAL PRIMARY KEY,
    project VARCHAR(255) NOT NULL,
    tag_name VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    body TEXT,
    draft BOOLEAN DEFAULT false,
    prerelease BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL,
    published_at TIMESTAMP NOT NULL,
    html_url TEXT,
    tarball_url TEXT,
    zipball_url TEXT,
    UNIQUE(project, tag_name)
);

-- Release资产表
CREATE TABLE IF NOT EXISTS release_assets (
    id SERIAL PRIMARY KEY,
    release_id INTEGER NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    label VARCHAR(255),
    content_type VARCHAR(255),
    state VARCHAR(50),
    size BIGINT,
    download_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    browser_download_url TEXT
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_releases_project ON releases(project);
CREATE INDEX IF NOT EXISTS idx_releases_published_at ON releases(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_releases_tag_name ON releases(tag_name);
CREATE INDEX IF NOT EXISTS idx_release_assets_release_id ON release_assets(release_id);
CREATE INDEX IF NOT EXISTS idx_release_configs_project ON release_configs(project);
CREATE INDEX IF NOT EXISTS idx_release_configs_auto_sync ON release_configs(auto_sync, enabled);

-- 添加外键约束
ALTER TABLE releases ADD CONSTRAINT fk_releases_project 
    FOREIGN KEY (project) REFERENCES projects(id) ON DELETE CASCADE;

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_release_configs_updated_at 
    BEFORE UPDATE ON release_configs 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入示例配置（可选）
-- INSERT INTO release_configs (project, repo_owner, repo_name, auto_sync, sync_interval, enabled)
-- VALUES ('paper', 'PaperMC', 'Paper', true, 30, true)
-- ON CONFLICT (project) DO NOTHING;