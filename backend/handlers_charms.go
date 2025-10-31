package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	// "net/http"
)

type charmHandler struct{ db *sql.DB }

type upsertCharmReq struct {
	Shape     string `json:"shape"`
	Color     string `json:"color"`
	Title     string `json:"title"`
	TextValue string `json:"text_value"`
}

var allowedShapes = map[string]bool{"square": true, "star": true, "circle": true, "triangle": true, "rectangle": true, "diamond": true, "heart": true, "clover": true, "spade": true, "hexagon": true, "squiggle": true}
var allowedColors = map[string]bool{"red": true, "green": true, "blue": true, "yellow": true, "purple": true, "pink": true, "gold": true, "black": true, "orange": true, "darkgray": true}

func (h *charmHandler) listByBook(c *gin.Context) {
	uid := mustUserID(c)
	pid := c.Param("id")
	// owner check
	var owner string
	var isPub bool
	if err := h.db.QueryRow(`SELECT owner_id, is_public FROM books WHERE id=?`, pid).Scan(&owner, &isPub); err != nil {
		c.JSON(404, gin.H{"error": "book"})
		return
	}
	if owner != uid {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}
	rows, err := h.db.Query(`SELECT id,book_id,shape,color,title,text_value,created_at,updated_at FROM charms WHERE book_id=? ORDER BY updated_at DESC`, pid)
	if err != nil {
		c.JSON(500, gin.H{"error": "db"})
		return
	}
	defer rows.Close()
	var out []Charm
	for rows.Next() {
		var ch Charm
		if err := rows.Scan(&ch.ID, &ch.BookID, &ch.Shape, &ch.Color, &ch.Title, &ch.TextValue, &ch.CreatedAt, &ch.UpdatedAt); err == nil {
			out = append(out, ch)
		}
	}
	c.JSON(200, out)
}

func (h *charmHandler) create(c *gin.Context) {
	uid := mustUserID(c)
	pid := c.Param("id")
	var owner string
	if err := h.db.QueryRow(`SELECT owner_id FROM books WHERE id=?`, pid).Scan(&owner); err != nil {
		c.JSON(404, gin.H{"error": "book"})
		return
	}
	if owner != uid {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}
	var req upsertCharmReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad json"})
		return
	}
	if !allowedShapes[req.Shape] || !allowedColors[req.Color] || len(req.TextValue) > 256 {
		c.JSON(400, gin.H{"error": "invalid fields"})
		return
	}
	id := uuid.NewString()
	_, err := h.db.Exec(`INSERT INTO charms (id,book_id,shape,color,title,text_value) VALUES (?,?,?,?,?,?)`, id, pid, req.Shape, req.Color, req.Title, req.TextValue)
	if err != nil {
		c.JSON(500, gin.H{"error": "db"})
		return
	}
	c.JSON(200, gin.H{"id": id})
}

func (h *charmHandler) update(c *gin.Context) {
	uid := mustUserID(c)
	id := c.Param("id")
	var pid, owner string
	if err := h.db.QueryRow(`SELECT c.book_id, p.owner_id FROM charms c JOIN books p ON p.id=c.book_id WHERE c.id=?`, id).Scan(&pid, &owner); err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if owner != uid {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}
	var req upsertCharmReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad json"})
		return
	}
	if req.Shape != "" && !allowedShapes[req.Shape] {
		c.JSON(400, gin.H{"error": "shape"})
		return
	}
	if req.Color != "" && !allowedColors[req.Color] {
		c.JSON(400, gin.H{"error": "color"})
		return
	}
	_, err := h.db.Exec(`UPDATE charms SET shape=COALESCE(NULLIF(?,''),shape), color=COALESCE(NULLIF(?,''),color), title=COALESCE(NULLIF(?,''),title), text_value=CASE WHEN ?='' THEN text_value ELSE ? END WHERE id=?`, req.Shape, req.Color, req.Title, req.TextValue, req.TextValue, id)
	if err != nil {
		c.JSON(500, gin.H{"error": "db"})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h *charmHandler) delete(c *gin.Context) {
	uid := mustUserID(c)
	id := c.Param("id")
	var owner string
	if err := h.db.QueryRow(`SELECT p.owner_id FROM charms c JOIN books p ON p.id=c.book_id WHERE c.id=?`, id).Scan(&owner); err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if owner != uid {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}
	_, err := h.db.Exec(`DELETE FROM charms WHERE id=?`, id)
	if err != nil {
		c.JSON(500, gin.H{"error": "db"})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
