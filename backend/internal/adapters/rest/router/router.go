// Package router exposes the HTTP composition boundary.
package router

import "github.com/go-chi/chi/v5"

// NewRouter creates the Chi router to which the application composition root
// attaches public, account, administration, and webhook route groups.
func NewRouter() *chi.Mux {
	return chi.NewRouter()
}
