package grpc_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/spiral/roadrunner-plugins/v2/tests/plugins/grpc/proto/health"

	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner-plugins/v2/config"
	grpcPlugin "github.com/spiral/roadrunner-plugins/v2/grpc"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/resetter"
	rpcPlugin "github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/tests/plugins/grpc/proto/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestGrpcInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-grpc-init.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&grpcPlugin.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
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
	stopCh <- struct{}{}

	wg.Wait()
}

// test panics -> https://github.com/grpc/grpc-go/blob/master/server.go#L644
func TestGrpcInitDuplicate(t *testing.T) {
	t.Skip("test panics, use locally")
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-grpc-init-duplicate.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&grpcPlugin.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	_, err = cont.Serve()
	assert.NoError(t, err)
}

// different services, same methods inside
func TestGrpcInitDup2(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-grpc-init-duplicate-2.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&grpcPlugin.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
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

func TestGrpcInitMultiple(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-grpc-init-multiple.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&grpcPlugin.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
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
	stopCh <- struct{}{}

	wg.Wait()
}

func TestGrpcRqRs(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-grpc-rq.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&grpcPlugin.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
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

	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)

	client := service.NewEchoClient(conn)
	resp, err := client.Ping(context.Background(), &service.Message{Msg: "TOST"})
	require.NoError(t, err)
	require.Equal(t, "TOST", resp.Msg)

	stopCh <- struct{}{}

	wg.Wait()
}

func TestGrpcRqRsMultiple(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-grpc-rq-multiple.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&grpcPlugin.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
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

	conn, err := grpc.Dial("127.0.0.1:9001", grpc.WithInsecure())
	require.NoError(t, err)
	require.NotNil(t, conn)

	client := service.NewEchoClient(conn)
	resp, err := client.Ping(context.Background(), &service.Message{Msg: "TOST"})
	require.NoError(t, err)
	require.Equal(t, "TOST", resp.Msg)

	hc := health.NewHealthClient(conn)
	hr, err := hc.Check(context.Background(), &health.HealthCheckRequest{})
	require.NoError(t, err)
	require.Equal(t, "SERVING", hr.Status.String())

	stopCh <- struct{}{}

	wg.Wait()
}

func TestGrpcRqRsTLS(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-grpc-rq-tls.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&grpcPlugin.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
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

	creds, err := credentials.NewClientTLSFromFile("./configs/test-certs/test.pem", "")
	require.NoError(t, err)

	conn, err := grpc.Dial("127.0.0.1:9002", grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	require.NotNil(t, conn)

	client := service.NewEchoClient(conn)
	resp, err := client.Ping(context.Background(), &service.Message{Msg: "TOST"})
	require.NoError(t, err)
	require.Equal(t, "TOST", resp.Msg)

	stopCh <- struct{}{}

	wg.Wait()
}

func TestGrpcRqRsTLSRootCA(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("root pool is not available on Windows")
	}
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-grpc-rq-tls-rootca.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&grpcPlugin.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
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

	ca, _ := os.ReadFile("./configs/test-certs/ca.cert")
	require.NotNil(t, ca)

	cert, err := tls.LoadX509KeyPair("./configs/test-certs/test.pem", "./configs/test-certs/test.key")
	require.NoError(t, err)

	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(ca))

	tlscfg := &tls.Config{
		InsecureSkipVerify: true, //nolint:gosec
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12,
		ClientCAs:          pool,
	}

	conn, err := grpc.Dial("127.0.0.1:9003", grpc.WithTransportCredentials(credentials.NewTLS(tlscfg)))
	require.NoError(t, err)
	require.NotNil(t, conn)

	client := service.NewEchoClient(conn)
	resp, err := client.Ping(context.Background(), &service.Message{Msg: "TOST"})
	require.NoError(t, err)
	require.Equal(t, "TOST", resp.Msg)

	stopCh <- struct{}{}

	wg.Wait()
}

func TestGrpcRqRsTLS_WithReset(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-grpc-rq-tls.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&grpcPlugin.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
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

	creds, err := credentials.NewClientTLSFromFile("./configs/test-certs/test.pem", "")
	require.NoError(t, err)

	conn, err := grpc.Dial("localhost:9002", grpc.WithTransportCredentials(creds))
	require.NoError(t, err)
	require.NotNil(t, conn)

	client := service.NewEchoClient(conn)
	resp, err := client.Ping(context.Background(), &service.Message{Msg: "TOST"})
	require.NoError(t, err)
	require.Equal(t, "TOST", resp.Msg)

	// reset
	t.Run("SendReset", sendReset)

	resp2, err2 := client.Ping(context.Background(), &service.Message{Msg: "TOST"})
	require.NoError(t, err2)
	require.Equal(t, "TOST", resp2.Msg)

	stopCh <- struct{}{}
	wg.Wait()
}

func sendReset(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.

	var ret bool
	err = client.Call("resetter.Reset", "grpc", &ret)
	assert.NoError(t, err)
	assert.True(t, ret)
	ret = false

	var services []string
	err = client.Call("resetter.List", nil, &services)
	assert.NotNil(t, services)
	assert.NoError(t, err)
	require.Equal(t, []string{"grpc"}, services)
}
