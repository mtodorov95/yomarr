package db

import (
	"database/sql"
	"encoding/json"

	"github.com/mtodorov95/yomarr/internal/models"
)

func ToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type SeriesStore interface {
	GetAll() ([]models.Series, error)
	GetById(id int64) (*models.Series, error)
	Insert(s *models.Series) error
	Delete(id int64) error
}

type SQLiteSeriesStore struct{}

func (store *SQLiteSeriesStore) GetAll() ([]models.Series, error) {
	query := `
		SELECT id, anilist_id, mangadex_id, title, alt_titles, path, status, 
		       total_chapters, thumbnail, historical_covers, author, genres, description 
		FROM series
		ORDER BY title ASC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []models.Series
	for rows.Next() {
		var s models.Series
		var altStr string
		var histStr string
		var genresStr string

		err := rows.Scan(
			&s.ID, &s.AnilistID, &s.MangadexID, &s.Title, &altStr, &s.Path, &s.Status,
			&s.TotalChapters, &s.Thumbnail, &histStr, &s.Author, &genresStr, &s.Description,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal([]byte(altStr), &s.AltTitles)
		_ = json.Unmarshal([]byte(histStr), &s.HistoricalCovers)
		_ = json.Unmarshal([]byte(genresStr), &s.Genres)
		list = append(list, s)
	}
	return list, nil
}

func (store *SQLiteSeriesStore) GetById(id int64) (*models.Series, error) {
	var s models.Series
	var altStr string
	var histStr string
	var genresStr string

	query := `
		SELECT id, anilist_id, mangadex_id, title, alt_titles, path, status, 
		       total_chapters, thumbnail, historical_covers, author, genres, description 
		FROM series 
		WHERE id = ?
	`

	err := DB.QueryRow(query, id).Scan(
		&s.ID, &s.AnilistID, &s.MangadexID, &s.Title, &altStr, &s.Path, &s.Status,
		&s.TotalChapters, &s.Thumbnail, &histStr, &s.Author, &genresStr, &s.Description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	_ = json.Unmarshal([]byte(altStr), &s.AltTitles)
	_ = json.Unmarshal([]byte(histStr), &s.HistoricalCovers)
	_ = json.Unmarshal([]byte(genresStr), &s.Genres)

	return &s, nil
}

func (store *SQLiteSeriesStore) Insert(s *models.Series) error {
	altJSON, _ := json.Marshal(s.AltTitles)
	histJSON, _ := json.Marshal(s.HistoricalCovers)
	genresJSON, _ := json.Marshal(s.Genres)

	query := `
		INSERT INTO series (
			anilist_id, mangadex_id, title, alt_titles, path, status, 
			total_chapters, thumbnail, historical_covers, author, genres, description
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := DB.Exec(
		query,
		s.AnilistID, s.MangadexID, s.Title, string(altJSON), s.Path, s.Status,
		s.TotalChapters, s.Thumbnail, string(histJSON), s.Author, string(genresJSON), s.Description,
	)

	if err != nil {
		return err
	}
	s.ID, _ = res.LastInsertId()
	return nil
}

func (store *SQLiteSeriesStore) Delete(id int64) error {
	_, err := DB.Exec("DELETE FROM series WHERE id = ?", id)
	return err
}
