package api

import (
	"fmt"
	"net/http"
	"path/filepath"

	"pension-manager/internal/documents"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// registerDocumentRoutes registers document management routes
func (s *Server) registerDocumentRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(s.auth))

		r.Route("/documents", func(r chi.Router) {
			r.Post("/upload", s.handleUploadDocument)
			r.Get("/{id}", s.handleGetDocument)
			r.Get("/{id}/download", s.handleDownloadDocument)
			r.Get("/{id}/url", s.handleGetDocumentURL)
			r.Delete("/{id}", s.handleDeleteDocument)
			r.Get("/entity/{entityType}/{entityID}", s.handleListEntityDocuments)
		})
	})
}

// handleUploadDocument handles POST /documents/upload
func (s *Server) handleUploadDocument(w http.ResponseWriter, r *http.Request) {
	schemeID := GetSchemeID(r)
	userID := GetUserID(r)

	// Parse multipart form (max 50MB)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	// Validate file extension
	ext := filepath.Ext(header.Filename)
	allowedExts := map[string]bool{
		".pdf": true, ".jpg": true, ".jpeg": true, ".png": true,
		".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".csv": true, ".txt": true,
	}
	if !allowedExts[ext] {
		respondError(w, http.StatusBadRequest, "file type not allowed. Allowed: pdf, jpg, png, doc, docx, xls, xlsx, csv, txt")
		return
	}

	entityType := r.FormValue("entity_type")
	entityID := r.FormValue("entity_id")
	documentType := r.FormValue("document_type")

	if entityType == "" || entityID == "" || documentType == "" {
		respondError(w, http.StatusBadRequest, "entity_type, entity_id, and document_type are required")
		return
	}

	validTypes := map[string]bool{
		"death_certificate": true, "id": true, "marriage_cert": true,
		"claim_form": true, "kra_pin": true, "bank_statement": true,
		"letter_of_admin": true, "affidavit": true, "sponsor_clearance": true,
		"beneficiary_form": true, "other": true,
	}
	if !validTypes[documentType] {
		respondError(w, http.StatusBadRequest, "invalid document_type")
		return
	}

	doc := &documents.Document{
		ID:           uuid.New().String(),
		EntityType:   entityType,
		EntityID:     entityID,
		SchemeID:     schemeID,
		DocumentType: documentType,
		UploadedBy:   userID,
	}

	if err := s.docService.UploadDocument(r.Context(), doc, file, header); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to upload document: %v", err))
		return
	}

	respondJSON(w, http.StatusCreated, doc)
}

// handleGetDocument handles GET /documents/{id}
func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	documentID := chi.URLParam(r, "id")
	if documentID == "" {
		respondError(w, http.StatusBadRequest, "document ID is required")
		return
	}

	doc, err := s.docService.GetDocument(r.Context(), documentID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get document: %v", err))
		return
	}
	if doc == nil {
		respondError(w, http.StatusNotFound, "document not found")
		return
	}

	respondJSON(w, http.StatusOK, doc)
}

// handleDownloadDocument handles GET /documents/{id}/download
func (s *Server) handleDownloadDocument(w http.ResponseWriter, r *http.Request) {
	documentID := chi.URLParam(r, "id")
	if documentID == "" {
		respondError(w, http.StatusBadRequest, "document ID is required")
		return
	}

	reader, doc, err := s.docService.DownloadDocument(r.Context(), documentID)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("document not found: %v", err))
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", doc.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", doc.FileName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", doc.FileSize))

	// Stream the file to the response
	buf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}

// handleGetDocumentURL handles GET /documents/{id}/url
func (s *Server) handleGetDocumentURL(w http.ResponseWriter, r *http.Request) {
	documentID := chi.URLParam(r, "id")
	if documentID == "" {
		respondError(w, http.StatusBadRequest, "document ID is required")
		return
	}

	url, err := s.docService.GetDocumentURL(r.Context(), documentID)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("document not found: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"url": url})
}

// handleDeleteDocument handles DELETE /documents/{id}
func (s *Server) handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	documentID := chi.URLParam(r, "id")
	if documentID == "" {
		respondError(w, http.StatusBadRequest, "document ID is required")
		return
	}

	if err := s.docService.DeleteDocument(r.Context(), documentID); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete document: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// handleListEntityDocuments handles GET /documents/entity/{entityType}/{entityID}
func (s *Server) handleListEntityDocuments(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")
	entityID := chi.URLParam(r, "entityID")

	if entityType == "" || entityID == "" {
		respondError(w, http.StatusBadRequest, "entity_type and entity_id are required")
		return
	}

	docs, err := s.docService.ListDocuments(r.Context(), entityType, entityID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list documents: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, docs)
}
