package db

import "github.com/mtodorov95/yomarr/internal/models"

type IndexerStore interface {
	GetAll() ([]models.Indexer, error)
	Insert(i *models.Indexer) error
}

type SQLiteIndexerStore struct{}

func (store *SQLiteIndexerStore) GetAll() ([]models.Indexer, error) {
	rows, err := DB.Query("SELECT id, name, url, api_key, priority FROM indexers ORDER BY priority DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Indexer
	for rows.Next() {
		var i models.Indexer
		if err := rows.Scan(&i.ID, &i.Name, &i.URL, &i.APIKey, &i.Priority); err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (store *SQLiteIndexerStore) Insert(i *models.Indexer) error {
	res, err := DB.Exec(
		"INSERT INTO indexers (name, url, api_key, priority) VALUES (?, ?, ?, ?)",
		i.Name, i.URL, i.APIKey, i.Priority,
	)
	if err != nil {
		return err
	}
	i.ID, _ = res.LastInsertId()
	return nil
}
