package services

import (
	"database/sql"
	"fmt"
	"webapi/internal/models"
)

type DownloadService struct {
	db *sql.DB
}

func NewDownloadService(db *sql.DB) *DownloadService {
	return &DownloadService{db: db}
}

func (s *DownloadService) GetDownloadURL(downloadSource, projectID, tag string) (string, error) {
	var url string
	err := s.db.QueryRow(`
		SELECT url FROM downloads 
		WHERE download_source = $1 AND project = $2 AND tag = $3
	`, downloadSource, projectID, tag).Scan(&url)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("download source not found")
		}
		return "", err
	}

	return url, nil
}

func (s *DownloadService) UpsertDownload(req models.CommitDownloadSourceRequest) error {
	// 先尝试更新
	result, err := s.db.Exec(`
		UPDATE downloads 
		SET url = $1 
		WHERE project = $2 AND tag = $3 AND download_source = $4
	`, req.URL, req.Project, req.Tag, req.DownloadSource)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// 如果没有更新任何行，则插入新记录
	if rowsAffected == 0 {
		_, err = s.db.Exec(`
			INSERT INTO downloads (project, tag, download_source, url)
			VALUES ($1, $2, $3, $4)
		`, req.Project, req.Tag, req.DownloadSource, req.URL)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *DownloadService) DeleteDownload(req models.DeleteDownloadSourceRequest) error {
	result, err := s.db.Exec(`
		DELETE FROM downloads 
		WHERE project = $1 AND tag = $2 AND download_source = $3
	`, req.Project, req.Tag, req.DownloadSource)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("download source %s for %s-%s not found", req.DownloadSource, req.Project, req.Tag)
	}

	return nil
}

func (s *DownloadService) DownloadSourceExists(projectID, tag, downloadSource string) (bool, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM downloads 
		WHERE project = $1 AND tag = $2 AND download_source = $3
	`, projectID, tag, downloadSource).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}