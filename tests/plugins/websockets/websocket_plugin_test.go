package websockets

import (
	"context"
	"net"
	"net/rpc"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/stretchr/testify/require"

	json "github.com/json-iterator/go"
	endure "github.com/spiral/endure/pkg/container"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner-plugins/v2/broadcast"
	"github.com/spiral/roadrunner-plugins/v2/config"
	httpPlugin "github.com/spiral/roadrunner-plugins/v2/http"
	websocketsv1 "github.com/spiral/roadrunner-plugins/v2/internal/proto/websockets/v1beta"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/memory"
	"github.com/spiral/roadrunner-plugins/v2/redis"
	rpcPlugin "github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/websockets"
	"github.com/spiral/roadrunner/v2/utils"
	"github.com/stretchr/testify/assert"
)

func TestWebsocketsInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-init.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&redis.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&memory.Plugin{},
		&broadcast.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("TestWSInit", wsInit)
	t.Run("RPCWsMemoryPubAsync", RPCWsPubAsync("11111"))
	t.Run("RPCWsMemory", RPCWsPub("11111"))

	stopCh <- struct{}{}

	wg.Wait()
}

func TestWSRedis(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-redis.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&redis.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&broadcast.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("RPCWsRedisPubAsync", RPCWsPubAsync("13235"))
	t.Run("RPCWsRedisPub", RPCWsPub("13235"))

	stopCh <- struct{}{}

	wg.Wait()
}

func TestWSRedisNoSection(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-broker-no-section.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&redis.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&broadcast.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	_, err = cont.Serve()
	assert.Error(t, err)
}

func TestWSDeny(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-deny.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&memory.Plugin{},
		&broadcast.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("RPCWsMemoryDeny", RPCWsDeny("15587"))

	stopCh <- struct{}{}

	wg.Wait()
}

func TestWSDeny2(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-deny2.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&redis.Plugin{},
		&broadcast.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("RPCWsRedisDeny", RPCWsDeny("15588"))

	stopCh <- struct{}{}

	wg.Wait()
}

func TestWSStop(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-stop.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&redis.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&memory.Plugin{},
		&broadcast.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("RPCWsStop", RPCWsMemoryStop("11114"))

	stopCh <- struct{}{}

	wg.Wait()
}

func RPCWsMemoryStop(port string) func(t *testing.T) {
	return func(t *testing.T) {
		connURL := url.URL{Scheme: "ws", Host: "127.0.0.1:" + port, Path: "/ws"}

		_, _, _, err := ws.Dial(context.Background(), connURL.String())
		require.Error(t, err)
		if nErr, ok := err.(ws.StatusError); ok {
			require.Equal(t, "unexpected HTTP response status: 401", nErr.Error())
		}
	}
}

func TestWSAllow(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-allow.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&redis.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&memory.Plugin{},
		&broadcast.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("RPCWsMemoryAllow", RPCWsPub("41278"))

	stopCh <- struct{}{}

	wg.Wait()
}

func TestWSAllow2(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-websockets-allow2.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&redis.Plugin{},
		&websockets.Plugin{},
		&httpPlugin.Plugin{},
		&memory.Plugin{},
		&broadcast.Plugin{},
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

	time.Sleep(time.Second * 1)
	t.Run("RPCWsMemoryAllow", RPCWsPub("41270"))

	stopCh <- struct{}{}

	wg.Wait()
}

func wsInit(t *testing.T) {
	connURL := url.URL{Scheme: "ws", Host: "127.0.0.1:11111", Path: "/ws"}
	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP{"Origin": []string{"127.0.0.1"}},
	}

	conn, _, _, err := dialer.Dial(context.Background(), connURL.String())
	require.NoError(t, err)

	d, err := json.Marshal(messageWS("join", []byte("hello websockets"), "foo", "foo2"))
	require.NoError(t, err)

	err = wsutil.WriteClientBinary(conn, d)
	assert.NoError(t, err)

	msg, err := wsutil.ReadServerBinary(conn)
	require.NoError(t, err)
	_ = msg
	retMsg := utils.AsString(msg)

	// subscription done
	assert.Equal(t, `{"topic":"@join","payload":["foo","foo2"]}`, retMsg)

	err = conn.Close()
	require.NoError(t, err)
}

func RPCWsPubAsync(port string) func(t *testing.T) {
	return func(t *testing.T) {
		connURL := url.URL{Scheme: "ws", Host: "127.0.0.1:" + port, Path: "/ws"}
		conn, _, _, err := ws.Dial(context.Background(), connURL.String())
		require.NoError(t, err)

		go func() {
			messagesToVerify := make([]string, 0, 10)
			messagesToVerify = append(messagesToVerify, `{"topic":"@join","payload":["foo","foo2"]}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"foo","payload":"hello, PHP"}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"@leave","payload":["foo"]}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"foo2","payload":"hello, PHP2"}`)
			i := 0
			for {
				msg, err2 := wsutil.ReadServerBinary(conn)
				retMsg := utils.AsString(msg)
				assert.NoError(t, err2)
				assert.Equal(t, messagesToVerify[i], retMsg)
				i++
				if i == 3 {
					return
				}
			}
		}()

		time.Sleep(time.Second)

		d, err := json.Marshal(messageWS("join", []byte("hello websockets"), "foo", "foo2"))
		require.NoError(t, err)

		err = wsutil.WriteClientBinary(conn, d)
		require.NoError(t, err)

		time.Sleep(time.Second)

		publishAsync(t, "foo")

		time.Sleep(time.Second)

		// //// LEAVE foo /////////
		d, err = json.Marshal(messageWS("leave", []byte("hello websockets"), "foo"))
		if err != nil {
			panic(err)
		}

		err = wsutil.WriteClientBinary(conn, d)
		require.NoError(t, err)

		time.Sleep(time.Second)

		// TRY TO PUBLISH TO UNSUBSCRIBED TOPIC
		publishAsync(t, "foo")

		go func() {
			time.Sleep(time.Second * 5)
			publishAsync(t, "foo2")
		}()

		err = conn.Close()
		assert.NoError(t, err)
	}
}

