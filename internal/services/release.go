package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"webapi/internal/logger"
	"webapi/internal/models"
)

type ReleaseService struct {
	db *sql.DB
}

func NewReleaseService(db *sql.DB) *ReleaseService {
	return &ReleaseService{db: db}
}

// GetReleaseConfig 获取项目的Release配置
func (s *ReleaseService) GetReleaseConfig(projectID string) (*models.ReleaseConfig, error) {
	var config models.ReleaseConfig
	var lastSyncAt sql.NullTime
	
	err := s.db.QueryRow(`
		SELECT id, project, repo_owner, repo_name, access_token, auto_sync, 
		       sync_interval, last_sync_at, enabled 
		FROM release_configs WHERE project = $1
	`, projectID).Scan(
		&config.ID, &config.Project, &config.RepoOwner, &config.RepoName,
		&config.AccessToken, &config.AutoSync, &config.SyncInterval,
		&lastSyncAt, &config.Enabled,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	if lastSyncAt.Valid {
		config.LastSyncAt = &lastSyncAt.Time
	}
	
	return &config, nil
}

// SaveReleaseConfig 保存Release配置
func (s *ReleaseService) SaveReleaseConfig(config *models.ReleaseConfig) error {
	_, err := s.db.Exec(`
		INSERT INTO release_configs (project, repo_owner, repo_name, access_token, 
		                           auto_sync, sync_interval, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (project) DO UPDATE SET
		    repo_owner = EXCLUDED.repo_owner,
		    repo_name = EXCLUDED.repo_name,
		    access_token = EXCLUDED.access_token,
		    auto_sync = EXCLUDED.auto_sync,
		    sync_interval = EXCLUDED.sync_interval,
		    enabled = EXCLUDED.enabled
	`, config.Project, config.RepoOwner, config.RepoName, config.AccessToken,
		config.AutoSync, config.SyncInterval, config.Enabled)
	
	return err
}

// FetchReleasesFromGitHub 从GitHub拉取Releases
func (s *ReleaseService) FetchReleasesFromGitHub(config *models.ReleaseConfig) ([]models.GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", config.RepoOwner, config.RepoName)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// 如果有访问令牌，添加认证头
	if config.AccessToken != "" {
		req.Header.Set("Authorization", "token "+config.AccessToken)
	}
	
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "MenthaMC-WebAPI")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var releases []models.GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	
	return releases, nil
}

// SyncReleases 同步Releases到数据库
func (s *ReleaseService) SyncReleases(projectID string) error {
	config, err := s.GetReleaseConfig(projectID)
	if err != nil {
		return err
	}
	
	if config == nil || !config.Enabled {
		return fmt.Errorf("release config not found or disabled for project: %s", projectID)
	}
	
	logger.Info(fmt.Sprintf("开始同步项目 %s 的 Releases", projectID))
	
	// 从GitHub获取Releases
	githubReleases, err := s.FetchReleasesFromGitHub(config)
	if err != nil {
		return fmt.Errorf("failed to fetch releases from GitHub: %v", err)
	}
	
	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	syncCount := 0
	for _, ghRelease := range githubReleases {
		// 检查Release是否已存在
		var existingID int
		err := tx.QueryRow("SELECT id FROM releases WHERE project = $1 AND tag_name = $2", 
			projectID, ghRelease.TagName).Scan(&existingID)
		
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		
		// 解析时间
		createdAt, _ := time.Parse(time.RFC3339, ghRelease.CreatedAt)
		publishedAt, _ := time.Parse(time.RFC3339, ghRelease.PublishedAt)
		
		if err == sql.ErrNoRows {
			// 插入新Release
			var releaseID int
			err = tx.QueryRow(`
				INSERT INTO releases (project, tag_name, name, body, draft, prerelease, 
				                    created_at, published_at, html_url, tarball_url, zipball_url)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
				RETURNING id
			`, projectID, ghRelease.TagName, ghRelease.Name, ghRelease.Body,
				ghRelease.Draft, ghRelease.Prerelease, createdAt, publishedAt,
				ghRelease.HTMLURL, ghRelease.TarballURL, ghRelease.ZipballURL).Scan(&releaseID)
			
			if err != nil {
				return err
			}
			
			// 插入Assets
			for _, asset := range ghRelease.Assets {
				assetCreatedAt, _ := time.Parse(time.RFC3339, asset.CreatedAt)
				assetUpdatedAt, _ := time.Parse(time.RFC3339, asset.UpdatedAt)
				
				_, err = tx.Exec(`
					INSERT INTO release_assets (release_id, name, label, content_type, state, 
					                          size, download_count, created_at, updated_at, browser_download_url)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
				`, releaseID, asset.Name, asset.Label, asset.ContentType, asset.State,
					asset.Size, asset.DownloadCount, assetCreatedAt, assetUpdatedAt, asset.BrowserDownloadURL)
				
				if err != nil {
					return err
				}
			}
			
			syncCount++
		} else {
			// 更新现有Release
			_, err = tx.Exec(`
				UPDATE releases SET 
				    name = $3, body = $4, draft = $5, prerelease = $6,
				    published_at = $7, html_url = $8, tarball_url = $9, zipball_url = $10
				WHERE project = $1 AND tag_name = $2
			`, projectID, ghRelease.TagName, ghRelease.Name, ghRelease.Body,
				ghRelease.Draft, ghRelease.Prerelease, publishedAt,
				ghRelease.HTMLURL, ghRelease.TarballURL, ghRelease.ZipballURL)
			
			if err != nil {
				return err
			}
		}
	}
	
	// 更新最后同步时间
	_, err = tx.Exec("UPDATE release_configs SET last_sync_at = $1 WHERE project = $2", 
		time.Now(), projectID)
	if err != nil {
		return err
	}
	
	if err = tx.Commit(); err != nil {
		return err
	}
	
	logger.Info(fmt.Sprintf("项目 %s 同步完成，新增/更新了 %d 个 Releases", projectID, syncCount))
	return nil
}

