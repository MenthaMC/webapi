package services

import (
	"database/sql"
	"fmt"
)

type VersionService struct {
	db *sql.DB
}

func NewVersionService(db *sql.DB) *VersionService {
	return &VersionService{db: db}
}

func (s *VersionService) GetVersionID(projectID, versionName string) (int, error) {
	var versionID int
	err := s.db.QueryRow(`
		SELECT id FROM versions 
		WHERE project = $1 AND name = $2
	`, projectID, versionName).Scan(&versionID)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("version not found")
		}
		return 0, err
	}
	
	return versionID, nil
}

func (s *VersionService) GetVersionGroupID(versionID int) (int, error) {
	var versionGroupID int
	err := s.db.QueryRow(`
		SELECT version_group FROM versions WHERE id = $1
	`, versionID).Scan(&versionGroupID)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("version not found")
		}
		return 0, err
	}
	
	return versionGroupID, nil
}

func (s *VersionService) GetVersionGroupIDByName(projectID, versionGroupName string) (int, error) {
	var versionGroupID int
	err := s.db.QueryRow(`
		SELECT id FROM version_groups 
		WHERE project = $1 AND name = $2
	`, projectID, versionGroupName).Scan(&versionGroupID)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("version group not found")
		}
		return 0, err
	}
	
	return versionGroupID, nil
}

func (s *VersionService) GetVersionsByGroupID(projectID string, versionGroupID int) ([]int, []string, error) {
	rows, err := s.db.Query(`
		SELECT id, name FROM versions 
		WHERE project = $1 AND version_group = $2
		ORDER BY name DESC
	`, projectID, versionGroupID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var ids []int
	var names []string
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		names = append(names, name)
	}

	return ids, names, nil
}

func (s *VersionService) GetLatestBuildID(projectID string, versionIDs []int) (int, error) {
	if len(versionIDs) == 0 {
		return 0, nil
	}

	query := `
		SELECT COALESCE(MAX(build_id), 0) 
		FROM builds 
		WHERE project = $1 AND version = ANY($2)
	`
	
	var latestBuildID int
	err := s.db.QueryRow(query, projectID, versionIDs).Scan(&latestBuildID)
	if err != nil {
		return 0, err
	}
	
	return latestBuildID, nil
}

// CreateVersionGroup 创建版本组
func (s *VersionService) CreateVersionGroup(projectID, name string) (int, error) {
	var versionGroupID int
	err := s.db.QueryRow(`
		INSERT INTO version_groups (project, name) 
		VALUES ($1, $2) 
		RETURNING id
	`, projectID, name).Scan(&versionGroupID)
	
	if err != nil {
		return 0, err
	}
	
	return versionGroupID, nil
}

// CreateVersion 创建版本
func (s *VersionService) CreateVersion(projectID, name string, versionGroupID int) (int, error) {
	var versionID int
	err := s.db.QueryRow(`
		INSERT INTO versions (name, project, version_group) 
		VALUES ($1, $2, $3) 
		RETURNING id
	`, name, projectID, versionGroupID).Scan(&versionID)
	
	if err != nil {
		return 0, err
	}
	
	return versionID, nil
}

// GetOrCreateVersionGroup 获取或创建版本组
func (s *VersionService) GetOrCreateVersionGroup(projectID, name string) (int, error) {
	// 先尝试获取现有的版本组
	versionGroupID, err := s.GetVersionGroupIDByName(projectID, name)
	if err == nil {
		return versionGroupID, nil
	}
	
	// 如果不存在，则创建新的版本组
	return s.CreateVersionGroup(projectID, name)
}

// GetOrCreateVersion 获取或创建版本
func (s *VersionService) GetOrCreateVersion(projectID, versionName string, versionGroupID int) (int, error) {
	// 先尝试获取现有版本
	versionID, err := s.GetVersionID(projectID, versionName)
	if err == nil {
		return versionID, nil
	}
	
	// 如果不存在，则创建新版本
	return s.CreateVersion(projectID, versionName, versionGroupID)
}
