package main

import (
	"log"
	rand2 "math/rand"
	"net"
	"net/rpc"
	"os"
	"path"
	"sync"

	"github.com/google/uuid"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	jobsv1beta "github.com/spiral/roadrunner-plugins/v2/api/proto/jobs/v1beta"
)

const (
	push    string = "jobs.Push"
	pause   string = "jobs.Pause"
	destroy string = "jobs.Destroy"
	declare string = "jobs.Declare"
	resume  string = "jobs.Resume"
	stat    string = "jobs.Stat"
)

func main() {
	wg := &sync.WaitGroup{}
	wg.Add(33)

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 1000; j++ {
					n := uuid.NewString()
					declareAMQPPipe(client, n)
					startPipelines(client, n)
					push100(client, n)
					pausePipelines(client, n)
					destroyPipelines(client, n)
				}
				wg.Done()
			}()
		}
	}()

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for i := 0; i < 5; i++ {
			go func() {
				for j := 0; j < 1000; j++ {
					n := uuid.NewString()
					declareBeanstalkPipe(client, n)
					startPipelines(client, n)
					push100(client, n)
					pausePipelines(client, n)
					destroyPipelines(client, n)
				}
				wg.Done()
			}()
		}
	}()

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for i := 0; i < 3; i++ {
			go func() {
				for j := 0; j < 10; j++ {
					n := uuid.NewString()
					declareBoltDBPipe(client, n, n)
					startPipelines(client, n)
					push100(client, n)
					pausePipelines(client, n)
					destroyPipelines(client, n)
					cur, err := os.Getwd()
					if err != nil {
						panic(err)
					}
					_ = os.Remove(path.Join(cur, n))
				}
				wg.Done()
			}()
		}
	}()

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 1000; j++ {
					n := uuid.NewString()
					declareMemoryPipe(client, n)
					startPipelines(client, n)
					push100(client, n)
					pausePipelines(client, n)
					destroyPipelines(client, n)
				}
				wg.Done()
			}()
		}
	}()

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:6001")
		if err != nil {
			log.Fatal(err)
		}
		client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
		for i := 0; i < 5; i++ {
			go func() {
				for j := 0; j < 1000; j++ {
					n := uuid.NewString()
					declareSQSPipe(client, n)
					startPipelines(client, n)
					push100(client, n)
					pausePipelines(client, n)
					destroyPipelines(client, n)
				}
				wg.Done()
			}()
		}
	}()

	wg.Wait()

	conn, err := net.Dial("tcp", "127.0.0.1:6001")
	if err != nil {
		log.Fatal(err)
	}
	client := rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn))
	ll := list(client)

	for i := 0; i < len(ll); i++ {
		destroyPipelines(client, ll[i])
	}
}

func push100(client *rpc.Client, pipe string) {
	for j := 0; j < 100; j++ {
		payloads := &jobsv1beta.PushRequest{
			Job: &jobsv1beta.Job{
				Job:     "Some/Super/PHP/Class",
				Id:      uuid.NewString(),
				Payload: "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.",
				Headers: map[string]*jobsv1beta.HeaderValue{"test": {Value: []string{"hello"}}},
				Options: &jobsv1beta.Options{
					Priority: int64(rand2.Intn(100) + 1),
					Pipeline: pipe,
				},
			},
		}

		resp := jobsv1beta.Empty{}
		err := client.Call(push, payloads, &resp)
		if err != nil {
			log.Println(err)
		}
	}
}

func startPipelines(client *rpc.Client, pipes ...string) {
	pipe := &jobsv1beta.Pipelines{Pipelines: make([]string, len(pipes))}

	for i := 0; i < len(pipes); i++ {
		pipe.GetPipelines()[i] = pipes[i]
	}

	er := &jobsv1beta.Empty{}
	err := client.Call(resume, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func pausePipelines(client *rpc.Client, pipes ...string) {
	pipe := &jobsv1beta.Pipelines{Pipelines: make([]string, len(pipes))}

	for i := 0; i < len(pipes); i++ {
		pipe.GetPipelines()[i] = pipes[i]
	}

	er := &jobsv1beta.Empty{}
	err := client.Call(pause, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func destroyPipelines(client *rpc.Client, pipes ...string) {
	pipe := &jobsv1beta.Pipelines{Pipelines: make([]string, len(pipes))}

	for i := 0; i < len(pipes); i++ {
		pipe.GetPipelines()[i] = pipes[i]
	}

	er := &jobsv1beta.Empty{}
	err := client.Call(destroy, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func list(client *rpc.Client) []string {
	resp := &jobsv1beta.Pipelines{}
	er := &jobsv1beta.Empty{}
	err := client.Call("jobs.List", er, resp)
	if err != nil {
		log.Println(err)
	}

	l := make([]string, len(resp.GetPipelines()))

	for i := 0; i < len(resp.GetPipelines()); i++ {
		l[i] = resp.GetPipelines()[i]
	}
	return l
}

func declareSQSPipe(client *rpc.Client, n string) {
	pipe := &jobsv1beta.DeclareRequest{Pipeline: map[string]string{
		"driver":             "sqs",
		"name":               n,
		"queue":              n,
		"prefetch":           "10",
		"priority":           "3",
		"visibility_timeout": "0",
		"wait_time_seconds":  "3",
		"tags":               `{"key":"value"}`,
	}}

	er := &jobsv1beta.Empty{}
	err := client.Call(declare, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func declareAMQPPipe(client *rpc.Client, p string) {
	pipe := &jobsv1beta.DeclareRequest{Pipeline: map[string]string{
		"driver":          "amqp",
		"name":            p,
		"routing_key":     "test-3",
		"queue":           "default",
		"exchange_type":   "direct",
		"exchange":        "amqp.default",
		"prefetch":        "100",
		"priority":        "4",
		"exclusive":       "false",
		"multiple_ask":    "false",
		"requeue_on_fail": "false",
	}}

	er := &jobsv1beta.Empty{}
	err := client.Call(declare, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func declareBeanstalkPipe(client *rpc.Client, n string) {
	pipe := &jobsv1beta.DeclareRequest{Pipeline: map[string]string{
		"driver":          "beanstalk",
		"name":            n,
		"tube":            n,
		"reserve_timeout": "1",
		"priority":        "3",
		"tube_priority":   "10",
	}}

	er := &jobsv1beta.Empty{}
	err := client.Call(declare, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func declareBoltDBPipe(client *rpc.Client, n, file string) {
	pipe := &jobsv1beta.DeclareRequest{Pipeline: map[string]string{
		"driver":   "boltdb",
		"name":     n,
		"prefetch": "100",
		"priority": "2",
		"file":     file,
	}}

	er := &jobsv1beta.Empty{}
	err := client.Call(declare, pipe, er)
	if err != nil {
		log.Println(err)
	}
}

func declareMemoryPipe(client *rpc.Client, p string) {
	pipe := &jobsv1beta.DeclareRequest{Pipeline: map[string]string{
		"driver":   "memory",
		"name":     p,
		"prefetch": "10000",
		"priority": "1",
	}}

	er := &jobsv1beta.Empty{}
	err := client.Call(declare, pipe, er)
	if err != nil {
		log.Println(err)
	}
}
