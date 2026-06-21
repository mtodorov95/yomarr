package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/mtodorov95/yomarr/internal/models"
)


type QueueStore interface {
	Insert(item *models.QueueItem) error
	Get(hash string) (*models.QueueItem, error)
	GetAll() ([]models.QueueItem, error)
	UpdateStatus(hash string, status models.QueueStatus, errMsg *string) error
	Remove(hash string) error
}

type SQLiteQueueStore struct {}

func (s *SQLiteQueueStore) Insert(item *models.QueueItem) error {
	query := `
		INSERT INTO download_queue (torrent_hash, series_id, release_type, start_num, end_num, language, status, error_message, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	
	var errMsg sql.NullString
	if item.ErrorMessage != nil {
		errMsg = sql.NullString{String: *item.ErrorMessage, Valid: true}
	}

	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	_, err := DB.Exec(query,
		item.TorrentHash,
		item.SeriesID,
		string(item.ReleaseType),
		item.StartNum,
		item.EndNum,
		item.Language,
		string(item.Status),
		errMsg,
		item.CreatedAt,
	)
	return err
}

func (s *SQLiteQueueStore) Get(hash string) (*models.QueueItem, error) {
	query := `
		SELECT torrent_hash, series_id, release_type, start_num, end_num, language, status, error_message, created_at
		FROM download_queue
		WHERE torrent_hash = ?;
	`
	
	var item models.QueueItem
	var relType string
	var status string
	var errMsg sql.NullString

	err := DB.QueryRow(query, hash).Scan(
		&item.TorrentHash,
		&item.SeriesID,
		&relType,
		&item.StartNum,
		&item.EndNum,
		&item.Language,
		&status,
		&errMsg,
		&item.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	item.ReleaseType = models.ReleaseType(relType)
	item.Status = models.QueueStatus(status)
	if errMsg.Valid {
		item.ErrorMessage = &errMsg.String
	}

	return &item, nil
}

func (s *SQLiteQueueStore) GetAll() ([]models.QueueItem, error) {
	query := `
		SELECT torrent_hash, series_id, release_type, start_num, end_num, language, status, error_message, created_at
		FROM download_queue
		ORDER BY created_at DESC;
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.QueueItem
	for rows.Next() {
		var item models.QueueItem
		var relType string
		var status string
		var errMsg sql.NullString

		err := rows.Scan(
			&item.TorrentHash,
			&item.SeriesID,
			&relType,
			&item.StartNum,
			&item.EndNum,
			&item.Language,
			&status,
			&errMsg,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.ReleaseType = models.ReleaseType(relType)
		item.Status = models.QueueStatus(status)
		if errMsg.Valid {
			item.ErrorMessage = &errMsg.String
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *SQLiteQueueStore) UpdateStatus(hash string, status models.QueueStatus, errMsg *string) error {
	query := `
		UPDATE download_queue
		SET status = ?, error_message = ?
		WHERE torrent_hash = ?;
	`
	
	var dbErrMsg sql.NullString
	if errMsg != nil {
		dbErrMsg = sql.NullString{String: *errMsg, Valid: true}
	}

	_, err := DB.Exec(query, string(status), dbErrMsg, hash)
	return err
}

func (s *SQLiteQueueStore) Remove(hash string) error {
	query := `DELETE FROM download_queue WHERE torrent_hash = ?;`
	_, err := DB.Exec(query, hash)
	return err
}
