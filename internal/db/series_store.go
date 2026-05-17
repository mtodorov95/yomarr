package db

import (
	"database/sql"

	"github.com/mtodorov95/yomarr/internal/models"
)

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

type SeriesStore interface {
	GetAll() ([]models.Series, error)
	GetById(id int64) (*models.Series, error)
	Insert(s *models.Series) error
}

type SQLiteSeriesStore struct{}

func (store *SQLiteSeriesStore) GetAll() ([]models.Series, error) {
	rows, err := DB.Query("SELECT id, anilist_id, mangadex_id, title, path, status FROM series")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(&s.ID, &s.AnilistID, &s.MangadexID, &s.Title, &s.Path, &s.Status); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (store *SQLiteSeriesStore) GetById(id int64) (*models.Series, error) {
	var s models.Series
	err := DB.QueryRow("SELECT id, anilist_id, mangadex_id, title, path, status FROM series WHERE id = ?", id).Scan(&s.ID, &s.AnilistID, &s.MangadexID, &s.Title, &s.Path, &s.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func (store *SQLiteSeriesStore) Insert(s *models.Series) error {
	res, err := DB.Exec(
		"INSERT INTO series (anilist_id, mangadex_id, title, path, status) VALUES (?,?,?,?,?)",
		nullIfEmpty(s.AnilistID), nullIfEmpty(s.MangadexID), s.Title, s.Path, s.Status,
	)
	if err != nil {
		return err
	}
	s.ID, _ = res.LastInsertId()
	return nil
}
