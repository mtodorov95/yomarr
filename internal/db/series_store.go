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
	GetAllMonitored() ([]models.Series, error)
	Insert(s *models.Series) error
	Update(s *models.Series) error
	Delete(id int64) error
	Count() (int64, error)
}

type SQLiteSeriesStore struct{
	DB *sql.DB
}

func (store *SQLiteSeriesStore) GetAll() ([]models.Series, error) {
	query := `
        SELECT id, anilist_id, mangadex_id, title, alt_titles, path, status, 
               total_chapters, thumbnail, historical_covers, author, artist, year, 
               genres, description, last_chapter, last_volume,
               (
                   SELECT COUNT(DISTINCT c.number) 
                   FROM chapters c 
                   WHERE c.series_id = series.id 
                     AND c.file_path IS NOT NULL 
                     AND c.file_path != ''
               ) AS downloaded_count
        FROM series
        ORDER BY title ASC
    `

	rows, err := store.DB.Query(query)
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
			&s.TotalChapters, &s.Thumbnail, &histStr, &s.Author, &s.Artist, &s.Year,
			&genresStr, &s.Description, &s.LastChapter, &s.LastVolume, &s.DownloadedCount,
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
               total_chapters, thumbnail, historical_covers, author, artist, year, 
               genres, description, last_chapter, last_volume,
               (
                   SELECT COUNT(DISTINCT c.number) 
                   FROM chapters c 
                   WHERE c.series_id = series.id 
                     AND c.file_path IS NOT NULL 
                     AND c.file_path != ''
               ) AS downloaded_count
        FROM series 
        WHERE id = ?
    `

	err := store.DB.QueryRow(query, id).Scan(
		&s.ID, &s.AnilistID, &s.MangadexID, &s.Title, &altStr, &s.Path, &s.Status,
		&s.TotalChapters, &s.Thumbnail, &histStr, &s.Author, &s.Artist, &s.Year,
		&genresStr, &s.Description, &s.LastChapter, &s.LastVolume, &s.DownloadedCount,
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

func (store *SQLiteSeriesStore) GetAllMonitored() ([]models.Series, error) {
	query := `
		SELECT id, title, alt_titles, status 
        FROM series
        WHERE status != ?
    `

	rows, err := store.DB.Query(query, models.SeriesUnmonitored)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var list []models.Series
	for rows.Next() {
		var s models.Series
		var altStr string

		err := rows.Scan(&s.ID, &s.Title, &altStr, &s.Status)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal([]byte(altStr), &s.AltTitles)
		list = append(list, s)
	}
	return list, nil
}

func (store *SQLiteSeriesStore) Insert(s *models.Series) error {
	altJSON, _ := json.Marshal(s.AltTitles)
	histJSON, _ := json.Marshal(s.HistoricalCovers)
	genresJSON, _ := json.Marshal(s.Genres)

	query := `
		INSERT INTO series (
			anilist_id, mangadex_id, title, alt_titles, path, status, 
			total_chapters, thumbnail, historical_covers, author, artist, year, 
			genres, description, last_chapter, last_volume
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := store.DB.Exec(
		query,
		s.AnilistID, s.MangadexID, s.Title, string(altJSON), s.Path, s.Status,
		s.TotalChapters, s.Thumbnail, string(histJSON), s.Author, s.Artist, s.Year,
		string(genresJSON), s.Description, s.LastChapter, s.LastVolume,
	)

	if err != nil {
		return err
	}
	s.ID, _ = res.LastInsertId()
	return nil
}

func (store *SQLiteSeriesStore) Update(s *models.Series) error {
	altJSON, _ := json.Marshal(s.AltTitles)
	histJSON, _ := json.Marshal(s.HistoricalCovers)
	genresJSON, _ := json.Marshal(s.Genres)

	query := `
		UPDATE series 
			SET anilist_id = ?, mangadex_id = ?, title = ?, alt_titles = ?, path = ?, status = ?, 
			    total_chapters = ?, thumbnail = ?, historical_covers = ?, author = ?, artist = ?, year = ?, 
			    genres = ?, description = ?, last_chapter = ?, last_volume = ?
		WHERE id = ?
	`

	_, err := store.DB.Exec(
		query,
		s.AnilistID, s.MangadexID, s.Title, string(altJSON), s.Path, s.Status,
		s.TotalChapters, s.Thumbnail, string(histJSON), s.Author, s.Artist, s.Year,
		string(genresJSON), s.Description, s.LastChapter, s.LastVolume, s.ID,
	)

	return err
}

func (store *SQLiteSeriesStore) Delete(id int64) error {
	_, err := store.DB.Exec("DELETE FROM series WHERE id = ?", id)
	return err
}

func (store *SQLiteSeriesStore) Count() (int64, error) {
	var count int64
	err := store.DB.QueryRow("SELECT COUNT(*) FROM series").Scan(&count)
	return count, err
}
