package config

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/spf13/viper"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config/configs/v2_6"
	"github.com/spiral/roadrunner-plugins/v2/config/configs/v2_7"
)

const (
	v26 string = "2.6.0"
	v27 string = "2.7.0"
)

const (
	jobsKey      string = "jobs"
	pipelinesKey string = "pipelines"
	configKey    string = "config"
)

const (
	amqpDriverName      string = "amqp"
	beanstalkDriverName string = "beanstalk"
	boltdbDriverName    string = "boltdb"
	memoryDriverName    string = "memory"
	natsDriverName      string = "nats"
	sqsDriverName       string = "sqs"
)

// transition used to upgrade configuration.
// configuration can be upgraded between related versions (2.5 -> 2.6, but not 2.5 -> 2.7).
func transition(from, to *version.Version, v *viper.Viper) error {
	segFrom := from.Segments64()
	segTo := to.Segments64()

	if (len(segTo) < 3 || len(segFrom) < 3) || (segTo[1]-segFrom[1]) != 1 {
		return errors.Errorf("incompatible versions passed: from: %s, to: %s", from.String(), to.String())
	}

	// use only 2 digits
	trFrom := fmt.Sprintf("%d.%d.0", segFrom[0], segFrom[1])
	trTo := fmt.Sprintf("%d.%d.0", segTo[0], segTo[1])

	switch trFrom {
	case v26:
		switch trTo { //nolint:gocritic
		case v27:
			// transition configuration from v2.6 to v2.7
			err := v26to27(v)
			if err != nil {
				return err
			}
		}
	case v27:
		return nil
	}

	return nil
}

func v26to27(v *viper.Viper) error {
	const op = errors.Op("v2.6_to_v2.7_transition")
	from := &v2_6.Config{}

	defer func() {
		from = nil
	}()

	err := v.UnmarshalKey(jobsKey, &from.Jobs)
	if err != nil {
		return errors.E(op, err)
	}

	// The user don't use a jobs, skip configuration convert
	if from.Jobs == nil {
		return nil
	}

	// iterate over old styled pipelines and fill the new configuration
	for key, val := range from.Jobs.Pipelines {
		dr := val.Driver()
		newConfigKey := fmt.Sprintf("%s.%s.%s.%s", jobsKey, pipelinesKey, key, configKey)
		oldConfigKey := fmt.Sprintf("%s.%s.%s", jobsKey, pipelinesKey, key)

		switch dr {
		case amqpDriverName:
			amqpConf := &v2_6.AMQPConfig{}
			err = v.UnmarshalKey(oldConfigKey, amqpConf)
			if err != nil {
				return errors.E(op, err)
			}

			conf27 := &v2_7.AMQPConfig{
				Priority:      amqpConf.Priority,
				Prefetch:      amqpConf.Prefetch,
				Queue:         amqpConf.Queue,
				Exchange:      amqpConf.Exchange,
				ExchangeType:  amqpConf.ExchangeType,
				RoutingKey:    amqpConf.RoutingKey,
				Exclusive:     amqpConf.Exclusive,
				MultipleAck:   amqpConf.MultipleAck,
				RequeueOnFail: amqpConf.RequeueOnFail,
			}

			v.Set(newConfigKey, conf27)
		case beanstalkDriverName:
			bsConf := &v2_6.BeanstalkConfig{}
			err = v.UnmarshalKey(oldConfigKey, bsConf)
			if err != nil {
				return errors.E(op, err)
			}

			conf27 := &v2_7.BeanstalkConfig{
				Priority:       bsConf.Priority,
				TubePriority:   bsConf.TubePriority,
				Tube:           bsConf.Tube,
				ReserveTimeout: bsConf.ReserveTimeout,
			}

			v.Set(newConfigKey, conf27)
		case sqsDriverName:
			sqsConf := &v2_6.SQSConfig{}
			err = v.UnmarshalKey(oldConfigKey, sqsConf)
			if err != nil {
				return errors.E(op, err)
			}

			conf27 := &v2_7.SQSConfig{
				Priority:          sqsConf.Priority,
				VisibilityTimeout: sqsConf.VisibilityTimeout,
				WaitTimeSeconds:   sqsConf.WaitTimeSeconds,
				Prefetch:          sqsConf.Prefetch,
				Queue:             sqsConf.Queue,
				Attributes:        sqsConf.Attributes,
				Tags:              sqsConf.Tags,
			}

			v.Set(newConfigKey, conf27)
		case memoryDriverName:
			memConf := &v2_6.MemoryConfig{}
			err = v.UnmarshalKey(oldConfigKey, memConf)
			if err != nil {
				return errors.E(op, err)
			}

			conf27 := &v2_7.MemoryConfig{
				Priority: memConf.Priority,
			}

			v.Set(newConfigKey, conf27)
		case boltdbDriverName:
			boltConf := &v2_6.BoltDBConfig{}
			err = v.UnmarshalKey(oldConfigKey, boltConf)
			if err != nil {
				return errors.E(op, err)
			}

			conf27 := &v2_7.BoltDBConfig{
				Priority: boltConf.Priority,
				File:     boltConf.File,
				Prefetch: boltConf.Prefetch,
			}

			v.Set(newConfigKey, conf27)
		case natsDriverName:
			natsConf := &v2_6.NATSConfig{}
			err = v.UnmarshalKey(oldConfigKey, natsConf)
			if err != nil {
				return errors.E(op, err)
			}

			conf27 := &v2_7.NATSConfig{
				Priority:           natsConf.Priority,
				Subject:            natsConf.Subject,
				Stream:             natsConf.Stream,
				Prefetch:           natsConf.Prefetch,
				RateLimit:          natsConf.RateLimit,
				DeleteAfterAck:     natsConf.DeleteAfterAck,
				DeliverNew:         natsConf.DeliverNew,
				DeleteStreamOnStop: natsConf.DeleteStreamOnStop,
			}

			v.Set(newConfigKey, conf27)
		default:
			return errors.E(op, errors.Errorf("unknown driver name: %s", dr))
		}
	}

	return nil
}
