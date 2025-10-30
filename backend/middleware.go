package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ctxKey string

const ctxUserID ctxKey = "uid"

func authMiddleware(jwt *jwtMgr) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok, err := c.Cookie("auth")
		if err != nil || tok == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
			return
		}
		cl, err := jwt.parse(tok)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(string(ctxUserID), cl.UserID)
		c.Next()
	}
}

func devAuthMiddleware(devUID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(string(ctxUserID), devUID)
		c.Next()
	}
}

func mustUserID(c *gin.Context) string {
	v, _ := c.Get(string(ctxUserID))
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
