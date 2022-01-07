package sqsjobs

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	cfgPlugin "github.com/roadrunner-server/api/plugins/v2/config"
	"github.com/roadrunner-server/api/plugins/v2/jobs"
	"github.com/roadrunner-server/api/plugins/v2/jobs/pipeline"
	"github.com/spiral/errors"
	priorityqueue "github.com/spiral/roadrunner/v2/priority_queue"
	"go.uber.org/zap"
)

const (
	pluginName string = "sqs"
)

type consumer struct {
	sync.Mutex
	pq       priorityqueue.Queue
	log      *zap.Logger
	pipeline atomic.Value

	// connection info
	key               string
	secret            string
	sessionToken      string
	region            string
	endpoint          string
	queue             *string
	messageGroupID    string
	waitTime          int32
	prefetch          int32
	visibilityTimeout int32

	// if user invoke several resume operations
	listeners uint32

	// queue optional parameters
	attributes map[string]string
	tags       map[string]string

	client   *sqs.Client
	queueURL *string

	pauseCh chan struct{}
}

func NewSQSConsumer(configKey string, log *zap.Logger, cfg cfgPlugin.Configurer, pq priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("new_sqs_consumer")

	// if no such key - error
	if !cfg.Has(configKey) {
		return nil, errors.E(op, errors.Errorf("no configuration by provided key: %s", configKey))
	}

	// if no global section
	if !cfg.Has(pluginName) {
		return nil, errors.E(op, errors.Str("no global sqs configuration, global configuration should contain sqs section"))
	}

	// PARSE CONFIGURATION -------
	var conf Config
	err := cfg.UnmarshalKey(configKey, &conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	conf.InitDefault()

	err = cfg.UnmarshalKey(pluginName, &conf)
	if err != nil {
		return nil, errors.E(op, err)
	}

	conf.InitDefault()

	// initialize job consumer
	jb := &consumer{
		pq:                pq,
		log:               log,
		messageGroupID:    uuid.NewString(),
		attributes:        conf.Attributes,
		tags:              conf.Tags,
		queue:             conf.Queue,
		prefetch:          conf.Prefetch,
		visibilityTimeout: conf.VisibilityTimeout,
		waitTime:          conf.WaitTimeSeconds,
		region:            conf.Region,
		key:               conf.Key,
		sessionToken:      conf.SessionToken,
		secret:            conf.Secret,
		endpoint:          conf.Endpoint,
		pauseCh:           make(chan struct{}, 1),
	}

	// PARSE CONFIGURATION -------

	awsConf, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(conf.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(jb.key, jb.secret, jb.sessionToken)))
	if err != nil {
		return nil, errors.E(op, err)
	}

	// config with retries
	jb.client = sqs.NewFromConfig(awsConf, sqs.WithEndpointResolver(sqs.EndpointResolverFromURL(jb.endpoint)), func(o *sqs.Options) {
		o.Retryer = retry.NewStandard(func(opts *retry.StandardOptions) {
			opts.MaxAttempts = 60
		})
	})

	out, err := jb.client.CreateQueue(context.Background(), &sqs.CreateQueueInput{QueueName: jb.queue, Attributes: jb.attributes, Tags: jb.tags})
	if err != nil {
		return nil, errors.E(op, err)
	}

	// assign a queue URL
	jb.queueURL = out.QueueUrl

	// To successfully create a new queue, you must provide a
	// queue name that adheres to the limits related to queues
	// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/limits-queues.html)
	// and is unique within the scope of your queues. After you create a queue, you
	// must wait at least one second after the queue is created to be able to use the <------------
	// queue. To get the queue URL, use the GetQueueUrl action. GetQueueUrl require
	time.Sleep(time.Second * 2)

	return jb, nil
}

