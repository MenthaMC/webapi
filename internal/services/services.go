package services

import (
	"database/sql"
)

type Services struct {
	Project     *ProjectService
	Version     *VersionService
	Build       *BuildService
	Download    *DownloadService
	Change      *ChangeService
	VersionGroup *VersionGroupService
}

func New(db *sql.DB) *Services {
	return &Services{
		Project:      NewProjectService(db),
		Version:      NewVersionService(db),
		Build:        NewBuildService(db),
		Download:     NewDownloadService(db),
		Change:       NewChangeService(db),
		VersionGroup: NewVersionGroupService(db),
	}
}