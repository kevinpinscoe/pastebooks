package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type pageHandler struct{ db *sql.DB }

type upsertPageReq struct {
	Title    string `json:"title"`
	Note     string `json:"note"`
	IsPublic bool   `json:"is_public"`
}

func (h *pageHandler) listMine(c *gin.Context) {
	uid := mustUserID(c)
	rows, err := h.db.Query(`
		SELECT id, owner_id, title, note, is_public, created_at, updated_at
		FROM pages
		WHERE owner_id=?
		ORDER BY updated_at DESC`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	defer rows.Close()
	var out []Page
	for rows.Next() {
		var p Page
		if err := rows.Scan(&p.ID, &p.OwnerID, &p.Title, &p.Note, &p.IsPublic, &p.CreatedAt, &p.UpdatedAt); err == nil {
			out = append(out, p)
		}
	}
	c.JSON(http.StatusOK, out)
}

func (h *pageHandler) getMine(c *gin.Context) {
	uid := mustUserID(c)
	id := c.Param("id")
	var p Page
	err := h.db.QueryRow(`
		SELECT id, owner_id, title, note, is_public, created_at, updated_at
		FROM pages
		WHERE id=? AND owner_id=?`, id, uid).
		Scan(&p.ID, &p.OwnerID, &p.Title, &p.Note, &p.IsPublic, &p.CreatedAt, &p.UpdatedAt)
	switch err {
	case sql.ErrNoRows:
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	case nil:
		c.JSON(http.StatusOK, p)
		return
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
}

func (h *pageHandler) create(c *gin.Context) {
	uid := mustUserID(c)
	var req upsertPageReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	id := uuid.NewString()
	_, err := h.db.Exec(`
		INSERT INTO pages (id, owner_id, title, note, is_public)
		VALUES (?,?,?,?,?)`, id, uid, req.Title, req.Note, req.IsPublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *pageHandler) update(c *gin.Context) {
	uid := mustUserID(c)
	id := c.Param("id")
	var req upsertPageReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	res, err := h.db.Exec(`
		UPDATE pages SET title=?, note=?, is_public=?
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

func (h *pageHandler) delete(c *gin.Context) {
	uid := mustUserID(c)
	id := c.Param("id")
	res, err := h.db.Exec(`DELETE FROM pages WHERE id=? AND owner_id=?`, id, uid)
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

func (h *pageHandler) getPublic(c *gin.Context) {
	id := c.Param("id")
	var p Page
	err := h.db.QueryRow(`
		SELECT id, owner_id, title, note, is_public, created_at, updated_at
		FROM pages
		WHERE id=?`, id).
		Scan(&p.ID, &p.OwnerID, &p.Title, &p.Note, &p.IsPublic, &p.CreatedAt, &p.UpdatedAt)
	switch err {
	case sql.ErrNoRows:
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	case nil:
		if !p.IsPublic {
			c.JSON(http.StatusForbidden, gin.H{"error": "private"})
			return
		}
		c.JSON(http.StatusOK, p)
		return
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
}
