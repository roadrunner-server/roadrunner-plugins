package tcp

import (
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	json "github.com/json-iterator/go"
	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	rpcPlugin "github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/tcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTCPInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-tcp-init.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&tcp.Plugin{},
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
	c, err := net.Dial("tcp", "127.0.0.1:7777")
	require.NoError(t, err)
	_, err = c.Write([]byte("wuzaaaa\n\r\n"))
	require.NoError(t, err)

	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	require.NoError(t, err)

	var d1 map[string]interface{}
	err = json.Unmarshal(buf[:n], &d1)
	require.NoError(t, err)

	require.Equal(t, d1["remote_addr"].(string), c.LocalAddr().String())

	// ---

	c, err = net.Dial("tcp", "127.0.0.1:8889")
	require.NoError(t, err)
	_, err = c.Write([]byte("helooooo\r\n"))
	require.NoError(t, err)

	buf = make([]byte, 1024)
	n, err = c.Read(buf)
	require.NoError(t, err)

	var d2 map[string]interface{}
	err = json.Unmarshal(buf[:n], &d2)
	require.NoError(t, err)

	require.Equal(t, d2["remote_addr"].(string), c.LocalAddr().String())

	// ---

	c, err = net.Dial("tcp", "127.0.0.1:8810")
	require.NoError(t, err)
	_, err = c.Write([]byte("HEEEEEEEEEEEEEYYYYYYYYYYYYY\r\n"))
	require.NoError(t, err)

	buf = make([]byte, 1024)
	n, err = c.Read(buf)
	require.NoError(t, err)

	var d3 map[string]interface{}
	err = json.Unmarshal(buf[:n], &d3)
	require.NoError(t, err)

	require.Equal(t, d3["remote_addr"].(string), c.LocalAddr().String())

	stopCh <- struct{}{}
	wg.Wait()
}

func TestTCPEmptySend(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-tcp-empty.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&tcp.Plugin{},
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
	c, err := net.Dial("tcp", "127.0.0.1:7779")
	require.NoError(t, err)
	_, err = c.Write([]byte(""))
	require.NoError(t, err)

	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	require.NoError(t, err)

	var d map[string]interface{}
	err = json.Unmarshal(buf[:n], &d)
	require.NoError(t, err)

	require.Equal(t, d["remote_addr"].(string), c.LocalAddr().String())

	// ---

	stopCh <- struct{}{}
	wg.Wait()
}

func TestTCPConnClose(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-tcp-close.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&tcp.Plugin{},
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
	c, err := net.Dial("tcp", "127.0.0.1:7788")
	require.NoError(t, err)
	_, err = c.Write([]byte("hello \r\n"))
	require.NoError(t, err)

	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	require.NoError(t, err)

	var d map[string]interface{}
	err = json.Unmarshal(buf[:n], &d)
	require.NoError(t, err)

	require.NotEmpty(t, d["uuid"].(string))

	t.Run("CloseConnection", closeConn(d["uuid"].(string)))
	// ---

	stopCh <- struct{}{}
	wg.Wait()
}

func TestTCPFull(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-tcp-full.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&server.Plugin{},
		&tcp.Plugin{},
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
	waitCh := make(chan struct{}, 3)

	go func() {
		c, err := net.Dial("tcp", "127.0.0.1:7778")
		require.NoError(t, err)

		buf := make([]byte, 1024)
		n, err := c.Read(buf)
		require.NoError(t, err)

		require.Equal(t, []byte("hello \r\n"), buf[:n])

		var d map[string]interface{}
		for i := 0; i < 100; i++ {
			_, err = c.Write([]byte("foo \r\n"))
			require.NoError(t, err)

			n, err = c.Read(buf)
			require.NoError(t, err)

			err = json.Unmarshal(buf[:n], &d)
			require.NoError(t, err)

			require.Equal(t, d["remote_addr"].(string), "foo1")
			require.Equal(t, d["body"].(string), "foo \r\n")
		}
		waitCh <- struct{}{}
	}()

	go func() {
		c, err := net.Dial("tcp", "127.0.0.1:8811")
		require.NoError(t, err)

		buf := make([]byte, 1024)
		n, err := c.Read(buf)
		require.NoError(t, err)

		require.Equal(t, []byte("hello \r\n"), buf[:n])

		var d map[string]interface{}
		for i := 0; i < 100; i++ {
			_, err = c.Write([]byte("bar \r\n"))
			require.NoError(t, err)

			n, err = c.Read(buf)
			require.NoError(t, err)

			err = json.Unmarshal(buf[:n], &d)
			require.NoError(t, err)

			require.Equal(t, d["remote_addr"].(string), "foo2")
			require.Equal(t, d["body"].(string), "bar \r\n")
		}
		waitCh <- struct{}{}
	}()

	go func() {
		c, err := net.Dial("tcp", "127.0.0.1:8812")
		require.NoError(t, err)

		buf := make([]byte, 1024)
		n, err := c.Read(buf)
		require.NoError(t, err)

		require.Equal(t, []byte("hello \r\n"), buf[:n])

		var d map[string]interface{}
		for i := 0; i < 100; i++ {
			_, err = c.Write([]byte("baz \r\n"))
			require.NoError(t, err)

			n, err = c.Read(buf)
			require.NoError(t, err)

			err = json.Unmarshal(buf[:n], &d)
			require.NoError(t, err)

			require.Equal(t, d["remote_addr"].(string), "foo3")
			require.Equal(t, d["body"].(string), "baz \r\n")
		}
		waitCh <- struct{}{}
	}()

	// ---

	<-waitCh
	<-waitCh
	<-waitCh
	stopCh <- struct{}{}
	wg.Wait()
}

func closeConn(uuid string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		require.NoError(t, err)
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		var ret bool
		err = client.Call("tcp.Close", uuid, &ret)
		require.NoError(t, err)
		require.True(t, ret)
	}
}
