package handler

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/payload"
	"go.uber.org/zap"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
	contentNone      = iota + 900
	contentStream
	contentMultipart
	contentFormData
)

// Request maps net/http requests to PSR7 compatible structure and managed state of temporary uploaded files.
type Request struct {
	// RemoteAddr contains ip address of client, make sure to check X-Real-Ip and X-Forwarded-For for real client address.
	RemoteAddr string `json:"remoteAddr"`

	// Protocol includes HTTP protocol version.
	Protocol string `json:"protocol"`

	// Method contains name of HTTP method used for the request.
	Method string `json:"method"`

	// URI contains full request URI with scheme and query.
	URI string `json:"uri"`

	// Header contains list of request headers.
	Header http.Header `json:"headers"`

	// Cookies contains list of request cookies.
	Cookies map[string]string `json:"cookies"`

	// RawQuery contains non parsed query string (to be parsed on php end).
	RawQuery string `json:"rawQuery"`

	// Parsed indicates that request body has been parsed on RR end.
	Parsed bool `json:"parsed"`

	// Uploads contains list of uploaded files, their names, sized and associations with temporary files.
	Uploads *Uploads `json:"uploads"`

	// Attributes can be set by chained mdwr to safely pass value from Golang to PHP. See: GetAttribute, SetAttribute functions.
	Attributes map[string]interface{} `json:"attributes"`

	// request body can be parsedData or []byte
	body interface{}
}

func FetchIP(pair string) string {
	if !strings.ContainsRune(pair, ':') {
		return pair
	}

	addr, _, _ := net.SplitHostPort(pair)
	return addr
}

func request(r *http.Request, req *Request) error {
	for _, c := range r.Cookies() {
		if v, err := url.QueryUnescape(c.Value); err == nil {
			req.Cookies[c.Name] = v
		}
	}

	switch req.contentType() {
	case contentNone:
		return nil

	case contentStream:
		var err error
		req.body, err = io.ReadAll(r.Body)
		return err

	case contentMultipart:
		if err := r.ParseMultipartForm(defaultMaxMemory); err != nil {
			return err
		}

		req.Uploads = parseUploads(r)
		fallthrough
	case contentFormData:
		if err := r.ParseForm(); err != nil {
			return err
		}

		req.body = parseData(r)
	}

	req.Parsed = true
	return nil
}

// Open moves all uploaded files to temporary directory so it can be given to php later.
func (r *Request) Open(log *zap.Logger, dir string, forbid, allow map[string]struct{}) {
	if r.Uploads == nil {
		return
	}

	r.Uploads.Open(log, dir, forbid, allow)
}

// Close clears all temp file uploads
func (r *Request) Close(log *zap.Logger) {
	if r.Uploads == nil {
		return
	}

	r.Uploads.Clear(log)
}

// Payload request marshaled RoadRunner payload based on PSR7 data. values encode method is JSON. Make sure to open
// files prior to calling this method.
func (r *Request) Payload(p *payload.Payload) error {
	const op = errors.Op("marshal_payload")

	var err error
	p.Context, err = json.Marshal(r)
	if err != nil {
		return err
	}

	// check if body was already parsed
	if r.Parsed {
		p.Body, err = json.Marshal(r.body)
		if err != nil {
			return errors.E(op, errors.Encode, err)
		}

		return nil
	}

	if r.body != nil {
		p.Body = r.body.([]byte)
	}

	return nil
}

// contentType returns the payload content type.
func (r *Request) contentType() int {
	if r.Method == "HEAD" || r.Method == "OPTIONS" {
		return contentNone
	}

	ct := r.Header.Get("content-type")
	if strings.Contains(ct, "application/x-www-form-urlencoded") {
		return contentFormData
	}

	if strings.Contains(ct, "multipart/form-data") {
		return contentMultipart
	}

	return contentStream
}

// URI fetches full uri from request in a form of string (including https scheme if TLS connection is enabled).
func URI(r *http.Request) string {
	// CVE: https://github.com/spiral/roadrunner-plugins/pull/184/checks?check_run_id=4635904339
	uri := r.URL.String()
	uri = strings.Replace(uri, "\n", "", -1) //nolint:gocritic
	uri = strings.Replace(uri, "\r", "", -1) //nolint:gocritic

	if r.URL.Host != "" {
		return uri
	}

	if r.TLS != nil {
		return fmt.Sprintf("https://%s%s", r.Host, uri)
	}

	return fmt.Sprintf("http://%s%s", r.Host, uri)
}