// GetReleases 获取项目的Releases
func (s *ReleaseService) GetReleases(projectID string, limit, offset int) ([]models.Release, error) {
	rows, err := s.db.Query(`
		SELECT id, project, tag_name, name, body, draft, prerelease, 
		       created_at, published_at, html_url, tarball_url, zipball_url
		FROM releases 
		WHERE project = $1 
		ORDER BY published_at DESC 
		LIMIT $2 OFFSET $3
	`, projectID, limit, offset)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var releases []models.Release
	for rows.Next() {
		var release models.Release
		err := rows.Scan(
			&release.ID, &release.Project, &release.TagName, &release.Name,
			&release.Body, &release.Draft, &release.Prerelease,
			&release.CreatedAt, &release.PublishedAt, &release.HTMLURL,
			&release.TarballURL, &release.ZipballURL,
		)
		if err != nil {
			return nil, err
		}
		
		// 获取Assets
		assets, err := s.GetReleaseAssets(release.ID)
		if err != nil {
			return nil, err
		}
		release.Assets = assets
		
		releases = append(releases, release)
	}
	
	return releases, nil
}

// GetReleaseAssets 获取Release的Assets
func (s *ReleaseService) GetReleaseAssets(releaseID int) ([]models.ReleaseAsset, error) {
	rows, err := s.db.Query(`
		SELECT id, release_id, name, label, content_type, state, size, 
		       download_count, created_at, updated_at, browser_download_url
		FROM release_assets 
		WHERE release_id = $1
		ORDER BY name
	`, releaseID)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var assets []models.ReleaseAsset
	for rows.Next() {
		var asset models.ReleaseAsset
		err := rows.Scan(
			&asset.ID, &asset.ReleaseID, &asset.Name, &asset.Label,
			&asset.ContentType, &asset.State, &asset.Size, &asset.DownloadCount,
			&asset.CreatedAt, &asset.UpdatedAt, &asset.BrowserDownloadURL,
		)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	
	return assets, nil
}

// GetLatestRelease 获取最新的Release
func (s *ReleaseService) GetLatestRelease(projectID string) (*models.Release, error) {
	var release models.Release
	err := s.db.QueryRow(`
		SELECT id, project, tag_name, name, body, draft, prerelease, 
		       created_at, published_at, html_url, tarball_url, zipball_url
		FROM releases 
		WHERE project = $1 AND draft = false
		ORDER BY published_at DESC 
		LIMIT 1
	`, projectID).Scan(
		&release.ID, &release.Project, &release.TagName, &release.Name,
		&release.Body, &release.Draft, &release.Prerelease,
		&release.CreatedAt, &release.PublishedAt, &release.HTMLURL,
		&release.TarballURL, &release.ZipballURL,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	// 获取Assets
	assets, err := s.GetReleaseAssets(release.ID)
	if err != nil {
		return nil, err
	}
	release.Assets = assets
	
	return &release, nil
}