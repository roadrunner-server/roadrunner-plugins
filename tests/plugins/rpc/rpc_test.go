package rpc

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/stretchr/testify/assert"
)

// graph https://bit.ly/3ensdNb
func TestRpcInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Plugin1{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Plugin2{})
	if err != nil {
		t.Fatal(err)
	}

	v := &config.Viper{}
	v.Path = "configs/.rr.yaml"
	v.Prefix = "rr"
	err = cont.Register(v)
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&rpc.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&logger.ZapLogger{})
	if err != nil {
		t.Fatal(err)
	}

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

	tt := time.NewTimer(time.Second * 3)

	go func() {
		defer wg.Done()
		defer tt.Stop()
		for {
			select {
			case e := <-ch:
				// just stop, this is ok
				assert.Error(t, e.Error)
				_ = cont.Stop()
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-tt.C:
				return
			}
		}
	}()

	wg.Wait()
}

// graph https://bit.ly/3ensdNb
func TestRpcDisabled(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Plugin1{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Plugin2{})
	if err != nil {
		t.Fatal(err)
	}

	v := &config.Viper{}
	v.Path = "configs/.rr-rpc-disabled.yaml"
	v.Prefix = "rr"
	err = cont.Register(v)
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&rpc.Plugin{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&logger.ZapLogger{})
	if err != nil {
		t.Fatal(err)
	}

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
	tt := time.NewTimer(time.Second * 20)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer tt.Stop()
		for {
			select {
			case e := <-ch:
				// RPC is turned off, should be and dial error
				if errors.Is(errors.Disabled, e.Error) {
					assert.FailNow(t, "should not be disabled error")
				}
				assert.Error(t, e.Error)
				err = cont.Stop()
				assert.Error(t, err)
				return
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-tt.C:
				// timeout
				return
			}
		}
	}()

	wg.Wait()
}
