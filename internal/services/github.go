package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"webapi/internal/logger"
	"webapi/internal/models"
)

type GitHubService struct {
	client *http.Client
	token  string
}

type GitHubRelease struct {
	ID          int64  `json:"id"`
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	CreatedAt   string `json:"created_at"`
	PublishedAt string `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	Size               int64  `json:"size"`
	DownloadCount      int64  `json:"download_count"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func NewGitHubService(token string) *GitHubService {
	return &GitHubService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}
}

// FetchReleases 从GitHub仓库获取所有releases
func (g *GitHubService) FetchReleases(repoURL string) ([]GitHubRelease, error) {
	// 解析仓库URL，提取owner和repo
	owner, repo, err := g.parseRepoURL(repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %v", err)
	}

	// 构建API URL
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 添加认证头
	if g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var releases []GitHubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %v", err)
	}

	return releases, nil
}

// parseRepoURL 解析GitHub仓库URL
func (g *GitHubService) parseRepoURL(repoURL string) (owner, repo string, err error) {
	// 支持多种格式的URL
	// https://github.com/owner/repo
	// https://github.com/owner/repo.git
	// git@github.com:owner/repo.git
	
	repoURL = strings.TrimSuffix(repoURL, ".git")
	
	if strings.HasPrefix(repoURL, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	} else if strings.HasPrefix(repoURL, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(repoURL, "git@github.com:"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}
	
	return "", "", fmt.Errorf("unsupported repository URL format")
}

// ConvertToVersionAndBuild 将GitHub Release转换为版本和构建信息
func (g *GitHubService) ConvertToVersionAndBuild(release GitHubRelease, projectID string) (*models.Version, *models.Build, error) {
	// 解析版本名称
	versionName := g.extractVersionName(release.TagName)
	
	// 解析时间
	publishedAt, err := time.Parse(time.RFC3339, release.PublishedAt)
	if err != nil {
		publishedAt = time.Now()
	}

	// 查找JAR文件
	var jarAsset *GitHubAsset
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".jar") {
			jarAsset = &asset
			break
		}
	}

	if jarAsset == nil {
		return nil, nil, fmt.Errorf("no JAR file found in release %s", release.TagName)
	}

	// 创建版本信息
	version := &models.Version{
		Name:    versionName,
		Project: projectID,
		// VersionGroup 需要根据实际情况设置
	}

	// 创建构建信息
	build := &models.Build{
		Project:      projectID,
		Time:         publishedAt,
		Experimental: release.Prerelease,
		JarName:      jarAsset.Name,
		SHA256:       "", // GitHub API不直接提供SHA256，需要单独计算
		Tag:          release.TagName,
		Changes:      []int64{}, // 需要从release body解析
		DownloadSources: []string{"github"},
	}

	return version, build, nil
}

// extractVersionName 从tag名称中提取版本名称
func (g *GitHubService) extractVersionName(tagName string) string {
	// 移除常见的前缀
	tagName = strings.TrimPrefix(tagName, "v")
	tagName = strings.TrimPrefix(tagName, "release-")
	tagName = strings.TrimPrefix(tagName, "version-")
	
	return tagName
}

// SyncProjectReleases 同步项目的所有releases
func (g *GitHubService) SyncProjectReleases(project *models.Project, services *Services) error {
	logger.Infof("Starting to sync releases for project: %s", project.ID)
	
	releases, err := g.FetchReleases(project.Repo)
	if err != nil {
		return fmt.Errorf("failed to fetch releases for project %s: %v", project.ID, err)
	}

	logger.Infof("Found %d releases for project %s", len(releases), project.ID)

	for _, release := range releases {
		// 跳过草稿版本
		if release.Draft {
			continue
		}

		err := g.syncSingleRelease(release, project, services)
		if err != nil {
			logger.Errorf("Failed to sync release %s for project %s: %v", release.TagName, project.ID, err)
			continue
		}
		
		logger.Infof("Successfully synced release %s for project %s", release.TagName, project.ID)
	}

	return nil
}

