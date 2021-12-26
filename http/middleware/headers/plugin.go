package headers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/api/v2/config"
)

// PluginName contains default service name.
const (
	RootPluginName string = "http"
	PluginName     string = "headers"
)

// Plugin serves headers files. Potentially convert into middleware?
type Plugin struct {
	// server configuration (location, forbidden files and etc)
	cfg *Config
}

// Init must return configure service and return true if service hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Plugin) Init(cfg config.Configurer) error {
	const op = errors.Op("headers_plugin_init")

	if !cfg.Has(RootPluginName) {
		return errors.E(op, errors.Disabled)
	}

	if !cfg.Has(fmt.Sprintf("%s.%s", RootPluginName, PluginName)) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(fmt.Sprintf("%s.%s", RootPluginName, PluginName), &s.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

// Middleware is HTTP plugin middleware to serve headers
func (s *Plugin) Middleware(next http.Handler) http.Handler {
	// Define the http.HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Request != nil {
			for k, v := range s.cfg.Request {
				r.Header.Add(k, v)
			}
		}

		if s.cfg.Response != nil {
			for k, v := range s.cfg.Response {
				w.Header().Set(k, v)
			}
		}

		if s.cfg.CORS != nil {
			if r.Method == http.MethodOptions {
				s.preflightRequest(w)
				return
			}
			s.corsHeaders(w)
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Plugin) Name() string {
	return PluginName
}

// configure OPTIONS response
func (s *Plugin) preflightRequest(w http.ResponseWriter) {
	headers := w.Header()

	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")

	if s.cfg.CORS.AllowedOrigin != "" {
		headers.Set("Access-Control-Allow-Origin", s.cfg.CORS.AllowedOrigin)
	}

	if s.cfg.CORS.AllowedHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", s.cfg.CORS.AllowedHeaders)
	}

	if s.cfg.CORS.AllowedMethods != "" {
		headers.Set("Access-Control-Allow-Methods", s.cfg.CORS.AllowedMethods)
	}

	if s.cfg.CORS.AllowCredentials != nil {
		headers.Set("Access-Control-Allow-Credentials", strconv.FormatBool(*s.cfg.CORS.AllowCredentials))
	}

	if s.cfg.CORS.MaxAge > 0 {
		headers.Set("Access-Control-Max-Age", strconv.Itoa(s.cfg.CORS.MaxAge))
	}

	w.WriteHeader(http.StatusOK)
}

// configure CORS headers
func (s *Plugin) corsHeaders(w http.ResponseWriter) {
	headers := w.Header()

	headers.Add("Vary", "Origin")

	if s.cfg.CORS.AllowedOrigin != "" {
		headers.Set("Access-Control-Allow-Origin", s.cfg.CORS.AllowedOrigin)
	}

	if s.cfg.CORS.AllowedHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", s.cfg.CORS.AllowedHeaders)
	}

	if s.cfg.CORS.ExposedHeaders != "" {
		headers.Set("Access-Control-Expose-Headers", s.cfg.CORS.ExposedHeaders)
	}

	if s.cfg.CORS.AllowCredentials != nil {
		headers.Set("Access-Control-Allow-Credentials", strconv.FormatBool(*s.cfg.CORS.AllowCredentials))
	}
}
