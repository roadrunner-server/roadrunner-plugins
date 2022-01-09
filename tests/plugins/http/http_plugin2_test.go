package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/goccy/go-json"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/fileserver"
	httpPlugin "github.com/spiral/roadrunner-plugins/v2/http"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/cache"
	newrelic "github.com/spiral/roadrunner-plugins/v2/http/middleware/new_relic"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/memory"
	rpcPlugin "github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPPost(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-post-test.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 1)
	t.Run("BombardWithPosts", echoHTTPPost)

	stopCh <- struct{}{}

	wg.Wait()
}

func echoHTTPPost(t *testing.T) {
	body := struct {
		Name  string `json:"name"`
		Index int    `json:"index"`
	}{
		Name:  "foo",
		Index: 111,
	}

	bd, err := json.Marshal(body)
	require.NoError(t, err)

	rdr := bytes.NewReader(bd)

	resp, err := http.Post("http://127.0.0.1:10084/", "", rdr)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	require.True(t, bytes.Equal(bd, b))

	_ = resp.Body.Close()

	for i := 0; i < 20; i++ {
		rdr = bytes.NewReader(bd)
		resp, err = http.Post("http://127.0.0.1:10084/", "application/json", rdr)
		assert.NoError(t, err)

		b, err = ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)

		require.True(t, bytes.Equal(bd, b))

		_ = resp.Body.Close()
	}
}

func TestSSLNoHTTP(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-ssl-no-http.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 1)
	t.Run("SSLEcho", sslEcho2)

	stopCh <- struct{}{}
	wg.Wait()
	time.Sleep(time.Second)
}

func sslEcho2(t *testing.T) {
	req, err := http.NewRequest("GET", "https://127.0.0.1:4455?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
}

func TestFileServer(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*30))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-http-static-new.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&fileserver.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second)
	t.Run("ServeSampleEtag", serveStaticSampleEtag2)

	stopCh <- struct{}{}
	wg.Wait()
}

func serveStaticSampleEtag2(t *testing.T) {
	// OK 200 response
	b, r, err := get("http://127.0.0.1:10101/foo/sample.txt")
	assert.NoError(t, err)
	assert.Contains(t, b, "sample")
	assert.Equal(t, r.StatusCode, http.StatusOK)
	etag := r.Header.Get("Etag")
	_ = r.Body.Close()

	// Should be 304 response with same etag
	c := http.Client{
		Timeout: time.Second * 5,
	}

	parsedURL, _ := url.Parse("http://127.0.0.1:10101/foo/sample.txt")

	req := &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
		Header: map[string][]string{"If-None-Match": {etag}},
	}

	resp, err := c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotModified, resp.StatusCode)
	_ = resp.Body.Close()
}

func TestHTTPNewRelic(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-http-new-relic.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&newrelic.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 1)
	req, err := http.NewRequest("GET", "http://127.0.0.1:19999", nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Nil(t, resp.Header["Rr_newrelic"])
	require.Equal(t, []string{"application/json"}, resp.Header["Content-Type"])

	stopCh <- struct{}{}

	wg.Wait()
}

func TestHTTPCache(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-http-cache.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&cache.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&memory.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestHTTPCacheDifferentRqs(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-http-cache.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&cache.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&memory.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:53123", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	time.Sleep(time.Second * 2)

	r, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	require.Greater(t, r.Header["Age"][0], "1")

	// -------------------

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:53123", nil)
	require.NoError(t, err)
	// typo
	req.Header.Set("Cache-Control", "max-age=abc")

	r, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, err)
	require.Equal(t, 200, r.StatusCode)

	// -----------------

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:53123", nil)
	require.NoError(t, err)
	// typo
	req.Header.Set("Cache-Control", "max-age=0")

	r, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, err)
	require.Equal(t, 200, r.StatusCode)

	r, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, err)
	require.Equal(t, 200, r.StatusCode)

	switch r.Header["Age"][0] {
	case "0":
	case "1":
	default:
		require.FailNow(t, "should be 0 or 1")
	}

	// -----------------

	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1:53123", nil)
	require.NoError(t, err)
	// typo
	req.Header.Set("Cache-Control", "max-age=10,public,foo,bar")

	r, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, err)
	require.Equal(t, 200, r.StatusCode)

	stopCh <- struct{}{}
	wg.Wait()
}
