package tcp

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	json "github.com/json-iterator/go"
	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"
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

	time.Sleep(time.Second * 3)
	c, err := net.Dial("tcp", "127.0.0.1:7777")
	require.NoError(t, err)
	_, err = c.Write([]byte("\n\r\n"))
	require.NoError(t, err)

	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	require.NoError(t, err)

	var d map[string]interface{}
	err = json.Unmarshal(buf[:n], &d)
	require.NoError(t, err)

	require.Equal(t, d["remote_addr"].(string), c.LocalAddr().String())

	// ---

	c, err = net.Dial("tcp", "127.0.0.1:8889")
	require.NoError(t, err)
	_, err = c.Write([]byte("helooooo\r\n"))
	require.NoError(t, err)

	buf = make([]byte, 1024)
	n, err = c.Read(buf)
	require.NoError(t, err)

	err = json.Unmarshal(buf[:n], &d)
	require.NoError(t, err)

	require.Equal(t, d["remote_addr"].(string), c.LocalAddr().String())

	// ---

	c, err = net.Dial("tcp", "127.0.0.1:8810")
	require.NoError(t, err)
	_, err = c.Write([]byte("HEEEEEEEEEEEEEYYYYYYYYYYYYY\r\n"))
	require.NoError(t, err)

	buf = make([]byte, 1024)
	n, err = c.Read(buf)
	require.NoError(t, err)

	err = json.Unmarshal(buf[:n], &d)
	require.NoError(t, err)

	require.Equal(t, d["remote_addr"].(string), c.LocalAddr().String())

	stopCh <- struct{}{}
	wg.Wait()
}

func TestTCPEmptySend(t *testing.T) {
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

	time.Sleep(time.Second * 3)
	c, err := net.Dial("tcp", "127.0.0.1:7777")
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

	waitCh := make(chan struct{}, 3)

	go func() {
		time.Sleep(time.Second * 3)
		c, err := net.Dial("tcp", "127.0.0.1:7778")
		require.NoError(t, err)

		buf := make([]byte, 1024)
		n, err := c.Read(buf)
		require.NoError(t, err)

		require.Equal(t, []byte("hello \r\n"), buf[:n])

		_, err = c.Write([]byte("hello \r\n"))
		require.NoError(t, err)

		n, err = c.Read(buf)
		require.NoError(t, err)

		var d map[string]interface{}
		err = json.Unmarshal(buf[:n], &d)
		require.NoError(t, err)

		require.Equal(t, d["remote_addr"].(string), c.LocalAddr().String())
		waitCh <- struct{}{}
	}()
	go func() {
		time.Sleep(time.Second * 3)
		c, err := net.Dial("tcp", "127.0.0.1:8811")
		require.NoError(t, err)

		buf := make([]byte, 1024)
		n, err := c.Read(buf)
		require.NoError(t, err)

		require.Equal(t, []byte("hello \r\n"), buf[:n])

		_, err = c.Write([]byte("hello \r\n"))
		require.NoError(t, err)

		n, err = c.Read(buf)
		require.NoError(t, err)

		var d map[string]interface{}
		err = json.Unmarshal(buf[:n], &d)
		require.NoError(t, err)

		require.Equal(t, d["remote_addr"].(string), c.LocalAddr().String())
		waitCh <- struct{}{}
	}()
	go func() {
		time.Sleep(time.Second * 3)
		c, err := net.Dial("tcp", "127.0.0.1:8812")
		require.NoError(t, err)

		buf := make([]byte, 1024)
		n, err := c.Read(buf)
		require.NoError(t, err)

		require.Equal(t, []byte("hello \r\n"), buf[:n])

		_, err = c.Write([]byte("hello \r\n"))
		require.NoError(t, err)

		n, err = c.Read(buf)
		require.NoError(t, err)

		var d map[string]interface{}
		err = json.Unmarshal(buf[:n], &d)
		require.NoError(t, err)

		require.Equal(t, d["remote_addr"].(string), c.LocalAddr().String())
		waitCh <- struct{}{}
	}()

	// ---

	<-waitCh
	<-waitCh
	<-waitCh
	stopCh <- struct{}{}
	wg.Wait()
}
