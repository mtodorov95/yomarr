package db

import (
	"database/sql"
	"strings"

	"github.com/mtodorov95/yomarr/internal/models"
)

type DownloadClientStore interface {
	GetAll() ([]models.DownloadClient, error)
	GetByID(id int64) (*models.DownloadClient, error)
	Insert(dc *models.DownloadClient) error
	Update(dc *models.DownloadClient) error
	Delete(id int64) error
}

type SQLiteDownloadClientStore struct{
	DB *sql.DB
}

func (store *SQLiteDownloadClientStore) GetAll() ([]models.DownloadClient, error) {
	rows, err := store.DB.Query("SELECT id, name, host, port, use_ssl, username, password, category FROM download_clients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.DownloadClient
	for rows.Next() {
		var dc models.DownloadClient
		if err := rows.Scan(&dc.ID, &dc.Name, &dc.Host, &dc.Port, &dc.UseSSL, &dc.Username, &dc.Password, &dc.Category); err != nil {
			return nil, err
		}
		list = append(list, dc)
	}

	for i := range list {
		if list[i].Password != "" {
			list[i].Password = strings.Repeat("*", len(list[i].Password))
		}
	}

	return list, nil
}

func (store *SQLiteDownloadClientStore) GetByID(id int64) (*models.DownloadClient, error) {
	var dc models.DownloadClient
	err := store.DB.QueryRow("SELECT id, name, host, port, use_ssl, username, password, category FROM download_clients WHERE id = ?", id).Scan(
		&dc.ID, &dc.Name, &dc.Host, &dc.Port, &dc.UseSSL, &dc.Username, &dc.Password, &dc.Category,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &dc, nil
}

func (store *SQLiteDownloadClientStore) Insert(dc *models.DownloadClient) error {
	res, err := store.DB.Exec(
		"INSERT INTO download_clients (name, host, port, use_ssl, username, password, category) VALUES (?, ?, ?, ?, ?, ?, ?)",
		dc.Name, dc.Host, dc.Port, dc.UseSSL, dc.Username, dc.Password, dc.Category,
	)
	if err != nil {
		return err
	}
	dc.ID, _ = res.LastInsertId()
	return nil
}

func (store *SQLiteDownloadClientStore) Update(dc *models.DownloadClient) error {
	_, err := store.DB.Exec(
		"UPDATE download_clients SET name=?, host=?, port=?, use_ssl=?, username=?, password=?, category=? WHERE id=?",
		dc.Name, dc.Host, dc.Port, dc.UseSSL, dc.Username, dc.Password, dc.Category, dc.ID,
	)
	return err
}

func (store *SQLiteDownloadClientStore) Delete(id int64) error {
	_, err := store.DB.Exec("DELETE FROM download_clients WHERE id = ?", id)
	return err
}
