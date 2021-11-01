package broadcast

import "github.com/spiral/roadrunner-plugins/v2/api/pubsub"

type Broadcaster interface {
	GetDriver(key string) (pubsub.SubReader, error)
}
