package services

// add file to sqlite db

import (
	"context"
	"database/sql"
	"log"
	"time"
	"zgdrive/model"

	_ "github.com/mattn/go-sqlite3"
)

type DBService struct {
	db *sql.DB
}

func NewDBService(dbname string) *DBService {
	// create db if not exists
	db, err := sql.Open("sqlite3", dbname)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil
	}

	// create table if not exists
	query := `
		CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL,
			hash TEXT NOT NULL,
			size INTEGER NOT NULL,
			tx_id TEXT DEFAULT NULL,
			is_uploaded BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT (datetime('now','localtime'))
		);
		CREATE TABLE IF NOT EXISTS downloaded_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			filename TEXT NOT NULL,
			hash TEXT NOT NULL,
			size INTEGER NOT NULL,
			downloaded_at TIMESTAMP DEFAULT (datetime('now','localtime')),
			is_processing BOOLEAN NOT NULL DEFAULT TRUE,
			is_removed BOOLEAN NOT NULL DEFAULT FALSE
		);
	`
	_, err = db.Exec(query)
	if err != nil {
		log.Printf("Error creating table: %v", err)
		db.Close()
		return nil
	}

	return &DBService{db: db}
}

func (d *DBService) AddFile(ctx context.Context, filename, hash string, size int64) (model.File, error) {
	query := `
		INSERT INTO files (filename, hash, size)
		VALUES (?, ?, ?) RETURNING id, filename, size, is_uploaded, created_at
	`
	var id int64
	var isUploaded bool
	var createdAt time.Time
	err := d.db.QueryRowContext(ctx, query, filename, hash, size).Scan(&id, &filename, &size, &isUploaded, &createdAt)
	if err != nil {
		return model.File{}, err
	}

	return model.File{
		ID:         id,
		Filename:   filename,
		Size:       size,
		IsUploaded: isUploaded,
		CreatedAt:  createdAt,
	}, nil
}

func (d *DBService) SetUploaded(ctx context.Context, filename string) error {
	query := `
		UPDATE files
		SET is_uploaded = TRUE
		WHERE filename = ?
	`
	_, err := d.db.ExecContext(ctx, query, filename)
	if err != nil {
		return err
	}
	return nil
}

func (d *DBService) GetFileById(ctx context.Context, fileId int64) (model.File, error) {
	query := `
		SELECT id, filename, size, hash, tx_id, is_uploaded, created_at
		FROM files
		WHERE id = ?
	`
	row := d.db.QueryRowContext(ctx, query, fileId)
	var id int64
	var filename string
	var size int64
	var hash string
	var txId sql.NullString
	var isUploaded bool
	var createdAt time.Time
	err := row.Scan(&id, &filename, &size, &hash, &txId, &isUploaded, &createdAt)
	if err != nil {
		return model.File{}, err
	}

	return model.File{
		ID:         id,
		Filename:   filename,
		Size:       size,
		Hash:       hash,
		TxId:       txId.String,
		IsUploaded: isUploaded,
		CreatedAt:  createdAt,
	}, nil
}

