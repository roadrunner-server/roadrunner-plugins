package jobs

import (
	"context"

	"github.com/spiral/roadrunner-plugins/v2/jobs/job"
	"github.com/spiral/roadrunner-plugins/v2/jobs/pipeline"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
)

// Consumer represents a single jobs driver interface
type Consumer interface {
	Push(ctx context.Context, job *job.Job) error
	Register(ctx context.Context, pipeline *pipeline.Pipeline) error
	Run(ctx context.Context, pipeline *pipeline.Pipeline) error
	Stop(ctx context.Context) error

	Pause(ctx context.Context, pipeline string)
	Resume(ctx context.Context, pipeline string)

	// State provide information about driver state
	State(ctx context.Context) (*State, error)
}

// Acknowledger provides queue specific item management
type Acknowledger interface {
	// Ack - acknowledge the Item after processing
	Ack() error

	// Nack - discard the Item
	Nack() error

	// Requeue - put the message back to the queue with the optional delay
	Requeue(headers map[string][]string, delay int64) error

	// Respond to the queue
	Respond(payload []byte, queue string) error
}

// Constructor constructs Consumer interface. Endure abstraction.
type Constructor interface {
	ConsumerFromConfig(configKey string, queue priorityqueue.Queue) (Consumer, error)
	ConsumerFromPipeline(pipe *pipeline.Pipeline, queue priorityqueue.Queue) (Consumer, error)
}
