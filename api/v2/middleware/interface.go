package middleware

import (
	"net/http"
)

// Middleware interface
type Middleware interface {
	Middleware(f http.Handler) http.Handler
}
