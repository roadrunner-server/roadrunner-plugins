package http

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/rpc"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/gzip"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/send"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/static"
	"github.com/spiral/roadrunner-plugins/v2/informer"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/resetter"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/tests/mocks"
	"github.com/spiral/roadrunner/v2/state/process"
	"github.com/stretchr/testify/require"
	"github.com/yookoala/gofast"

	httpPlugin "github.com/spiral/roadrunner-plugins/v2/http"
	rpcPlugin "github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/stretchr/testify/assert"
)

var sslClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		},
	},
}

func TestHTTPInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-init.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
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

	stopCh <- struct{}{}
	wg.Wait()
}

func TestHTTPAccessLogs(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-access-logs.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("http server is running", "address", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info(
		gomock.Any(),
		"status", gomock.Any(),
		"method", gomock.Any(),
		"URI", gomock.Any(),
		"remote_address", gomock.Any(),
		"query", gomock.Any(),
		"request_length", gomock.Any(),
		"bytes_sent", gomock.Any(),
		"host", gomock.Any(),
		"user_agent", "Go-http-client/1.1",
		"referer", gomock.Any(),
		"time_local", gomock.Any(),
		"request_time", gomock.Any(),
		"start", gomock.Any(),
		"elapsed", gomock.Any()).MinTimes(1)

	err = cont.RegisterAll(
		cfg,
		mockLogger,
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

	time.Sleep(time.Second * 2)

	t.Run("AccessLogsEcho", echoAccessLogs)

	stopCh <- struct{}{}
	wg.Wait()
}

func echoAccessLogs(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:58332", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NotNil(t, r)
	b, err := io.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "hello world", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func TestHTTPXSendFile(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-sendfile.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&send.Plugin{},
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

	time.Sleep(time.Second * 2)
	t.Run("X-Sendfile", xsendfile)
	stopCh <- struct{}{}
	wg.Wait()
}

func xsendfile(t *testing.T) {
	parsedURL, _ := url.Parse("http://127.0.0.1:44444")
	client := http.Client{}
	pwd, _ := os.Getwd()
	req := &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
		Header: map[string][]string{"x-sendfile": {fmt.Sprintf("%s/attributes_test.go", pwd)}},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	require.True(t, len(b) > 0)
	require.Empty(t, resp.Header.Get("X-Sendfile"))

	file, err := os.ReadFile(fmt.Sprintf("%s/attributes_test.go", pwd))
	require.NoError(t, err)
	require.Equal(t, file, b)
	require.NoError(t, resp.Body.Close())
	_, _ = io.Discard.Write(file)
}

func TestHTTPNoConfigSection(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-no-http.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
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

	time.Sleep(time.Second * 2)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestHTTPInformerReset(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-resetter.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&informer.Plugin{},
		&resetter.Plugin{},
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
	t.Run("HTTPInformerTest", informerTest("127.0.0.1:6008"))
	t.Run("HTTPEchoTestBefore", echoHTTP)
	t.Run("HTTPResetTest", resetTest("127.0.0.1:6008"))
	t.Run("HTTPEchoTestAfter", echoHTTP)

	stopCh <- struct{}{}

	wg.Wait()
}

func TestSSL(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-ssl.yaml",
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
	t.Run("SSLEcho", sslEcho)
	t.Run("SSLNoRedirect", sslNoRedirect)
	t.Run("FCGEcho", fcgiEcho)

	stopCh <- struct{}{}
	wg.Wait()
}

func sslNoRedirect(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:8085?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)

	assert.Nil(t, r.TLS)

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

func sslEcho(t *testing.T) {
	req, err := http.NewRequest("GET", "https://127.0.0.1:8893?hello=world", nil)
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

func fcgiEcho(t *testing.T) {
	fcgiConnFactory := gofast.SimpleConnFactory("tcp", "0.0.0.0:16920")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/?hello=world", nil)
	fcgiHandler.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Result().Body) //nolint:bodyclose

	defer func() {
		_ = w.Result().Body.Close()
		w.Body.Reset()
	}()

	assert.NoError(t, err)
	assert.Equal(t, 201, w.Result().StatusCode) //nolint:bodyclose
	assert.Equal(t, "WORLD", string(body))
}

func TestSSLRedirect(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-ssl-redirect.yaml",
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
	t.Run("SSLRedirect", sslRedirect)

	stopCh <- struct{}{}
	wg.Wait()
}

func sslRedirect(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:8087?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, r.TLS)

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

func TestSSLPushPipes(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-ssl-push.yaml",
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
	t.Run("SSLPush", sslPush)

	stopCh <- struct{}{}
	wg.Wait()
}

func sslPush(t *testing.T) {
	req, err := http.NewRequest("GET", "https://127.0.0.1:8894?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)

	assert.NotNil(t, r.TLS)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, "", r.Header.Get("Http2-Release"))

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
}

func TestFastCGI_Echo(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-fcgi.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&static.Plugin{},
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
	t.Run("FastCGIEcho", fcgiEcho1)

	stopCh <- struct{}{}
	wg.Wait()
}

func fcgiEcho1(t *testing.T) {
	time.Sleep(time.Second * 2)
	fcgiConnFactory := gofast.SimpleConnFactory("tcp", "127.0.0.1:6920")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/hello-world", nil)
	fcgiHandler.ServeHTTP(w, req)

	_, err := ioutil.ReadAll(w.Result().Body) //nolint:bodyclose
	assert.NoError(t, err)
	assert.Equal(t, 201, w.Result().StatusCode) //nolint:bodyclose
}

func TestFastCGI_RequestUri(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-fcgi-reqUri.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
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
	t.Run("FastCGIServiceRequestUri", fcgiReqURI)

	stopCh <- struct{}{}
	wg.Wait()
}

func fcgiReqURI(t *testing.T) {
	time.Sleep(time.Second * 2)
	fcgiConnFactory := gofast.SimpleConnFactory("tcp", "127.0.0.1:6921")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/hello-world", nil)
	fcgiHandler.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Result().Body) //nolint:bodyclose
	assert.NoError(t, err)
	assert.Equal(t, 200, w.Result().StatusCode) //nolint:bodyclose
	assert.Equal(t, "http://site.local/hello-world", string(body))
}

func TestH2CUpgrade(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-h2c.yaml",
		Prefix: "rr",
	}
	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error("server internal error", "message", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Debug("http server is running", "address", gomock.Any()).AnyTimes()

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		mockLogger,
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
	t.Run("H2cUpgrade", h2cUpgrade)

	stopCh <- struct{}{}
	wg.Wait()
}

func h2cUpgrade(t *testing.T) {
	req, err := http.NewRequest("PRI", "http://127.0.0.1:8083?hello=world", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Upgrade", "h2c")
	req.Header.Add("Connection", "HTTP2-Settings")
	req.Header.Add("HTTP2-Settings", "")

	r, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "101 Switching Protocols", r.Status)

	err3 := r.Body.Close()
	if err3 != nil {
		t.Fatal(err)
	}
}

func TestH2C(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-h2c.yaml",
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
	t.Run("H2c", h2c)

	stopCh <- struct{}{}
	wg.Wait()
}

func h2c(t *testing.T) {
	req, err := http.NewRequest("PRI", "http://127.0.0.1:8083?hello=world", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Connection", "HTTP2-Settings")
	req.Header.Add("HTTP2-Settings", "")

	r, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "201 Created", r.Status)

	err3 := r.Body.Close()
	if err3 != nil {
		t.Fatal(err)
	}
}

func TestHttpMiddleware(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&PluginMiddleware{},
		&PluginMiddleware2{},
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
	t.Run("MiddlewareTest", middleware)

	stopCh <- struct{}{}
	wg.Wait()
}

func middleware(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:18903?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)

	req, err = http.NewRequest("GET", "http://127.0.0.1:18903/halt", nil)
	assert.NoError(t, err)

	r, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err = ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 500, r.StatusCode)
	assert.Equal(t, "halted", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func TestHttpEchoErr(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	rIn := `
rpc:
  listen: tcp://127.0.0.1:6003
  disabled: false

server:
  command: "php ../../php_test_files/http/client.php echoerr pipes"
  relay: "pipes"
  relay_timeout: "20s"

http:
  debug: true
  address: 127.0.0.1:34999
  max_request_size: 1024
  middleware: [ "pluginMiddleware", "pluginMiddleware2" ]
  uploads:
    forbid: [ "" ]
  trusted_subnets: [ "10.0.0.0/8", "127.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "::1/128", "fc00::/7", "fe80::/10" ]
  pool:
    num_workers: 2
    max_jobs: 0
    allocate_timeout: 60s
    destroy_timeout: 60s

logs:
  mode: development
  level: debug
`

	cfg := &config.Viper{
		Type:      "yaml",
		ReadInCfg: []byte(rIn),
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&PluginMiddleware{},
		&PluginMiddleware2{},
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

	time.Sleep(time.Second * 3)

	t.Run("HttpEchoError", echoError)

	stopCh <- struct{}{}
	wg.Wait()
}

func echoError(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:34999?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	require.NotNil(t, r)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))
	err = r.Body.Close()
	assert.NoError(t, err)
}

func TestHttpEnvVariables(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-env.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&PluginMiddleware{},
		&PluginMiddleware2{},
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
	t.Run("EnvVariablesTest", envVarsTest)

	stopCh <- struct{}{}
	wg.Wait()
}

func envVarsTest(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:12084", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "ENV_VALUE", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func TestHttpBrokenPipes(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-broken-pipes.yaml",
		Prefix: "rr",
		Type:   "yaml",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&PluginMiddleware{},
		&PluginMiddleware2{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

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
			// should be error from the plugin
			case e := <-ch:
				assert.Error(t, e.Error)
				return
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

	wg.Wait()
}

func TestHTTPSupervisedPool(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-supervised-pool.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&informer.Plugin{},
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
	t.Run("HTTPEchoRunActivateWorker", echoHTTP2)
	// bigger timeout to handle idle_ttl on slow systems
	time.Sleep(time.Second * 10)
	t.Run("HTTPInformerCompareWorkersTestBefore", informerTestBefore)
	t.Run("HTTPEchoShouldBeNewWorker", echoHTTP2)
	// worker should be destructed (idle_ttl)
	t.Run("HTTPInformerCompareWorkersTestAfter", informerTestAfter)

	stopCh <- struct{}{}
	wg.Wait()
}

func echoHTTP2(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:18888?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

// get worker
// sleep
// supervisor destroy worker
// compare pid's
var workerPid int = 0

func informerTestBefore(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:15432")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	list := struct {
		// Workers is list of workers.
		Workers []process.State `json:"workers"`
	}{}

	err = client.Call("informer.Workers", "http", &list)
	assert.NoError(t, err)
	assert.Len(t, list.Workers, 1)
	// save the pid
	workerPid = list.Workers[0].Pid
}

func informerTestAfter(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:15432")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	list := struct {
		// Workers is list of workers.
		Workers []process.State `json:"workers"`
	}{}

	assert.NotZero(t, workerPid)

	time.Sleep(time.Second * 5)

	err = client.Call("informer.Workers", "http", &list)
	assert.NoError(t, err)
	assert.Len(t, list.Workers, 1)
	assert.NotEqual(t, workerPid, list.Workers[0].Pid)
}

// get request and return body
func getHeader(url string, h map[string]string) (string, *http.Response, error) {
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
	if err != nil {
		return "", nil, err
	}

	for k, v := range h {
		req.Header.Set(k, v)
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return "", nil, err
	}
	return string(b), r, err
}

func TestHTTPBigRequestSize(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-big-req-size.yaml",
		Prefix: "rr",
		Type:   "yaml",
	}

	err = cont.RegisterAll(
		cfg,
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

	time.Sleep(time.Second * 2)

	t.Run("HTTPBigEcho10Mb", bigEchoHTTP)

	stopCh <- struct{}{}
	wg.Wait()
}

func bigEchoHTTP(t *testing.T) {
	buf := make([]byte, 1024*1024*10)

	_, err := rand.Read(buf)
	assert.NoError(t, err)

	bt := bytes.NewBuffer(buf)

	req, err := http.NewRequest("GET", "http://127.0.0.1:10085?hello=world", bt)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 400, r.StatusCode)
	assert.Equal(t, "http_handler_max_size: request body max size is exceeded\n", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func TestStaticEtagPlugin(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-static.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
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
	t.Run("ServeSampleEtag", serveStaticSampleEtag)
	t.Run("NoStaticHeaders", noStaticHeaders)

	stopCh <- struct{}{}
	wg.Wait()
}

func serveStaticSampleEtag(t *testing.T) {
	// OK 200 response
	b, r, err := get("http://127.0.0.1:21603/sample.txt")
	assert.NoError(t, err)
	assert.Contains(t, b, "sample")
	assert.Equal(t, r.StatusCode, http.StatusOK)
	etag := r.Header.Get("Etag")

	_ = r.Body.Close()

	// Should be 304 response with same etag
	c := http.Client{
		Timeout: time.Second * 5,
	}

	parsedURL, _ := url.Parse("http://127.0.0.1:21603/sample.txt")

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

// regular request should not contain static headers
func noStaticHeaders(t *testing.T) {
	// OK 200 response
	_, r, err := get("http://127.0.0.1:21603")
	assert.NoError(t, err)
	assert.NotContains(t, r.Header["input"], "custom-header")
	assert.NotContains(t, r.Header["output"], "output-header")
	assert.Equal(t, r.StatusCode, http.StatusOK)

	_ = r.Body.Close()
}

func TestStaticPluginSecurity(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-static-security.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
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
	t.Run("ServeSampleNotAllowedPath", serveStaticSampleNotAllowedPath)

	stopCh <- struct{}{}
	wg.Wait()
}

func serveStaticSampleNotAllowedPath(t *testing.T) {
	// Should be 304 response with same etag
	c := http.Client{
		Timeout: time.Second * 5,
	}

	parsedURL := &url.URL{
		Scheme: "http",
		User:   nil,
		Host:   "127.0.0.1:21603",
		Path:   "%2e%2e%/tests/",
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
	}

	resp, err := c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()

	parsedURL = &url.URL{
		Scheme: "http",
		User:   nil,
		Host:   "127.0.0.1:21603",
		Path:   "%2e%2e%5ctests/",
	}

	req = &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
	}

	resp, err = c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()

	parsedURL = &url.URL{
		Scheme: "http",
		User:   nil,
		Host:   "127.0.0.1:21603",
		Path:   "..%2ftests/",
	}

	req = &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
	}

	resp, err = c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()

	parsedURL = &url.URL{
		Scheme: "http",
		User:   nil,
		Host:   "127.0.0.1:21603",
		Path:   "%2e%2e%2ftests/",
	}

	req = &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
	}

	resp, err = c.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()

	_, r, err := get("http://127.0.0.1:21603/../../../../tests/../static/sample.txt")
	assert.NoError(t, err)
	assert.Equal(t, 403, r.StatusCode)
	_ = r.Body.Close()
}

func TestStaticPlugin(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-static.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
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
	t.Run("ServeSample", serveStaticSample)
	t.Run("StaticNotForbid", staticNotForbid)
	t.Run("StaticHeaders", staticHeaders)

	stopCh <- struct{}{}
	wg.Wait()
}

func staticHeaders(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:21603/php_test_files/client.php", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Header.Get("Output") != "output-header" {
		t.Fatal("can't find output header in response")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, all("../../php_test_files/client.php"), string(b))
	require.Equal(t, all("../../php_test_files/client.php"), string(b))
}

