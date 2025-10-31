package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type bookHandler struct{ db *sql.DB }

type upsertBookReq struct {
	Title    string `json:"title"`
	Note     string `json:"note"`
	IsPublic bool   `json:"is_public"`
}

func (h *bookHandler) listMine(c *gin.Context) {
	uid := mustUserID(c)
	rows, err := h.db.Query(`
		SELECT id, owner_id, title, note, is_public, created_at, updated_at
		FROM books
		WHERE owner_id=?
		ORDER BY updated_at DESC`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	defer rows.Close()
	var out []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.OwnerID, &b.Title, &b.Note, &b.IsPublic, &b.CreatedAt, &b.UpdatedAt); err == nil {
			out = append(out, b)
		}
	}
	c.JSON(http.StatusOK, out)
}

func (h *bookHandler) getMine(c *gin.Context) {
	uid := mustUserID(c)
	id := c.Param("id")
	var b Book
	err := h.db.QueryRow(`
		SELECT id, owner_id, title, note, is_public, created_at, updated_at
		FROM books
		WHERE id=? AND owner_id=?`, id, uid).
		Scan(&b.ID, &b.OwnerID, &b.Title, &b.Note, &b.IsPublic, &b.CreatedAt, &b.UpdatedAt)
	switch err {
	case sql.ErrNoRows:
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	case nil:
		c.JSON(http.StatusOK, b)
		return
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
}

func (h *bookHandler) create(c *gin.Context) {
	uid := mustUserID(c)
	var req upsertBookReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	id := uuid.NewString()
	_, err := h.db.Exec(`
		INSERT INTO books (id, owner_id, title, note, is_public)
		VALUES (?,?,?,?,?)`, id, uid, req.Title, req.Note, req.IsPublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *bookHandler) update(c *gin.Context) {
	uid := mustUserID(c)
	id := c.Param("id")
	var req upsertBookReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	res, err := h.db.Exec(`
		UPDATE books SET title=?, note=?, is_public=?
		WHERE id=? AND owner_id=?`, req.Title, req.Note, req.IsPublic, id, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *bookHandler) delete(c *gin.Context) {
	uid := mustUserID(c)
	id := c.Param("id")
	res, err := h.db.Exec(`DELETE FROM books WHERE id=? AND owner_id=?`, id, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *bookHandler) getPublic(c *gin.Context) {
	id := c.Param("id")
	var b Book
	err := h.db.QueryRow(`
		SELECT id, owner_id, title, note, is_public, created_at, updated_at
		FROM books
		WHERE id=?`, id).
		Scan(&b.ID, &b.OwnerID, &b.Title, &b.Note, &b.IsPublic, &b.CreatedAt, &b.UpdatedAt)
	switch err {
	case sql.ErrNoRows:
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	case nil:
		if !b.IsPublic {
			c.JSON(http.StatusForbidden, gin.H{"error": "private"})
			return
		}
		c.JSON(http.StatusOK, b)
		return
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
}