func (d *DBService) ListFiles(ctx context.Context) ([]model.File, error) {
	query := `
		SELECT id, filename, size, hash, tx_id, is_uploaded, created_at
		FROM files
		ORDER BY created_at DESC
	`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []model.File{}
	for rows.Next() {
		var id int64
		var filename string
		var size int64
		var hash string
		var txId sql.NullString
		var isUploaded bool
		var createdAt time.Time
		err := rows.Scan(&id, &filename, &size, &hash, &txId, &isUploaded, &createdAt)
		if err != nil {
			return nil, err
		}

		file := model.File{
			ID:         id,
			Filename:   filename,
			Size:       size,
			Hash:       hash,
			IsUploaded: isUploaded,
			CreatedAt:  createdAt,
			TxId:       txId.String,
		}
		file.SetSizeReadable()
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (d *DBService) UpdateTxId(ctx context.Context, fileId int64, txId string) error {
	query := `
		UPDATE files
		SET tx_id = ?
		WHERE id = ?
	`
	_, err := d.db.ExecContext(ctx, query, txId, fileId)
	if err != nil {
		return err
	}

	return nil
}

func (d *DBService) GetUnuploadedFiles(ctx context.Context) ([]model.File, error) {
	query := `
		SELECT id, filename, hash, size, tx_id
		FROM files
		WHERE is_uploaded = FALSE AND tx_id IS NOT NULL
	`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []model.File{}
	for rows.Next() {
		var id int64
		var filename string
		var size int64
		var txId string
		var hash string
		err := rows.Scan(&id, &filename, &hash, &size, &txId)
		if err != nil {
			return nil, err
		}

		file := model.File{
			ID:       id,
			Filename: filename,
			Size:     size,
			TxId:     txId,
			Hash:     hash,
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (d *DBService) NewDownloadedFile(ctx context.Context, file model.File) error {
	query := `
		INSERT INTO downloaded_files (file_id, filename, hash, size)
		VALUES (?, ?, ?, ?)
	`
	_, err := d.db.ExecContext(ctx, query, file.ID, file.Filename, file.Hash, file.Size)
	if err != nil {
		return err
	}
	return nil
}

func (d *DBService) GetExpiredDownloadedFiles(ctx context.Context, duration time.Duration) ([]model.File, error) {
	query := `
		SELECT id, filename, hash, size
		FROM downloaded_files
		WHERE is_removed = FALSE AND is_processing = FALSE AND downloaded_at < ?
	`
	rows, err := d.db.QueryContext(ctx, query, time.Now().Add(-duration))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []model.File{}
	for rows.Next() {
		var id int64
		var filename string
		var size int64
		var hash string
		err := rows.Scan(&id, &filename, &hash, &size)
		if err != nil {
			return nil, err
		}

		file := model.File{
			ID:       id,
			Filename: filename,
			Size:     size,
			Hash:     hash,
		}
		file.SetSizeReadable()
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (d *DBService) SetProcessing(ctx context.Context, fileId int64) error {
	query := `
		UPDATE downloaded_files
		SET is_processing = FALSE
		WHERE file_id = ?
	`
	_, err := d.db.ExecContext(ctx, query, fileId)
	if err != nil {
		return err
	}
	return nil
}

func (d *DBService) ListDownloadedFiles(ctx context.Context) ([]model.DownloadedFile, error) {
	query := `
		SELECT id, file_id, filename, hash, size, is_processing, downloaded_at
		FROM downloaded_files
		WHERE is_removed = FALSE AND is_processing = FALSE
	`
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []model.DownloadedFile{}
	for rows.Next() {
		var id int64
		var fileId int64
		var filename string
		var size int64
		var hash string
		var downloadedAt time.Time
		var isDownloading bool
		err := rows.Scan(&id, &fileId, &filename, &hash, &size, &isDownloading, &downloadedAt)
		if err != nil {
			return nil, err
		}

		file := model.DownloadedFile{
			ID:            id,
			FileId:        fileId,
			Filename:      filename,
			Size:          size,
			IsDownloading: isDownloading,
			DownloadedAt:  downloadedAt,
		}
		file.SetSizeReadable()
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func (d *DBService) RemoveDownloadedFile(ctx context.Context, fileId int64) error {
	query := `
		UPDATE downloaded_files
		SET is_removed = TRUE
		WHERE id = ?
	`
	_, err := d.db.ExecContext(ctx, query, fileId)
	if err != nil {
		return err
	}
	return nil
}

func (d *DBService) CheckIsFileAlreadyDownloaded(ctx context.Context, hash string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM downloaded_files
		WHERE hash = ? AND is_removed = FALSE
	`
	row := d.db.QueryRowContext(ctx, query, hash)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *DBService) CheckDownloadStatus(ctx context.Context, hash string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM downloaded_files
		WHERE hash = ? AND is_removed = FALSE AND is_processing = FALSE
	`
	row := d.db.QueryRowContext(ctx, query, hash)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *DBService) GetHash(ctx context.Context, filename string) (string, error) {
	query := `
		SELECT hash
		FROM files
		WHERE filename = ?
	`
	row := d.db.QueryRowContext(ctx, query, filename)
	var hash string
	err := row.Scan(&hash)
	if err != nil {
		return "", err
	}
	return hash, nil
}
