package models

import (
	"time"
	"github.com/lib/pq"
)

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Repo string `json:"repo"`
}

type Version struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Project      string `json:"project"`
	VersionGroup int    `json:"version_group"`
}

type VersionGroup struct {
	ID      int    `json:"id"`
	Project string `json:"project"`
	Name    string `json:"name"`
}

type Build struct {
	ID              int           `json:"id"`
	Project         string        `json:"project"`
	BuildID         int           `json:"build"`
	Time            time.Time     `json:"time"`
	Experimental    bool          `json:"experimental"`
	JarName         string        `json:"jar_name"`
	SHA256          string        `json:"sha256"`
	Version         int           `json:"version"`
	Tag             string        `json:"tag"`
	Changes         pq.Int64Array `json:"changes"`
	DownloadSources pq.StringArray `json:"download_sources"`
}

type BuildResponse struct {
	Build     int                    `json:"build"`
	Time      string                 `json:"time"`
	Channel   string                 `json:"channel"`
	Promoted  bool                   `json:"promoted"`
	Changes   []ChangeResponse       `json:"changes"`
	Downloads map[string]DownloadInfo `json:"downloads"`
}

type Change struct {
	ID      int    `json:"id"`
	Project string `json:"project"`
	Commit  string `json:"commit"`
	Summary string `json:"summary"`
	Message string `json:"message"`
}

type ChangeResponse struct {
	Commit  string `json:"commit"`
	Summary string `json:"summary"`
	Message string `json:"message"`
}

type Download struct {
	ID             int    `json:"id"`
	Project        string `json:"project"`
	Tag            string `json:"tag"`
	DownloadSource string `json:"download_source"`
	URL            string `json:"url"`
}

type DownloadInfo struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

type CommitBuildRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Version   string `json:"version" binding:"required"`
	Channel   string `json:"channel" binding:"required"`
	Changes   string `json:"changes" binding:"required"`
	JarName   string `json:"jar_name" binding:"required"`
	SHA256    string `json:"sha256" binding:"required"`
	Tag       string `json:"tag" binding:"required"`
}

type CommitDownloadSourceRequest struct {
	DownloadSource string `json:"download_source" binding:"required"`
	URL            string `json:"url" binding:"required"`
	Project        string `json:"project" binding:"required"`
	Tag            string `json:"tag" binding:"required"`
}

type DeleteDownloadSourceRequest struct {
	DownloadSource string `json:"download_source" binding:"required"`
	Project        string `json:"project" binding:"required"`
	Tag            string `json:"tag" binding:"required"`
}