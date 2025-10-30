package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	configPath := flag.String("config", "./config.yaml", "path to config")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := openDB(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	devUID := "dev-user"
	if cfg.AuthDisabled {
		log.Printf("[auth] DISABLED (dev mode) using uid=%s", devUID)
		_, _ = db.Exec(`
		INSERT INTO users (id, email, pass_hash)
		VALUES (?, ?, '')
		ON DUPLICATE KEY UPDATE email = email
	`, devUID, "dev@local")
	}

	jwt := &jwtMgr{secret: []byte(cfg.JWTSecret)}
	ah := &authHandler{db: db, jwt: jwt, secure: cfg.CookieSecure}
	ph := &pageHandler{db: db}
	ch := &charmHandler{db: db}

	r := gin.Default()

	// Serve frontend safely (no wildcard at "/")
	r.Static("/static", "./frontend") // /static/app.js, /static/styles.css, etc.
	r.GET("/", func(c *gin.Context) { // index route
		c.File("./frontend/index.html")
	})
	// Optional: SPA-style fallback (but don't swallow /api/*)
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 5 && c.Request.URL.Path[:5] == "/api/" {
			c.Status(http.StatusNotFound)
			return
		}
		c.File("./frontend/index.html")
	})

	api := r.Group("/api")
	{
		api.POST("/register", ah.register)
		api.POST("/login", ah.login)
		api.POST("/logout", ah.logout)
		api.GET("/me", func(c *gin.Context) {
			if cfg.AuthDisabled {
				c.JSON(200, gin.H{"user_id": "dev-user", "dev": true})
				return
			}
			if tok, err := c.Cookie("auth"); err == nil {
				if cl, err := jwt.parse(tok); err == nil {
					c.JSON(200, gin.H{"user_id": cl.UserID})
					return
				}
			}
			c.JSON(200, gin.H{"user_id": ""})
		})
		api.GET("/public/pages/:id", ph.getPublic)
	}

	auth := r.Group("/api")
	if cfg.AuthDisabled {
		auth.Use(devAuthMiddleware(devUID))
	} else {
		auth.Use(authMiddleware(jwt))
	}
	{
		auth.GET("/pages", ph.listMine)
		auth.POST("/pages", ph.create)
		auth.GET("/pages/:id", ph.getMine)
		auth.PUT("/pages/:id", ph.update)
		auth.DELETE("/pages/:id", ph.delete)

		auth.GET("/pages/:id/charms", ch.listByPage)
		auth.POST("/pages/:id/charms", ch.create)
		auth.PUT("/charms/:id", ch.update)
		auth.DELETE("/charms/:id", ch.delete)
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("listening on %s", addr)
	log.Fatal((&http.Server{Addr: addr, Handler: r}).ListenAndServe())
}
