package informer

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/http"
	"github.com/spiral/roadrunner-plugins/v2/informer"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/resetter"
	rpcPlugin "github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/status"
	"github.com/spiral/roadrunner/v2/state/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInformerInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Viper{
		Path:   "configs/.rr-informer.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&logger.ZapLogger{},
		&informer.Plugin{},
		&rpcPlugin.Plugin{},
		&Plugin1{},
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

	stopCh := make(chan struct{}, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
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

	time.Sleep(time.Second)
	t.Run("InformerWorkersRpcTest", informerWorkersRPCTest("informer.plugin1"))
	t.Run("InformerListRpcTest", informerListRPCTest)
	t.Run("InformerPluginWithoutWorkersRpcTest", informerPluginWOWorkersRPCTest)

	stopCh <- struct{}{}
	wg.Wait()
}

func TestInformerEarlyCall(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Viper{
		Path:   "configs/.rr-informer-early-call.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&server.Plugin{},
		&logger.ZapLogger{},
		&http.Plugin{},
		&informer.Plugin{},
		&resetter.Plugin{},
		&status.Plugin{},
		&rpcPlugin.Plugin{},
		&Plugin2{},
	)

	require.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	require.NoError(t, err)

	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	list := struct {
		// list of workers.
		Workers []process.State `json:"workers"`
	}{}

	err = client.Call("informer.Workers", "informer.plugin2", &list)
	require.NoError(t, err)
	require.Len(t, list.Workers, 0)

	sig := make(chan os.Signal, 0)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	stopCh := make(chan struct{}, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
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

	time.Sleep(time.Second)
	stopCh <- struct{}{}
	wg.Wait()
}

func informerPluginWOWorkersRPCTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	list := struct {
		// Workers is list of workers.
		Workers []process.State `json:"workers"`
	}{}

	err = client.Call("informer.Workers", "informer.config", &list)
	assert.NoError(t, err)
	assert.Len(t, list.Workers, 0)
}

func informerWorkersRPCTest(service string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		assert.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		// WorkerList contains list of workers.
		list := struct {
			// Workers is list of workers.
			Workers []process.State `json:"workers"`
		}{}

		err = client.Call("informer.Workers", service, &list)
		assert.NoError(t, err)
		assert.Len(t, list.Workers, 10)
	}
}

func informerListRPCTest(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	assert.NoError(t, err)
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	// WorkerList contains list of workers.
	list := make([]string, 0, 5)
	// Plugins which are expected to be in the list
	expected := []string{"informer.plugin1"}

	err = client.Call("informer.List", true, &list)
	assert.NoError(t, err)
	assert.ElementsMatch(t, list, expected)
}
