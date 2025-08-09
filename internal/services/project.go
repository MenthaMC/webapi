package services

import (
	"database/sql"
	"webapi-v2-neo/internal/models"
)

type ProjectService struct {
	db *sql.DB
}

func NewProjectService(db *sql.DB) *ProjectService {
	return &ProjectService{db: db}
}

func (s *ProjectService) GetAll() ([]models.Project, error) {
	rows, err := s.db.Query("SELECT id, name, repo FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		if err := rows.Scan(&project.ID, &project.Name, &project.Repo); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (s *ProjectService) GetByID(projectID string) (*models.Project, error) {
	var project models.Project
	err := s.db.QueryRow("SELECT id, name, repo FROM projects WHERE id = $1", projectID).
		Scan(&project.ID, &project.Name, &project.Repo)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &project, nil
}

func (s *ProjectService) GetVersions(projectID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT v.name 
		FROM versions v 
		WHERE v.project = $1 
		ORDER BY v.name DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

func (s *ProjectService) GetVersionGroups(projectID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT vg.name 
		FROM version_groups vg 
		WHERE vg.project = $1 
		ORDER BY vg.name DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versionGroups []string
	for rows.Next() {
		var versionGroup string
		if err := rows.Scan(&versionGroup); err != nil {
			return nil, err
		}
		versionGroups = append(versionGroups, versionGroup)
	}

	return versionGroups, nil
}