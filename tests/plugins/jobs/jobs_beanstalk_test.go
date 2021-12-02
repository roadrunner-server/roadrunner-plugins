package jobs

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	jobState "github.com/spiral/roadrunner-plugins/v2/api/jobs"
	jobsv1beta "github.com/spiral/roadrunner-plugins/v2/api/proto/jobs/v1beta"
	"github.com/spiral/roadrunner-plugins/v2/beanstalk"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/informer"
	"github.com/spiral/roadrunner-plugins/v2/jobs"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/resetter"
	rpcPlugin "github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBeanstalkInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*60))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "beanstalk/.rr-beanstalk-init.yaml",
		Prefix: "rr",
	}

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	// general
	mockLogger.EXPECT().Debug("RPC plugin started", "address", "tcp://127.0.0.1:6001", "plugins", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Debug("pipeline started", "driver", "beanstalk", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Debug("pipeline started", "driver", "beanstalk", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Debug("pipeline stopped", "driver", "beanstalk", "pipeline", "test-1", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)
	mockLogger.EXPECT().Debug("pipeline stopped", "driver", "beanstalk", "pipeline", "test-2", "start", gomock.Any(), "elapsed", gomock.Any()).Times(1)

	mockLogger.EXPECT().Warn("beanstalk reserve", "error", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error("job processed with errors", "error", gomock.Any(), "ID", gomock.Any(), "start", gomock.Any(), "elapsed", gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug("beanstalk reserve timeout", "warn", "reserve-with-timeout").AnyTimes()
	mockLogger.EXPECT().Debug("beanstalk listener stopped").AnyTimes()

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		mockLogger,
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&beanstalk.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

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
	stopCh <- struct{}{}
	wg.Wait()
}

func TestBeanstalkDeclare(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*60))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "beanstalk/.rr-beanstalk-declare.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&beanstalk.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

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

	t.Run("DeclareBeanstalkPipeline", declareBeanstalkPipe)
	t.Run("ConsumeBeanstalkPipeline", resumePipes("test-3"))
	t.Run("PushBeanstalkPipeline", pushToPipe("test-3"))
	t.Run("PauseBeanstalkPipeline", pausePipelines("test-3"))
	time.Sleep(time.Second * 3)
	t.Run("DestroyBeanstalkPipeline", destroyPipelines("test-3"))

	time.Sleep(time.Second * 5)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestBeanstalkJobsError(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*60))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "beanstalk/.rr-beanstalk-jobs-err.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&beanstalk.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

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

	t.Run("DeclareBeanstalkPipeline", declareBeanstalkPipe)
	t.Run("ConsumeBeanstalkPipeline", resumePipes("test-3"))
	t.Run("PushBeanstalkPipeline", pushToPipe("test-3"))
	time.Sleep(time.Second * 25)
	t.Run("PauseBeanstalkPipeline", pausePipelines("test-3"))
	time.Sleep(time.Second)
	t.Run("DestroyBeanstalkPipeline", destroyPipelines("test-3"))

	time.Sleep(time.Second * 5)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestBeanstalkStats(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*60))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "beanstalk/.rr-beanstalk-declare.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&beanstalk.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

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

	t.Run("DeclarePipeline", declareBeanstalkPipe)
	t.Run("ConsumePipeline", resumePipes("test-3"))
	t.Run("PushPipeline", pushToPipe("test-3"))
	time.Sleep(time.Second * 2)
	t.Run("PausePipeline", pausePipelines("test-3"))
	time.Sleep(time.Second * 3)
	t.Run("PushPipelineDelayed", pushToPipeDelayed("test-3", 8))
	t.Run("PushPipeline", pushToPipe("test-3"))
	time.Sleep(time.Second)

	out := &jobState.State{}
	t.Run("Stats", stats(out))

	assert.Equal(t, out.Pipeline, "test-3")
	assert.Equal(t, out.Driver, "beanstalk")
	assert.NotEmpty(t, out.Queue)

	out = &jobState.State{}
	t.Run("Stats", stats(out))

	assert.Equal(t, int64(1), out.Active)
	assert.Equal(t, int64(1), out.Delayed)
	assert.Equal(t, int64(0), out.Reserved)

	time.Sleep(time.Second)
	t.Run("ResumePipeline", resumePipes("test-3"))
	time.Sleep(time.Second * 15)

	out = &jobState.State{}
	t.Run("Stats", stats(out))

	assert.Equal(t, out.Pipeline, "test-3")
	assert.Equal(t, out.Driver, "beanstalk")
	assert.NotEmpty(t, out.Queue)

	assert.Equal(t, int64(0), out.Active)
	assert.Equal(t, int64(0), out.Delayed)
	assert.Equal(t, int64(0), out.Reserved)

	time.Sleep(time.Second)
	t.Run("DestroyPipeline", destroyPipelines("test-3"))

	time.Sleep(time.Second)
	stopCh <- struct{}{}
	wg.Wait()
}

func TestBeanstalkNoGlobalSection(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*60))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "beanstalk/.rr-no-global.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&beanstalk.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	_, err = cont.Serve()
	require.Error(t, err)
}

func TestBeanstalkRespond(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*60))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "beanstalk/.rr-beanstalk-respond.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&jobs.Plugin{},
		&resetter.Plugin{},
		&informer.Plugin{},
		&beanstalk.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

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

	t.Run("DeclareBeanstalkPipeline", declareBeanstalkPipe)
	t.Run("ConsumeBeanstalkPipeline", resumePipes("test-3"))
	t.Run("PushBeanstalkPipeline", pushToPipe("test-3"))
	time.Sleep(time.Second * 3)
	t.Run("DestroyBeanstalkPipeline", destroyPipelines("test-3"))
	t.Run("DestroyBeanstalkPipeline", destroyPipelines("test-1"))

	time.Sleep(time.Second * 5)
	stopCh <- struct{}{}
	wg.Wait()
	time.Sleep(time.Second)
}

func declareBeanstalkPipe(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	require.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	pipe := &jobsv1beta.DeclareRequest{Pipeline: map[string]string{
		"driver":          "beanstalk",
		"name":            "test-3",
		"tube":            uuid.NewString(),
		"reserve_timeout": "60s",
		"priority":        "3",
		"tube_priority":   "10",
	}}

	er := &jobsv1beta.Empty{}
	err = client.Call("jobs.Declare", pipe, er)
	require.NoError(t, err)
}
