package services

import (
	"database/sql"
)

type VersionGroupService struct {
	db *sql.DB
}

func NewVersionGroupService(db *sql.DB) *VersionGroupService {
	return &VersionGroupService{db: db}
}

func (s *VersionGroupService) GetVersionGroupID(projectID, versionGroupName string) (int, error) {
	var versionGroupID int
	err := s.db.QueryRow(`
		SELECT id FROM version_groups 
		WHERE project = $1 AND name = $2
	`, projectID, versionGroupName).Scan(&versionGroupID)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	
	return versionGroupID, nil
}

func (s *VersionGroupService) GetVersionsByGroupID(projectID string, versionGroupID int) ([]int, []string, error) {
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