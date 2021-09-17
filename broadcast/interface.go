package broadcast

import "github.com/spiral/roadrunner-plugins/v2/internal/common/pubsub"

type Broadcaster interface {
	GetDriver(key string) (pubsub.SubReader, error)
}
