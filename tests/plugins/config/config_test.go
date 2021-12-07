package config

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner-plugins/v2/amqp"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/jobs"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViperProvider_Init(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	vp := &config.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = nil

	err = container.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&Foo{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Init()
	if err != nil {
		t.Fatal(err)
	}

	errCh, err := container.Serve()
	if err != nil {
		t.Fatal(err)
	}

	// stop by CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	tt := time.NewTicker(time.Second * 2)
	defer tt.Stop()

	for {
		select {
		case e := <-errCh:
			assert.NoError(t, e.Error)
			assert.NoError(t, container.Stop())
			return
		case <-c:
			er := container.Stop()
			assert.NoError(t, er)
			return
		case <-tt.C:
			assert.NoError(t, container.Stop())
			return
		}
	}
}

func TestConfigOverwriteFail(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(false), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	vp := &config.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = []string{"rpc.listen=tcp//not_exist"}

	err = container.RegisterAll(
		&logger.ZapLogger{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.Error(t, err)
}

func TestConfigOverwriteFail_2(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(false), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	vp := &config.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = []string{"rpc.listen="}

	err = container.RegisterAll(
		&logger.ZapLogger{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.Error(t, err)
}

func TestConfigOverwriteFail_3(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(false), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	vp := &config.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = []string{"="}

	err = container.RegisterAll(
		&logger.ZapLogger{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.Error(t, err)
}

func TestConfigOverwriteValid(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(false), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	vp := &config.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = []string{"rpc.listen=tcp://127.0.0.1:36643"}

	err = container.RegisterAll(
		&logger.ZapLogger{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.NoError(t, err)

	errCh, err := container.Serve()
	assert.NoError(t, err)

	// stop by CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	tt := time.NewTicker(time.Second * 3)
	defer tt.Stop()

	for {
		select {
		case e := <-errCh:
			assert.NoError(t, e.Error)
			assert.NoError(t, container.Stop())
			return
		case <-c:
			er := container.Stop()
			assert.NoError(t, er)
			return
		case <-tt.C:
			assert.NoError(t, container.Stop())
			return
		}
	}
}

func TestConfigEnvVariables(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(false), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("SUPER_RPC_ENV", "tcp://127.0.0.1:36643")
	assert.NoError(t, err)

	vp := &config.Plugin{}
	vp.Path = "configs/.rr-env.yaml"
	vp.Prefix = "rr"

	err = container.RegisterAll(
		&logger.ZapLogger{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.NoError(t, err)

	errCh, err := container.Serve()
	assert.NoError(t, err)

	// stop by CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	tt := time.NewTicker(time.Second * 3)
	defer tt.Stop()

	for {
		select {
		case e := <-errCh:
			assert.NoError(t, e.Error)
			assert.NoError(t, container.Stop())
			return
		case <-c:
			er := container.Stop()
			assert.NoError(t, er)
			return
		case <-tt.C:
			assert.NoError(t, container.Stop())
			return
		}
	}
}

func TestConfigEnvVariablesFail(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(false), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("SUPER_RPC_ENV", "tcp://127.0.0.1:6065")
	assert.NoError(t, err)

	vp := &config.Plugin{}
	vp.Path = "configs/.rr-env.yaml"
	vp.Prefix = "rr"

	err = container.RegisterAll(
		&logger.ZapLogger{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.NoError(t, err)

	_, err = container.Serve()
	assert.Error(t, err)
}

func TestConfigProvider_GeneralSection(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	vp := &config.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = nil
	vp.CommonConfig = &config.General{GracefulTimeout: time.Second * 10}

	err = container.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = container.Register(&Foo3{})
	if err != nil {
		t.Fatal(err)
	}

	err = container.Init()
	if err != nil {
		t.Fatal(err)
	}

	errCh, err := container.Serve()
	if err != nil {
		t.Fatal(err)
	}

	// stop by CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	tt := time.NewTicker(time.Second * 2)
	defer tt.Stop()

	for {
		select {
		case e := <-errCh:
			assert.NoError(t, e.Error)
			assert.NoError(t, container.Stop())
			return
		case <-c:
			er := container.Stop()
			assert.NoError(t, er)
			return
		case <-tt.C:
			assert.NoError(t, container.Stop())
			return
		}
	}
}

// VERSIONS

func TestViperProvider_Init_Version(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}

	vp := &config.Plugin{}
	vp.Path = "configs/.rr-init-version.yaml"
	vp.Prefix = "rr"
	vp.Flags = nil
	vp.RRVersion = "2.7.2"

	err = container.RegisterAll(
		&jobs.Plugin{},
		&amqp.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		vp,
	)

	require.NoError(t, err)

	err = container.Init()
	require.NoError(t, err)

	ch, err := container.Serve()
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
				err = container.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-sig:
				err = container.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = container.Stop()
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
