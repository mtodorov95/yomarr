package db

import (
	"database/sql"

	"github.com/mtodorov95/yomarr/internal/models"
)

type IndexerStore interface {
	GetAll() ([]models.Indexer, error)
	Insert(i *models.Indexer) error
	Update(i *models.Indexer) error
	Delete(id int64) error
}

type SQLiteIndexerStore struct{
	DB *sql.DB
}

func (store *SQLiteIndexerStore) GetAll() ([]models.Indexer, error) {
	rows, err := store.DB.Query("SELECT id, name, url, api_key, priority, enable_rss, enable_search, additional_parameters, minimum_seeders, seed_time FROM indexers ORDER BY priority DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Indexer
	for rows.Next() {
		var i models.Indexer
		if err := rows.Scan(&i.ID, &i.Name, &i.URL, &i.APIKey, &i.Priority, &i.EnableRSS, &i.EnableSearch, &i.AdditionalParameters, &i.MinimumSeeders, &i.SeedTime); err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (store *SQLiteIndexerStore) Insert(i *models.Indexer) error {
	res, err := store.DB.Exec(
		"INSERT INTO indexers (name, url, api_key, priority, enable_rss, enable_search, additional_parameters, minimum_seeders, seed_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		i.Name, i.URL, i.APIKey, i.Priority, i.EnableRSS, i.EnableSearch, i.AdditionalParameters, i.MinimumSeeders, i.SeedTime,
	)
	if err != nil {
		return err
	}
	i.ID, _ = res.LastInsertId()
	return nil
}

func (store *SQLiteIndexerStore) Update(i *models.Indexer) error {
	_, err := store.DB.Exec(
		"UPDATE indexers SET name=?, url=?, api_key=?, priority=?, enable_rss=?, enable_search=?, additional_parameters=?, minimum_seeders=?, seed_time=? WHERE id=?",
		i.Name, i.URL, i.APIKey, i.Priority, i.EnableRSS, i.EnableSearch, i.AdditionalParameters, i.MinimumSeeders, i.SeedTime, i.ID,
	)
	return err
}

func (store *SQLiteIndexerStore) Delete(id int64) error {
	_, err := store.DB.Exec("DELETE FROM indexers WHERE id = ?", id)
	return err
}
