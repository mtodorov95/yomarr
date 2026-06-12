package db

import "github.com/mtodorov95/yomarr/internal/models"

type ChapterStore interface {
	GetBySeriesId(seriesId int64) ([]models.Chapters, error)
	Insert(c *models.Chapters) error
	Update(c *models.Chapters) error
	GetMissingBySeriesID(seriesID int64) ([]*models.Chapters, error)
	GetByStatus(status string) ([]models.Chapters, error)
	CountByStatus(status string) (int64, error)
}

type SQLiteChapterStore struct{}

func (store *SQLiteChapterStore) GetBySeriesId(seriesId int64) ([]models.Chapters, error) {
	rows, err := DB.Query(
		"SELECT id, series_id, number, volume, file_path, status, release_date, language FROM chapters WHERE series_id = ? ORDER BY number ASC",
		seriesId,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Chapters
	for rows.Next() {
		var c models.Chapters
		if err := rows.Scan(&c.ID, &c.SeriesID, &c.Number, &c.Volume, &c.FilePath, &c.Status, &c.ReleaseDate, &c.Language); err != nil {
			return nil, err
		}
		list = append(list, c)
	}

	return list, nil
}

func (store *SQLiteChapterStore) Insert(c *models.Chapters) error {
	res, err := DB.Exec(
		"INSERT INTO chapters (series_id, number, volume, file_path, status, release_date, language) VALUES (?,?,?,?,?,?,?)",
		c.SeriesID, c.Number, c.Volume, c.FilePath, c.Status, c.ReleaseDate, c.Language,
	)
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	return nil
}

func (store *SQLiteChapterStore) GetMissingBySeriesID(seriesID int64) ([]*models.Chapters, error) {
	query := `
		SELECT id, series_id, number, volume, status, language 
		FROM chapters 
		WHERE series_id = ? AND status = 'Missing'
		ORDER BY number ASC`

	rows, err := DB.Query(query, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Chapters
	for rows.Next() {
		var ch models.Chapters
		err := rows.Scan(
			&ch.ID,
			&ch.SeriesID,
			&ch.Number,
			&ch.Volume,
			&ch.Status,
			&ch.Language,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &ch)
	}

	return list, nil
}

func (store *SQLiteChapterStore) GetByStatus(status string) ([]models.Chapters, error) {
	query := `
		SELECT id, series_id, number, volume, status, language 
		FROM chapters 
		WHERE status = ?
		ORDER BY number ASC`

	rows, err := DB.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Chapters
	for rows.Next() {
		var ch models.Chapters
		err := rows.Scan(
			&ch.ID,
			&ch.SeriesID,
			&ch.Number,
			&ch.Volume,
			&ch.Status,
			&ch.Language,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, ch)
	}

	return list, nil
}

func (store *SQLiteChapterStore) Update(c *models.Chapters) error {
	query := `
		UPDATE chapters 
		SET volume = ?, status = ?, language = ?, file_path = ?
		WHERE id = ?
	`
	_, err := DB.Exec(
		query, 
		c.Volume, 
		c.Status, 
		c.Language, 
		c.FilePath, 
		c.ID,
	)
	return err
}

func (s *SQLiteChapterStore) CountByStatus(status string) (int64, error) {
	var count int64
	err := DB.QueryRow("SELECT COUNT(*) FROM chapters WHERE status = ?", status).Scan(&count)
	return count, err
}