func staticNotForbid(t *testing.T) {
	b, r, err := get("http://127.0.0.1:21603/php_test_files/client.php")
	require.NoError(t, err)
	require.Equal(t, all("../../php_test_files/client.php"), b)
	require.Equal(t, all("../../php_test_files/client.php"), b)
	_ = r.Body.Close()
}

func serveStaticSample(t *testing.T) {
	b, r, err := get("http://127.0.0.1:21603/sample.txt")
	require.NoError(t, err)
	require.Contains(t, b, "sample")
	_ = r.Body.Close()
}

func TestStaticDisabled_Error(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-static-disabled.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
	)
	assert.NoError(t, err)
	assert.Error(t, cont.Init())
}

func TestStaticFilesDisabled(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-static-files-disable.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
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
	t.Run("StaticFilesDisabled", staticFilesDisabled)

	stopCh <- struct{}{}
	wg.Wait()
}

func staticFilesDisabled(t *testing.T) {
	b, r, err := get("http://127.0.0.1:45877/php_test_files/client.php?hello=world")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "WORLD", b)
	_ = r.Body.Close()
}

func TestStaticFilesForbid(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-static-files.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Info("http log", "status", 201, "method", "GET", "URI", "http://127.0.0.1:34653/http?hello=world", "remote_address", "127.0.0.1", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("http log", "status", 201, "method", "GET", "URI", "http://127.0.0.1:34653/client.XXX?hello=world", "remote_address", "127.0.0.1", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Info("http log", "status", 201, "method", "GET", "URI", "http://127.0.0.1:34653/client.php?hello=world", "remote_address", "127.0.0.1", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)

	mockLogger.EXPECT().Error("file open error", "error", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("no such file or directory", "error", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("possible path to dir provided").AnyTimes()
	mockLogger.EXPECT().Debug("file extension is forbidden", gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes() // placeholder for the workerlogerror
	mockLogger.EXPECT().Debug("http server is running", "address", gomock.Any()).AnyTimes()

	err = cont.RegisterAll(
		cfg,
		mockLogger,
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&gzip.Plugin{},
		&static.Plugin{},
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
	t.Run("StaticTestFilesDir", staticTestFilesDir)
	t.Run("StaticNotFound", staticNotFound)
	t.Run("StaticFilesForbid", staticFilesForbid)

	stopCh <- struct{}{}
	wg.Wait()
}

func staticTestFilesDir(t *testing.T) {
	b, r, err := get("http://127.0.0.1:34653/http?hello=world")
	assert.NoError(t, err)
	assert.Equal(t, "WORLD", b)
	_ = r.Body.Close()
}

func staticNotFound(t *testing.T) {
	b, _, _ := get("http://127.0.0.1:34653/client.XXX?hello=world") //nolint:bodyclose
	assert.Equal(t, "WORLD", b)
}

func staticFilesForbid(t *testing.T) {
	b, r, err := get("http://127.0.0.1:34653/client.php?hello=world")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "WORLD", b)
	_ = r.Body.Close()
}

func TestHTTPIssue659(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-issue659.yaml",
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
	t.Run("HTTPIssue659", echoIssue659)

	stopCh <- struct{}{}

	wg.Wait()
}

func TestHTTPIPv6Long(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-ipv6.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("RPC plugin started", "address", "tcp://[0:0:0:0:0:0:0:1]:6005", "plugins", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info("http log", "status", 201, "method", "GET", "URI", "http://[0:0:0:0:0:0:0:1]:10684/?hello=world", "remote_address", "::1", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Debug("http server is running", "address", gomock.Any()).AnyTimes()

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		mockLogger,
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
	t.Run("HTTPEchoIPv6-long", echoHTTPIPv6Long)

	stopCh <- struct{}{}

	wg.Wait()
}

func TestHTTPIPv6Short(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http-ipv6-2.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("RPC plugin started", "address", "tcp://[::1]:6003", "plugins", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info("http log", "status", 201, "method", "GET", "URI", "http://[::1]:10784/?hello=world", "remote_address", "::1", "start", gomock.Any(), "elapsed", gomock.Any()).MinTimes(1)
	mockLogger.EXPECT().Debug("http server is running", "address", gomock.Any()).AnyTimes()

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		mockLogger,
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
	t.Run("HTTPEchoIPv6-short", echoHTTPIPv6Short)

	stopCh <- struct{}{}

	wg.Wait()
}

func echoIssue659(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:32552", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Empty(t, b)
	assert.Equal(t, 444, r.StatusCode)

	err = r.Body.Close()
	assert.NoError(t, err)
}

func echoHTTP(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:10084?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func echoHTTPIPv6Long(t *testing.T) {
	req, err := http.NewRequest("GET", "http://[0:0:0:0:0:0:0:1]:10684?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func echoHTTPIPv6Short(t *testing.T) {
	req, err := http.NewRequest("GET", "http://[::1]:10784?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func resetTest(address string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", address)
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		// WorkerList contains list of workers.

		var ret bool
		err = client.Call("resetter.Reset", "http", &ret)
		assert.NoError(t, err)
		assert.True(t, ret)
		ret = false

		var services []string
		err = client.Call("resetter.List", nil, &services)
		assert.NoError(t, err)
		if services[0] != "http" {
			t.Fatal("no enough services")
		}
	}
}

func informerTest(address string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", address)
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		// WorkerList contains list of workers.
		list := struct {
			// Workers is list of workers.
			Workers []process.State `json:"workers"`
		}{}

		err = client.Call("informer.Workers", "http", &list)
		assert.NoError(t, err)
		assert.Len(t, list.Workers, 2)
	}
}

// HELPERS
func get(url string) (string, *http.Response, error) {
	r, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", nil, err
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return "", nil, err
	}

	return string(b), r, err
}

func all(fn string) string {
	f, err := os.Open(fn)
	if err != nil {
		panic(err)
	}

	b := new(bytes.Buffer)
	_, err = io.Copy(b, f)
	if err != nil {
		return ""
	}

	err = f.Close()
	if err != nil {
		return ""
	}

	return b.String()
}
