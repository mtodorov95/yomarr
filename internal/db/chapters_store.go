package db

import "github.com/mtodorov95/yomarr/internal/models"

type ChapterStore interface {
	GetBySeriesId(seriesId int64) ([]models.Chapters, error)
	Insert(c *models.Chapters) error
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
