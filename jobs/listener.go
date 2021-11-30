package jobs

import (
	"sync/atomic"
	"time"

	"github.com/spiral/roadrunner-plugins/v2/api/jobs"
)

func (p *Plugin) listener() { //nolint:gocognit
	for i := uint8(0); i < p.cfg.NumPollers; i++ {
		go func() {
			for {
				select {
				case <-p.stopCh:
					p.log.Debug("------> job poller stopped <------")
					return
				default:
					start := time.Now()
					// get prioritized JOB from the queue
					jb := p.queue.ExtractMin()

					// parse the context
					// for each job, context contains:
					/*
						1. Job class
						2. Job ID provided from the outside
						3. Job Headers map[string][]string
						4. Timeout in seconds
						5. Pipeline name
					*/

					p.log.Debug("job processing started", "ID", jb.ID(), "start", start, "elapsed", time.Since(start))

					ctx, err := jb.Context()
					if err != nil {
						atomic.AddUint64(p.metrics.jobsErr, 1)
						p.log.Error("job marshal error", "error", err, "ID", jb.ID(), "start", start, "elapsed", time.Since(start))

						errNack := jb.(jobs.Acknowledger).Nack()
						if errNack != nil {
							p.log.Error("negatively acknowledge failed", "ID", jb.ID(), "error", errNack)
						}
						continue
					}

					// get payload from the sync.Pool
					exec := p.getPayload(jb.Body(), ctx)

					// protect from the pool reset
					p.RLock()
					resp, err := p.workersPool.Exec(exec)
					p.RUnlock()
					if err != nil {
						atomic.AddUint64(p.metrics.jobsErr, 1)
						p.log.Error("job processed with errors", "error", err, "ID", jb.ID(), "start", start, "elapsed", time.Since(start))
						if _, ok := jb.(jobs.Acknowledger); !ok {
							p.log.Error("job execute failed, job is not a Acknowledger, skipping Ack/Nack")
							p.putPayload(exec)
							continue
						}
						// RR protocol level error, Nack the job
						errNack := jb.(jobs.Acknowledger).Nack()
						if errNack != nil {
							p.log.Error("negatively acknowledge failed", "ID", jb.ID(), "error", errNack)
						}

						p.log.Error("job execute failed", "error", err)
						p.putPayload(exec)
						jb = nil
						continue
					}

					if _, ok := jb.(jobs.Acknowledger); !ok {
						// can't acknowledge, just continue
						p.putPayload(exec)
						continue
					}

					// if response is nil or body is nil, just acknowledge the job
					if resp == nil || resp.Body == nil {
						p.putPayload(exec)
						err = jb.(jobs.Acknowledger).Ack()
						if err != nil {
							atomic.AddUint64(p.metrics.jobsErr, 1)
							p.log.Error("acknowledge error, job might be missed", "error", err, "ID", jb.ID(), "start", start, "elapsed", time.Since(start))
							jb = nil
							continue
						}

						p.log.Debug("job processed successfully", "ID", jb.ID(), "start", start, "elapsed", time.Since(start))
						// metrics
						atomic.AddUint64(p.metrics.jobsOk, 1)

						jb = nil
						continue
					}

					// handle the response protocol
					err = p.respHandler.Handle(resp, jb.(jobs.Acknowledger))
					if err != nil {
						atomic.AddUint64(p.metrics.jobsErr, 1)
						p.log.Error("response handler error", "error", err, "ID", jb.ID(), "start", start, "elapsed", time.Since(start))
						p.putPayload(exec)
						errNack := jb.(jobs.Acknowledger).Nack()
						if errNack != nil {
							p.log.Error("negatively acknowledge failed, job might be lost", "ID", jb.ID(), "root error", err, "error nack", errNack)
							jb = nil
							continue
						}

						p.log.Error("job negatively acknowledged", "error", err)
						jb = nil
						continue
					}

					// metrics
					atomic.AddUint64(p.metrics.jobsOk, 1)

					p.log.Debug("job processed successfully", "ID", jb.ID(), "start", start, "elapsed", time.Since(start))

					// return payload
					p.putPayload(exec)
					jb = nil
				}
			}
		}()
	}
}