func FromPipeline(pipe *pipeline.Pipeline, log *zap.Logger, cfg cfgPlugin.Configurer, pq priorityqueue.Queue) (*consumer, error) {
	const op = errors.Op("new_sqs_consumer")

	// if no global section
	if !cfg.Has(pluginName) {
		return nil, errors.E(op, errors.Str("no global sqs configuration, global configuration should contain sqs section"))
	}

	// PARSE CONFIGURATION -------
	var conf Config
	err := cfg.UnmarshalKey(pluginName, &conf)
	if err != nil {
		return nil, errors.E(op, err)
	}
	conf.InitDefault()

	attr := make(map[string]string)
	err = pipe.Map(attributes, attr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	tg := make(map[string]string)
	err = pipe.Map(tags, tg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// initialize job consumer
	jb := &consumer{
		pq:                pq,
		log:               log,
		messageGroupID:    uuid.NewString(),
		attributes:        attr,
		tags:              tg,
		queue:             aws.String(pipe.String(queue, "default")),
		prefetch:          int32(pipe.Int(pref, 10)),
		visibilityTimeout: int32(pipe.Int(visibility, 0)),
		waitTime:          int32(pipe.Int(waitTime, 0)),
		region:            conf.Region,
		key:               conf.Key,
		sessionToken:      conf.SessionToken,
		secret:            conf.Secret,
		endpoint:          conf.Endpoint,
		pauseCh:           make(chan struct{}, 1),
	}

	// PARSE CONFIGURATION -------

	awsConf, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(conf.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(jb.key, jb.secret, jb.sessionToken)))
	if err != nil {
		return nil, errors.E(op, err)
	}

	// config with retries
	jb.client = sqs.NewFromConfig(awsConf, sqs.WithEndpointResolver(sqs.EndpointResolverFromURL(jb.endpoint)), func(o *sqs.Options) {
		o.Retryer = retry.NewStandard(func(opts *retry.StandardOptions) {
			opts.MaxAttempts = 60
		})
	})

	out, err := jb.client.CreateQueue(context.Background(), &sqs.CreateQueueInput{QueueName: jb.queue, Attributes: jb.attributes, Tags: jb.tags})
	if err != nil {
		return nil, errors.E(op, err)
	}

	// assign a queue URL
	jb.queueURL = out.QueueUrl

	// To successfully create a new queue, you must provide a
	// queue name that adheres to the limits related to queues
	// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/limits-queues.html)
	// and is unique within the scope of your queues. After you create a queue, you
	// must wait at least one second after the queue is created to be able to use the <------------
	// queue. To get the queue URL, use the GetQueueUrl action. GetQueueUrl require
	time.Sleep(time.Second * 2)

	return jb, nil
}

func (c *consumer) Push(ctx context.Context, jb *jobs.Job) error {
	const op = errors.Op("sqs_push")
	// check if the pipeline registered

	// load atomic value
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != jb.Options.Pipeline {
		return errors.E(op, errors.Errorf("no such pipeline: %s, actual: %s", jb.Options.Pipeline, pipe.Name()))
	}

	// The length of time, in seconds, for which to delay a specific message. Valid
	// values: 0 to 900. Maximum: 15 minutes.
	if jb.Options.Delay > 900 {
		return errors.E(op, errors.Errorf("unable to push, maximum possible delay is 900 seconds (15 minutes), provided: %d", jb.Options.Delay))
	}

	err := c.handleItem(ctx, fromJob(jb))
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (c *consumer) State(ctx context.Context) (*jobs.State, error) {
	const op = errors.Op("sqs_state")
	attr, err := c.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: c.queueURL,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
			types.QueueAttributeNameApproximateNumberOfMessagesDelayed,
			types.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
		},
	})

	if err != nil {
		return nil, errors.E(op, err)
	}

	pipe := c.pipeline.Load().(*pipeline.Pipeline)

	out := &jobs.State{
		Pipeline: pipe.Name(),
		Driver:   pipe.Driver(),
		Queue:    *c.queueURL,
		Ready:    ready(atomic.LoadUint32(&c.listeners)),
	}

	nom, err := strconv.Atoi(attr.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)])
	if err == nil {
		out.Active = int64(nom)
	}

	delayed, err := strconv.Atoi(attr.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessagesDelayed)])
	if err == nil {
		out.Delayed = int64(delayed)
	}

	nv, err := strconv.Atoi(attr.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessagesNotVisible)])
	if err == nil {
		out.Reserved = int64(nv)
	}

	return out, nil
}

func (c *consumer) Register(_ context.Context, p *pipeline.Pipeline) error {
	c.pipeline.Store(p)
	return nil
}

func (c *consumer) Run(_ context.Context, p *pipeline.Pipeline) error {
	start := time.Now()
	const op = errors.Op("sqs_run")

	c.Lock()
	defer c.Unlock()

	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p.Name() {
		return errors.E(op, errors.Errorf("no such pipeline registered: %s", pipe.Name()))
	}

	atomic.AddUint32(&c.listeners, 1)

	// start listener
	// TODO(rustatian) context with cancel to cancel receive operation on stop
	go c.listen(context.Background())

	c.log.Debug("pipeline is active", zap.String("driver", pipe.Driver()), zap.String("pipeline", pipe.Name()), zap.Time("start", start), zap.Duration("elapsed", time.Since(start)))
	return nil
}

func (c *consumer) Stop(context.Context) error {
	start := time.Now()
	if atomic.LoadUint32(&c.listeners) > 0 {
		c.pauseCh <- struct{}{}
	}

	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	c.log.Debug("pipeline was stopped", zap.String("driver", pipe.Driver()), zap.String("pipeline", pipe.Name()), zap.Time("start", time.Now()), zap.Duration("elapsed", time.Since(start)))
	return nil
}

func (c *consumer) Pause(_ context.Context, p string) {
	start := time.Now()
	// load atomic value
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", zap.String("requested", p), zap.String("actual", pipe.Name()))
		return
	}

	l := atomic.LoadUint32(&c.listeners)
	// no active listeners
	if l == 0 {
		c.log.Warn("no active listeners, nothing to pause")
		return
	}

	atomic.AddUint32(&c.listeners, ^uint32(0))

	// stop consume
	c.pauseCh <- struct{}{}
	c.log.Debug("pipeline was paused", zap.String("driver", pipe.Driver()), zap.String("pipeline", pipe.Name()), zap.Time("start", time.Now()), zap.Duration("elapsed", time.Since(start)))
}

func (c *consumer) Resume(_ context.Context, p string) {
	start := time.Now()
	// load atomic value
	pipe := c.pipeline.Load().(*pipeline.Pipeline)
	if pipe.Name() != p {
		c.log.Error("no such pipeline", zap.String("requested", p), zap.String("actual", pipe.Name()))
		return
	}

	l := atomic.LoadUint32(&c.listeners)
	// no active listeners
	if l == 1 {
		c.log.Warn("sqs listener already in the active state")
		return
	}

	// start listener
	go c.listen(context.Background())

	// increase num of listeners
	atomic.AddUint32(&c.listeners, 1)
	c.log.Debug("pipeline was resumed", zap.String("driver", pipe.Driver()), zap.String("pipeline", pipe.Name()), zap.Time("start", time.Now()), zap.Duration("elapsed", time.Since(start)))
}

func (c *consumer) handleItem(ctx context.Context, msg *Item) error {
	d, err := msg.pack(c.queueURL)
	if err != nil {
		return err
	}
	_, err = c.client.SendMessage(ctx, d)
	if err != nil {
		return err
	}

	return nil
}

func ready(r uint32) bool {
	return r > 0
}
