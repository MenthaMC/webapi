package services

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
	"webapi-v2-neo/internal/models"

	"github.com/lib/pq"
)

type BuildService struct {
	db *sql.DB
}

func NewBuildService(db *sql.DB) *BuildService {
	return &BuildService{db: db}
}

func (s *BuildService) GetBuildsByVersion(projectID string, versionID int) ([]models.Build, error) {
	rows, err := s.db.Query(`
		SELECT id, project, build_id, time, experimental, jar_name, sha256, version, tag, changes, download_sources
		FROM builds 
		WHERE project = $1 AND version = $2
		ORDER BY build_id ASC
	`, projectID, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builds []models.Build
	for rows.Next() {
		var build models.Build
		if err := rows.Scan(
			&build.ID, &build.Project, &build.BuildID, &build.Time,
			&build.Experimental, &build.JarName, &build.SHA256,
			&build.Version, &build.Tag, &build.Changes, &build.DownloadSources,
		); err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}

	return builds, nil
}

func (s *BuildService) GetBuildsByVersions(projectID string, versionIDs []int) ([]models.Build, error) {
	if len(versionIDs) == 0 {
		return []models.Build{}, nil
	}

	rows, err := s.db.Query(`
		SELECT id, project, build_id, time, experimental, jar_name, sha256, version, tag, changes, download_sources
		FROM builds 
		WHERE project = $1 AND version = ANY($2)
		ORDER BY build_id ASC
	`, projectID, pq.Array(versionIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builds []models.Build
	for rows.Next() {
		var build models.Build
		if err := rows.Scan(
			&build.ID, &build.Project, &build.BuildID, &build.Time,
			&build.Experimental, &build.JarName, &build.SHA256,
			&build.Version, &build.Tag, &build.Changes, &build.DownloadSources,
		); err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}

	return builds, nil
}

func (s *BuildService) GetBuild(projectID string, versionID int, buildID int) (*models.Build, error) {
	var build models.Build
	err := s.db.QueryRow(`
		SELECT id, project, build_id, time, experimental, jar_name, sha256, version, tag, changes, download_sources
		FROM builds 
		WHERE project = $1 AND version = $2 AND build_id = $3
	`, projectID, versionID, buildID).Scan(
		&build.ID, &build.Project, &build.BuildID, &build.Time,
		&build.Experimental, &build.JarName, &build.SHA256,
		&build.Version, &build.Tag, &build.Changes, &build.DownloadSources,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("build not found")
		}
		return nil, err
	}

	return &build, nil
}

func (s *BuildService) ParseBuildID(projectID string, versionID int, buildIDStr string) (int, error) {
	if buildIDStr == "latest" {
		return s.getLatestBuildID(projectID, versionID)
	}

	buildID, err := strconv.Atoi(buildIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid build ID")
	}

	return buildID, nil
}

func (s *BuildService) getLatestBuildID(projectID string, versionID int) (int, error) {
	var buildID int
	err := s.db.QueryRow(`
		SELECT COALESCE(MAX(build_id), 0) 
		FROM builds 
		WHERE project = $1 AND version = $2
	`, projectID, versionID).Scan(&buildID)
	
	if err != nil {
		return 0, err
	}
	
	if buildID == 0 {
		return 0, fmt.Errorf("no builds found")
	}
	
	return buildID, nil
}

func (s *BuildService) CreateBuild(req models.CommitBuildRequest, versionID int, buildID int, changes []int64) error {
	experimental := req.Channel == "experimental"
	tag := req.Tag
	if len(req.Version) > 0 && len(tag) > len(req.Version)+1 {
		if tag[:len(req.Version)+1] == req.Version+"-" {
			tag = tag[len(req.Version)+1:]
		}
	}

	_, err := s.db.Exec(`
		INSERT INTO builds (project, build_id, time, experimental, jar_name, sha256, version, tag, changes, download_sources)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, req.ProjectID, buildID, time.Now(), experimental, req.JarName, req.SHA256, versionID, tag, pq.Array(changes), pq.Array([]string{"application"}))

	return err
}

func (s *BuildService) GetDownloadSources(projectID, tag string) ([]string, error) {
	var downloadSources pq.StringArray
	err := s.db.QueryRow(`
		SELECT download_sources FROM builds 
		WHERE project = $1 AND tag = $2
	`, projectID, tag).Scan(&downloadSources)

	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}, nil
		}
		return nil, err
	}

	return []string(downloadSources), nil
}

func (s *BuildService) AddDownloadSource(projectID, tag, downloadSource string) error {
	_, err := s.db.Exec(`
		UPDATE builds 
		SET download_sources = array_append(download_sources, $1) 
		WHERE project = $2 AND tag = $3
	`, downloadSource, projectID, tag)

	return err
}

func (s *BuildService) RemoveDownloadSource(projectID, tag, downloadSource string) error {
	_, err := s.db.Exec(`
		UPDATE builds 
		SET download_sources = array_remove(download_sources, $1) 
		WHERE project = $2 AND tag = $3
	`, downloadSource, projectID, tag)

	return err
}