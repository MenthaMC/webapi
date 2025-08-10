package models

import (
	"time"
)

// Release 表示一个发布版本
type Release struct {
	ID          int       `json:"id"`
	Project     string    `json:"project"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	TarballURL  string    `json:"tarball_url"`
	ZipballURL  string    `json:"zipball_url"`
	Assets      []ReleaseAsset `json:"assets"`
}

// ReleaseAsset 表示发布版本的附件
type ReleaseAsset struct {
	ID                 int       `json:"id"`
	ReleaseID          int       `json:"release_id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int64     `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}

// GitHubRelease GitHub API返回的Release结构
type GitHubRelease struct {
	ID          int                `json:"id"`
	TagName     string             `json:"tag_name"`
	Name        string             `json:"name"`
	Body        string             `json:"body"`
	Draft       bool               `json:"draft"`
	Prerelease  bool               `json:"prerelease"`
	CreatedAt   string             `json:"created_at"`
	PublishedAt string             `json:"published_at"`
	HTMLURL     string             `json:"html_url"`
	TarballURL  string             `json:"tarball_url"`
	ZipballURL  string             `json:"zipball_url"`
	Assets      []GitHubAsset      `json:"assets"`
}

// GitHubAsset GitHub API返回的Asset结构
type GitHubAsset struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Label              string `json:"label"`
	ContentType        string `json:"content_type"`
	State              string `json:"state"`
	Size               int64  `json:"size"`
	DownloadCount      int    `json:"download_count"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ReleaseConfig 拉取配置
type ReleaseConfig struct {
	ID           int    `json:"id"`
	Project      string `json:"project"`
	RepoOwner    string `json:"repo_owner"`
	RepoName     string `json:"repo_name"`
	AccessToken  string `json:"access_token,omitempty"` // 私有仓库需要
	AutoSync     bool   `json:"auto_sync"`
	SyncInterval int    `json:"sync_interval"` // 分钟
	LastSyncAt   *time.Time `json:"last_sync_at"`
	Enabled      bool   `json:"enabled"`
}