package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authHandler struct {
	db     *sql.DB
	jwt    *jwtMgr
	secure bool
}

type registerReq struct{ Email, Passcode string }
type loginReq struct{ Email, Passcode string }

func (h *authHandler) register(c *gin.Context) {
	var req registerReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Passcode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and passcode required"})
		return
	}
	ph, _ := bcrypt.GenerateFromPassword([]byte(req.Passcode), bcrypt.DefaultCost)
	id := uuid.NewString()
	_, err := h.db.Exec(`INSERT INTO users (id,email,pass_hash) VALUES (?,?,?)`, id, email, ph)
	if isDup(err) {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	tok, _ := h.jwt.issue(id, 24*time.Hour*30)
	setAuthCookie(c, tok, h.secure)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *authHandler) login(c *gin.Context) {
	var req loginReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	row := h.db.QueryRow(`SELECT id, pass_hash FROM users WHERE email=?`, strings.TrimSpace(strings.ToLower(req.Email)))
	var id string
	var ph []byte
	if err := row.Scan(&id, &ph); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword(ph, []byte(req.Passcode)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	tok, _ := h.jwt.issue(id, 24*time.Hour*30)
	setAuthCookie(c, tok, h.secure)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *authHandler) logout(c *gin.Context) {
	clearAuthCookie(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func isDup(err error) bool { return err != nil && strings.Contains(err.Error(), "Duplicate entry") }

// --- minimal JWT-like cookie manager (HMAC signed compact payload) ---

type jwtMgr struct{ secret []byte }

type jwtClaims struct {
	UserID string    `json:"uid"`
	Exp    time.Time `json:"exp"`
}

func (j *jwtMgr) issue(uid string, ttl time.Duration) (string, error) {
	cl := jwtClaims{UserID: uid, Exp: time.Now().Add(ttl)}
	return signCompact(cl, j.secret), nil
}

func (j *jwtMgr) parse(tok string) (*jwtClaims, error) {
	cl, err := parseCompact(tok, j.secret)
	if err != nil {
		return nil, err
	}
	if time.Now().After(cl.Exp) {
		return nil, errors.New("expired")
	}
	return cl, nil
}

func setAuthCookie(c *gin.Context, tok string, secure bool) {
	// name, value, maxAge, path, domain, secure, httpOnly
	c.SetCookie("auth", tok, 60*60*24*30, "/", "", secure, true)
}
// Logout
func clearAuthCookie(c *gin.Context) {
	c.SetCookie("auth", "", -1, "/", "", true, true)
}