// syncSingleRelease 同步单个release
func (g *GitHubService) syncSingleRelease(release GitHubRelease, project *models.Project, services *Services) error {
	// 检查版本是否已存在
	versionName := g.extractVersionName(release.TagName)
	_, err := services.Version.GetVersionID(project.ID, versionName)
	if err == nil {
		// 版本已存在，跳过
		logger.Infof("Version %s already exists for project %s, skipping", versionName, project.ID)
		return nil
	}

	// 转换为内部数据结构
	_, build, err := g.ConvertToVersionAndBuild(release, project.ID)
	if err != nil {
		return err
	}

	// 创建或获取版本组（使用主版本号作为版本组名）
	versionGroupName := g.extractVersionGroup(versionName)
	versionGroupID, err := services.Version.GetOrCreateVersionGroup(project.ID, versionGroupName)
	if err != nil {
		return fmt.Errorf("failed to create version group: %v", err)
	}

	// 创建版本
	versionID, err := services.Version.CreateVersion(project.ID, versionName, versionGroupID)
	if err != nil {
		return fmt.Errorf("failed to create version: %v", err)
	}

	// 获取下一个构建ID
	latestBuildID, err := services.Version.GetLatestBuildID(project.ID, []int{versionID})
	if err != nil {
		latestBuildID = 0 // 如果没有构建，从0开始
	}
	newBuildID := latestBuildID + 1

	// 解析变更信息（从release body）
	changesData := g.parseChangesFromReleaseBody(release.Body, project.ID)
	changeIDs, err := services.Change.InsertChanges(project.ID, changesData)
	if err != nil {
		logger.Errorf("Failed to insert changes for release %s: %v", release.TagName, err)
		changeIDs = []int64{} // 如果失败，使用空的变更列表
	}

	// 设置构建信息
	build.BuildID = newBuildID
	build.Version = versionID
	build.Changes = changeIDs

	// 创建构建记录
	err = g.createBuildFromRelease(build, services)
	if err != nil {
		return fmt.Errorf("failed to create build: %v", err)
	}

	// 创建下载记录
	err = g.createDownloadRecords(release, project.ID, release.TagName, services)
	if err != nil {
		logger.Errorf("Failed to create download records for release %s: %v", release.TagName, err)
	}

	logger.Infof("Successfully synced release %s for project %s", release.TagName, project.ID)
	return nil
}

// extractVersionGroup 从版本名称中提取版本组名称
func (g *GitHubService) extractVersionGroup(versionName string) string {
	// 简单的版本组提取逻辑，可以根据需要调整
	// 例如：1.20.4 -> 1.20, 1.19.2 -> 1.19
	parts := strings.Split(versionName, ".")
	if len(parts) >= 2 {
		return fmt.Sprintf("%s.%s", parts[0], parts[1])
	}
	return versionName
}

// parseChangesFromReleaseBody 从release body解析变更信息
func (g *GitHubService) parseChangesFromReleaseBody(body, projectID string) []models.ChangeResponse {
	// 简单的解析逻辑，可以根据实际格式调整
	changes := []models.ChangeResponse{}
	
	if body == "" {
		return changes
	}

	// 创建一个通用的变更记录
	changes = append(changes, models.ChangeResponse{
		Commit:  "github-release",
		Summary: "GitHub Release",
		Message: body,
	})

	return changes
}

// createBuildFromRelease 从release创建构建记录
func (g *GitHubService) createBuildFromRelease(build *models.Build, services *Services) error {
	_, err := services.Build.db.Exec(`
		INSERT INTO builds (project, build_id, time, experimental, jar_name, sha256, version, tag, changes, download_sources)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, build.Project, build.BuildID, build.Time, build.Experimental, build.JarName, build.SHA256, 
	   build.Version, build.Tag, build.Changes, build.DownloadSources)
	
	return err
}

// createDownloadRecords 创建下载记录
func (g *GitHubService) createDownloadRecords(release GitHubRelease, projectID, tag string, services *Services) error {
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".jar") {
			_, err := services.Download.db.Exec(`
				INSERT INTO downloads (project, tag, download_source, url)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (project, tag, download_source) DO UPDATE SET url = EXCLUDED.url
			`, projectID, tag, "github", asset.BrowserDownloadURL)
			
			if err != nil {
				return err
			}
		}
	}
	return nil
}
