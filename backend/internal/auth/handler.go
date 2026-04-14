package authentication

import (
	"database/sql"
	"html/template"
)

// The struct is shared across the files in this package
type Handler struct {
	DB  *sql.DB            // Database connection
	Tpl *template.Template // HTML templates
}

// Optional: A constructor to make initialization cleaner
func NewHandler(db *sql.DB, tpl *template.Template) *Handler {
	return &Handler{
		DB:  db,
		Tpl: tpl,
	}
}
