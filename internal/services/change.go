package services

import (
	"database/sql"
	"fmt"
	"strings"
	"webapi/internal/models"
)

type ChangeService struct {
	db *sql.DB
}

func NewChangeService(db *sql.DB) *ChangeService {
	return &ChangeService{db: db}
}

func (s *ChangeService) GetChangesByIDs(changeIDs []int64) ([]models.ChangeResponse, error) {
	if len(changeIDs) == 0 {
		return []models.ChangeResponse{}, nil
	}

	// 构建 IN 查询
	placeholders := make([]string, len(changeIDs))
	args := make([]interface{}, len(changeIDs))
	for i, id := range changeIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT commit, summary, message 
		FROM changes 
		WHERE id IN (%s)
		ORDER BY id
	`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []models.ChangeResponse
	for rows.Next() {
		var change models.ChangeResponse
		if err := rows.Scan(&change.Commit, &change.Summary, &change.Message); err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}

	return changes, nil
}

func (s *ChangeService) InsertChanges(projectID string, changesData []models.ChangeResponse) ([]int64, error) {
	var changeIDs []int64

	for _, change := range changesData {
		var changeID int64
		err := s.db.QueryRow(`
			INSERT INTO changes (project, commit, summary, message)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, projectID, change.Commit, change.Summary, change.Message).Scan(&changeID)

		if err != nil {
			return nil, err
		}
		changeIDs = append(changeIDs, changeID)
	}

	return changeIDs, nil
}

func (s *ChangeService) GetChangeIDByCommitPrefix(projectID, commitPrefix string) (int, error) {
	rows, err := s.db.Query(`
		SELECT id FROM changes 
		WHERE project = $1 AND commit LIKE $2
	`, projectID, commitPrefix+"%")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var changeIDs []int
	for rows.Next() {
		var changeID int
		if err := rows.Scan(&changeID); err != nil {
			return 0, err
		}
		changeIDs = append(changeIDs, changeID)
	}

	if len(changeIDs) == 0 {
		return 0, fmt.Errorf("no changes found for version reference %s", commitPrefix)
	}

	if len(changeIDs) > 1 {
		return 0, fmt.Errorf("multiple changes found for version reference prefix %s. Please specify a more precise reference", commitPrefix)
	}

	return changeIDs[0], nil
}

func (s *ChangeService) GetBuildIDByChange(versionID, changeID int) (int, error) {
	var buildID int
	err := s.db.QueryRow(`
		SELECT build_id FROM builds 
		WHERE version = $1 AND changes @> $2
	`, versionID, fmt.Sprintf("{%d}", changeID)).Scan(&buildID)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("build not found for change")
		}
		return 0, err
	}

	return buildID, nil
}

func (s *ChangeService) ParseChanges(changesStr string) ([]models.ChangeResponse, error) {
	if changesStr == "" {
		return []models.ChangeResponse{}, nil
	}

	entries := strings.Split(changesStr, ">>>")
	var changes []models.ChangeResponse

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		parts := strings.Split(entry, "<<<")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid change format: %s", entry)
		}

		commit := strings.TrimSpace(parts[0])
		summary := strings.TrimSpace(parts[1])

		if commit == "" {
			return nil, fmt.Errorf("commit hash is missing in change entry")
		}
		if summary == "" {
			return nil, fmt.Errorf("summary is missing in change entry")
		}

		changes = append(changes, models.ChangeResponse{
			Commit:  commit,
			Summary: summary,
			Message: summary + "\n",
		})
	}

	return changes, nil
}