func RPCWsPub(port string) func(t *testing.T) {
	return func(t *testing.T) {
		connURL := url.URL{Scheme: "ws", Host: "127.0.0.1:" + port, Path: "/ws"}

		conn, _, _, err := ws.Dial(context.Background(), connURL.String())
		require.NoError(t, err)

		go func() {
			messagesToVerify := make([]string, 0, 10)
			messagesToVerify = append(messagesToVerify, `{"topic":"@join","payload":["foo","foo2"]}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"foo","payload":"hello, PHP"}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"@leave","payload":["foo"]}`)
			messagesToVerify = append(messagesToVerify, `{"topic":"foo2","payload":"hello, PHP2"}`)
			i := 0
			for {
				msg, err2 := wsutil.ReadServerBinary(conn)
				retMsg := utils.AsString(msg)
				assert.NoError(t, err2)
				assert.Equal(t, messagesToVerify[i], retMsg)
				i++
				if i == 3 {
					return
				}
			}
		}()

		time.Sleep(time.Second)

		d, err := json.Marshal(messageWS("join", []byte("hello websockets"), "foo", "foo2"))
		if err != nil {
			panic(err)
		}

		err = wsutil.WriteClientBinary(conn, d)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		publish("", "foo")

		time.Sleep(time.Second)

		// //// LEAVE foo /////////
		d, err = json.Marshal(messageWS("leave", []byte("hello websockets"), "foo"))
		if err != nil {
			panic(err)
		}

		err = wsutil.WriteClientBinary(conn, d)
		assert.NoError(t, err)

		time.Sleep(time.Second)

		// TRY TO PUBLISH TO UNSUBSCRIBED TOPIC
		publish("", "foo")

		go func() {
			time.Sleep(time.Second * 5)
			publish2(t, "", "foo2")
		}()

		err = conn.Close()
		assert.NoError(t, err)
	}
}

func RPCWsDeny(port string) func(t *testing.T) {
	return func(t *testing.T) {
		connURL := url.URL{Scheme: "ws", Host: "127.0.0.1:" + port, Path: "/ws"}
		conn, _, _, err := ws.Dial(context.Background(), connURL.String())
		require.NoError(t, err)

		d, err := json.Marshal(messageWS("join", []byte("hello websockets"), "foo", "foo2"))
		if err != nil {
			panic(err)
		}

		err = wsutil.WriteClientBinary(conn, d)
		assert.NoError(t, err)

		msg, err := wsutil.ReadServerBinary(conn)
		require.NoError(t, err)
		retMsg := utils.AsString(msg)

		// subscription done
		assert.Equal(t, `{"topic":"#join","payload":["foo","foo2"]}`, retMsg)

		// //// LEAVE foo, foo2 /////////
		d, err = json.Marshal(messageWS("leave", []byte("hello websockets"), "foo"))
		if err != nil {
			panic(err)
		}

		err = wsutil.WriteClientBinary(conn, d)
		assert.NoError(t, err)

		msg, err = wsutil.ReadServerBinary(conn)
		require.NoError(t, err)
		retMsg = utils.AsString(msg)

		// subscription done
		assert.Equal(t, `{"topic":"@leave","payload":["foo"]}`, retMsg)

		err = conn.Close()
		assert.NoError(t, err)
	}
}

// ---------------------------------------------------------------------------------------------------

func publish(topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	ret := &websocketsv1.Response{}
	err = client.Call("broadcast.Publish", makeMessage([]byte("hello, PHP"), topics...), ret)
	if err != nil {
		panic(err)
	}
}

func publishAsync(t *testing.T, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	ret := &websocketsv1.Response{}
	err = client.Call("broadcast.PublishAsync", makeMessage([]byte("hello, PHP"), topics...), ret)
	assert.NoError(t, err)
	assert.True(t, ret.Ok)
}

func publish2(t *testing.T, topics ...string) {
	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))

	ret := &websocketsv1.Response{}
	err = client.Call("broadcast.Publish", makeMessage([]byte("hello, PHP2"), topics...), ret)
	assert.NoError(t, err)
	assert.True(t, ret.Ok)
}

func messageWS(command string, payload []byte, topics ...string) *websocketsv1.Message {
	return &websocketsv1.Message{
		Topics:  topics,
		Command: command,
		Payload: payload,
	}
}

func makeMessage(payload []byte, topics ...string) *websocketsv1.Request {
	m := &websocketsv1.Request{
		Messages: []*websocketsv1.Message{
			{
				Topics:  topics,
				Payload: payload,
			},
		},
	}

	return m
}
