package main

import (
	"time"
)

type User struct {
	ID        string
	Email     string
	PassHash  []byte
	CreatedAt time.Time
}

type Book struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Title     string    `json:"title"`
	Note      string    `json:"note"`
	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Charm struct {
	ID        string    `json:"id"`
	BookID    string    `json:"book_id"`
	Shape     string    `json:"shape"`
	Color     string    `json:"color"`
	Title     string    `json:"title"`
	TextValue string    `json:"text_value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
