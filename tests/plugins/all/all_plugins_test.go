package all

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner-plugins/v2/amqp"
	"github.com/spiral/roadrunner-plugins/v2/beanstalk"
	"github.com/spiral/roadrunner-plugins/v2/boltdb"
	"github.com/spiral/roadrunner-plugins/v2/broadcast"
	"github.com/spiral/roadrunner-plugins/v2/config"
	grpcPlugin "github.com/spiral/roadrunner-plugins/v2/grpc"
	httpPlugin "github.com/spiral/roadrunner-plugins/v2/http"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/gzip"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/headers"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/send"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/static"
	"github.com/spiral/roadrunner-plugins/v2/informer"
	"github.com/spiral/roadrunner-plugins/v2/jobs"
	"github.com/spiral/roadrunner-plugins/v2/kv"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/memcached"
	"github.com/spiral/roadrunner-plugins/v2/memory"
	"github.com/spiral/roadrunner-plugins/v2/metrics"
	"github.com/spiral/roadrunner-plugins/v2/redis"
	"github.com/spiral/roadrunner-plugins/v2/reload"
	"github.com/spiral/roadrunner-plugins/v2/resetter"
	rpcPlugin "github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/service"
	"github.com/spiral/roadrunner-plugins/v2/sqs"
	"github.com/spiral/roadrunner-plugins/v2/status"
	"github.com/spiral/roadrunner-plugins/v2/tcp"
	"github.com/spiral/roadrunner-plugins/v2/websockets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllPluginsInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.GracefulShutdownTimeout(time.Second*30))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-all-init.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&metrics.Plugin{},
		&httpPlugin.Plugin{},
		&reload.Plugin{},
		&informer.Plugin{},
		&send.Plugin{},
		&resetter.Plugin{},
		&rpcPlugin.Plugin{},
		&server.Plugin{},
		&service.Plugin{},
		&jobs.Plugin{},
		&tcp.Plugin{},
		&amqp.Plugin{},
		&sqs.Plugin{},
		&beanstalk.Plugin{},
		&grpcPlugin.Plugin{},
		&memory.Plugin{},
		&boltdb.Plugin{},
		&broadcast.Plugin{},
		&websockets.Plugin{},
		&redis.Plugin{},
		&kv.Plugin{},
		&memcached.Plugin{},
		&static.Plugin{},
		&headers.Plugin{},
		&status.Plugin{},
		&gzip.Plugin{},
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

	time.Sleep(time.Second * 5)

	stopCh <- struct{}{}

	wg.Wait()

	time.Sleep(time.Second)
	require.NoError(t, os.RemoveAll("rr.db"))
	require.NoError(t, os.RemoveAll("rr-kv.db"))
}
