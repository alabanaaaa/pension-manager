package documents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"pension-manager/internal/db"
)

// Document represents a stored document
type Document struct {
	ID           string    `json:"id"`
	EntityType   string    `json:"entity_type"`
	EntityID     string    `json:"entity_id"`
	SchemeID     string    `json:"scheme_id"`
	DocumentType string    `json:"document_type"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	StoragePath  string    `json:"storage_path"`
	UploadedBy   string    `json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}

// Storage is the interface for document storage (S3/MinIO/local)
type Storage interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	URL(ctx context.Context, key string) (string, error)
}

// LocalStorage implements Storage using local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local file storage
func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

func (s *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	// For now, we'll store metadata in DB and return success
	// In production, this would write to disk
	return nil
}

func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, errors.New("local storage download not implemented")
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func (s *LocalStorage) URL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("/api/documents/download/%s", key), nil
}

// S3Storage implements Storage using AWS S3 or MinIO
type S3Storage struct {
	endpoint string
	bucket   string
	region   string
	// In production, would include S3 client
}

// NewS3Storage creates a new S3/MinIO storage
func NewS3Storage(endpoint, bucket, region string) *S3Storage {
	return &S3Storage{
		endpoint: endpoint,
		bucket:   bucket,
		region:   region,
	}
}

func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	// In production, would use AWS SDK to upload to S3/MinIO
	return nil
}

func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, errors.New("S3 storage download not implemented")
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	// In production, would delete from S3/MinIO
	return nil
}

func (s *S3Storage) URL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key), nil
}

// Service manages document operations
type Service struct {
	db      *db.DB
	storage Storage
}

// NewService creates a new document service
func NewService(db *db.DB, storage Storage) *Service {
	return &Service{
		db:      db,
		storage: storage,
	}
}

// UploadDocument stores a new document
func (s *Service) UploadDocument(ctx context.Context, doc *Document, file multipart.File, header *multipart.FileHeader) error {
	if doc == nil {
		return errors.New("document cannot be nil")
	}
	if file == nil || header == nil {
		return errors.New("file cannot be nil")
	}

	// Generate storage key
	ext := filepath.Ext(header.Filename)
	key := fmt.Sprintf("%s/%s/%s%s", doc.SchemeID, doc.EntityType, doc.ID, ext)

	// Upload to storage
	if err := s.storage.Upload(ctx, key, file, header.Header.Get("Content-Type")); err != nil {
		return fmt.Errorf("upload file: %w", err)
	}

	// Save metadata to database
	doc.StoragePath = key
	doc.FileName = header.Filename
	doc.FileSize = header.Size
	doc.MimeType = header.Header.Get("Content-Type")
	if doc.MimeType == "" {
		doc.MimeType = "application/octet-stream"
	}
	doc.CreatedAt = time.Now()

	query := `
		INSERT INTO documents (id, entity_type, entity_id, scheme_id, document_type,
		                       file_name, file_size, mime_type, storage_path, uploaded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := s.db.ExecContext(ctx, query,
		doc.ID, doc.EntityType, doc.EntityID, doc.SchemeID, doc.DocumentType,
		doc.FileName, doc.FileSize, doc.MimeType, doc.StoragePath, doc.UploadedBy, doc.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save document metadata: %w", err)
	}

	return nil
}

// GetDocument retrieves document metadata by ID
func (s *Service) GetDocument(ctx context.Context, documentID string) (*Document, error) {
	query := `
		SELECT id, entity_type, entity_id, scheme_id, document_type, file_name,
		       file_size, mime_type, storage_path, uploaded_by, created_at
		FROM documents WHERE id = $1
	`
	doc := &Document{}
	err := s.db.QueryRowContext(ctx, query, documentID).Scan(
		&doc.ID, &doc.EntityType, &doc.EntityID, &doc.SchemeID, &doc.DocumentType,
		&doc.FileName, &doc.FileSize, &doc.MimeType, &doc.StoragePath, &doc.UploadedBy, &doc.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	return doc, nil
}

// ListDocuments retrieves all documents for an entity
func (s *Service) ListDocuments(ctx context.Context, entityType, entityID string) ([]*Document, error) {
	query := `
		SELECT id, entity_type, entity_id, scheme_id, document_type, file_name,
		       file_size, mime_type, storage_path, uploaded_by, created_at
		FROM documents WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var docs []*Document
	for rows.Next() {
		doc := &Document{}
		if err := rows.Scan(
			&doc.ID, &doc.EntityType, &doc.EntityID, &doc.SchemeID, &doc.DocumentType,
			&doc.FileName, &doc.FileSize, &doc.MimeType, &doc.StoragePath, &doc.UploadedBy, &doc.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, doc)
	}
	return docs, rows.Err()
}

// DownloadDocument retrieves the document file
func (s *Service) DownloadDocument(ctx context.Context, documentID string) (io.ReadCloser, *Document, error) {
	doc, err := s.GetDocument(ctx, documentID)
	if err != nil {
		return nil, nil, err
	}
	if doc == nil {
		return nil, nil, errors.New("document not found")
	}

	reader, err := s.storage.Download(ctx, doc.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("download file: %w", err)
	}

	return reader, doc, nil
}

// DeleteDocument removes a document
func (s *Service) DeleteDocument(ctx context.Context, documentID string) error {
	doc, err := s.GetDocument(ctx, documentID)
	if err != nil {
		return err
	}
	if doc == nil {
		return errors.New("document not found")
	}

	// Delete from storage
	if err := s.storage.Delete(ctx, doc.StoragePath); err != nil {
		return fmt.Errorf("delete from storage: %w", err)
	}

	// Delete from database
	_, err = s.db.ExecContext(ctx, `DELETE FROM documents WHERE id = $1`, documentID)
	if err != nil {
		return fmt.Errorf("delete document metadata: %w", err)
	}

	return nil
}

// GetDocumentURL returns the download URL for a document
func (s *Service) GetDocumentURL(ctx context.Context, documentID string) (string, error) {
	doc, err := s.GetDocument(ctx, documentID)
	if err != nil {
		return "", err
	}
	if doc == nil {
		return "", errors.New("document not found")
	}

	return s.storage.URL(ctx, doc.StoragePath)
}
