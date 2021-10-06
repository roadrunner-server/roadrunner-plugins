package send

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
)

const xSendHeader string = "X-Sendfile"

// PluginName contains default service name.
const (
	RootPluginName string = "http"
	PluginName     string = "sendfile"
)

type Plugin struct{}

func (s *Plugin) Init(cfg config.Configurer) error {
	const op = errors.Op("sendfile_plugin_init")
	if !cfg.Has(RootPluginName) {
		return errors.E(op, errors.Disabled)
	}

	return nil
}

// Middleware is HTTP plugin middleware to serve headers
func (s *Plugin) Middleware(next http.Handler) http.Handler {
	// Define the http.HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path := r.Header.Get(xSendHeader); path != "" {
			defer func() {
				_ = r.Body.Close()
			}()

			// do not allow paths like ../../resource, security
			// only specified folder and resources in it
			// see: https://lgtm.com/rules/1510366186013/
			if strings.Contains(path, "..") {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// check if the file exists
			_, err := os.Stat(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			data, err := os.ReadFile(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			buf := bytes.NewReader(data)

			_, err = io.Copy(w, buf)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Header.Del(xSendHeader)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Plugin) Name() string {
	return PluginName
}

// Available interface implementation
func (s *Plugin) Available() {}
