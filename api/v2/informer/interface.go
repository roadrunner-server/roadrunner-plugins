package informer

import (
	"context"

	"github.com/spiral/roadrunner-plugins/v2/api/v2/jobs"
	"github.com/spiral/roadrunner/v2/state/process"
)

// Statistic interfaces ==============

// Informer used to get workers from particular plugin or set of plugins
type Informer interface {
	Workers() []*process.State
}

// JobsStat interface provide statistic for the jobs plugin
type JobsStat interface {
	// JobsState returns slice with the attached drivers information
	JobsState(ctx context.Context) ([]*jobs.State, error)
}
