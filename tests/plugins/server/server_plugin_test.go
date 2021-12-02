package server

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppPipes(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"

	err = container.RegisterAll(
		vp,
		&server.Plugin{},
		&Foo{},
		&logger.ZapLogger{},
	)
	require.NoError(t, err)

	err = container.Init()
	require.NoError(t, err)

	errCh, err := container.Serve()
	require.NoError(t, err)

	// stop by CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	tt := time.NewTimer(time.Second * 10)
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
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
	}()

	wg.Wait()
}

func TestAppSockets(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-sockets.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	require.NoError(t, err)

	err = container.Register(&server.Plugin{})
	require.NoError(t, err)

	err = container.Register(&Foo2{})
	require.NoError(t, err)

	err = container.Register(&logger.ZapLogger{})
	require.NoError(t, err)

	err = container.Init()
	require.NoError(t, err)

	errCh, err := container.Serve()
	require.NoError(t, err)

	// stop by CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// stop after 10 seconds
	tt := time.NewTicker(time.Second * 10)

	for {
		select {
		case e := <-errCh:
			assert.NoError(t, e.Error)
			assert.NoError(t, container.Stop())
			return
		case <-c:
			er := container.Stop()
			if er != nil {
				panic(er)
			}
			return
		case <-tt.C:
			tt.Stop()
			assert.NoError(t, container.Stop())
			return
		}
	}
}

func TestAppTCPOnInit(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-tcp-on-init.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	require.NoError(t, err)

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	// process interrupt error
	mockLogger.EXPECT().Info("The number is: 0\n").Times(1)
	mockLogger.EXPECT().Info("The number is: 1\n").Times(1)
	mockLogger.EXPECT().Info("The number is: 2\n").Times(1)
	mockLogger.EXPECT().Info("The number is: 3\n").Times(1)
	mockLogger.EXPECT().Info("The number is: 4\n").Times(1)
	mockLogger.EXPECT().Info("The number is: 5\n").Times(1)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	err = container.RegisterAll(
		mockLogger,
		&server.Plugin{},
		&Foo2{},
	)

	err = container.Init()
	require.NoError(t, err)

	ch, err := container.Serve()
	require.NoError(t, err)

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

	time.Sleep(time.Second * 10)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestAppSocketsOnInit(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-sockets-on-init.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	require.NoError(t, err)

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	// process interrupt error
	mockLogger.EXPECT().Info("The number is: 0\n").Times(1)
	mockLogger.EXPECT().Info("The number is: 1\n").Times(1)
	mockLogger.EXPECT().Info("The number is: 2\n").Times(1)
	mockLogger.EXPECT().Info("The number is: 3\n").Times(1)
	mockLogger.EXPECT().Info("The number is: 4\n").Times(1)
	mockLogger.EXPECT().Info("The number is: 5\n").Times(1)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	err = container.RegisterAll(
		mockLogger,
		&server.Plugin{},
		&Foo2{},
	)

	err = container.Init()
	require.NoError(t, err)

	ch, err := container.Serve()
	require.NoError(t, err)

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

	time.Sleep(time.Second * 10)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestAppSocketsOnInitFastClose(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}
	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-sockets-on-init-fast-close.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	require.NoError(t, err)

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	// process interrupt error
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error("process wait", "error", gomock.Any()).Times(1)

	err = container.RegisterAll(
		mockLogger,
		&server.Plugin{},
		&Foo2{},
	)

	err = container.Init()
	require.NoError(t, err)

	ch, err := container.Serve()
	require.NoError(t, err)

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

	time.Sleep(time.Second * 10)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestAppTCP(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-tcp.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	require.NoError(t, err)

	err = container.Register(&server.Plugin{})
	require.NoError(t, err)

	err = container.Register(&Foo3{})
	require.NoError(t, err)

	err = container.Register(&logger.ZapLogger{})
	require.NoError(t, err)

	err = container.Init()
	require.NoError(t, err)

	errCh, err := container.Serve()
	require.NoError(t, err)

	// stop by CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// stop after 10 seconds
	tt := time.NewTicker(time.Second * 10)

	for {
		select {
		case e := <-errCh:
			assert.NoError(t, e.Error)
			assert.NoError(t, container.Stop())
			return
		case <-c:
			er := container.Stop()
			if er != nil {
				panic(er)
			}
			return
		case <-tt.C:
			tt.Stop()
			assert.NoError(t, container.Stop())
			return
		}
	}
}

func TestAppWrongConfig(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)
	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rrrrrrrrrr.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	require.NoError(t, err)

	err = container.Register(&server.Plugin{})
	require.NoError(t, err)

	err = container.Register(&Foo3{})
	require.NoError(t, err)

	err = container.Register(&logger.ZapLogger{})
	require.NoError(t, err)

	require.Error(t, container.Init())
}

func TestAppWrongRelay(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-wrong-relay.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	require.NoError(t, err)

	err = container.Register(&server.Plugin{})
	require.NoError(t, err)

	err = container.Register(&Foo3{})
	require.NoError(t, err)

	err = container.Register(&logger.ZapLogger{})
	require.NoError(t, err)

	err = container.Init()
	require.Error(t, err)

	_, err = container.Serve()
	require.Error(t, err)

	_ = container.Stop()
}

func TestAppWrongCommand(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-wrong-command.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	require.NoError(t, err)

	err = container.Register(&server.Plugin{})
	require.NoError(t, err)

	err = container.Register(&Foo3{})
	require.NoError(t, err)

	err = container.Register(&logger.ZapLogger{})
	require.NoError(t, err)

	err = container.Init()
	require.NoError(t, err)

	_, err = container.Serve()
	require.Error(t, err)
}

func TestAppWrongCommandOnInit(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-wrong-command-on-init.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	require.NoError(t, err)

	err = container.Register(&server.Plugin{})
	require.NoError(t, err)

	err = container.Register(&Foo3{})
	require.NoError(t, err)

	err = container.Register(&logger.ZapLogger{})
	require.NoError(t, err)

	err = container.Init()
	require.NoError(t, err)

	_, err = container.Serve()
	require.Error(t, err)
}

func TestAppNoAppSectionInConfig(t *testing.T) {
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.SetLogLevel(endure.ErrorLevel))
	require.NoError(t, err)

	// config plugin
	vp := &config.Viper{}
	vp.Path = "configs/.rr-wrong-command.yaml"
	vp.Prefix = "rr"
	err = container.Register(vp)
	require.NoError(t, err)

	err = container.Register(&server.Plugin{})
	require.NoError(t, err)

	err = container.Register(&Foo3{})
	require.NoError(t, err)

	err = container.Register(&logger.ZapLogger{})
	require.NoError(t, err)

	err = container.Init()
	require.NoError(t, err)

	_, err = container.Serve()
	require.Error(t, err)
}
