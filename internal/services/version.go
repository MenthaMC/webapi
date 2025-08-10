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