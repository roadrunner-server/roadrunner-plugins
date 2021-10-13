package jobs

import (
	"time"

	"github.com/spiral/roadrunner/v2/events"
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

					start := time.Now()
					p.events.Push(events.JobEvent{
						Event:   events.EventJobStart,
						ID:      jb.ID(),
						Start:   start,
						Elapsed: 0,
					})

					ctx, err := jb.Context()
					if err != nil {
						p.events.Push(events.JobEvent{
							Event:   events.EventJobError,
							Error:   err,
							ID:      jb.ID(),
							Start:   start,
							Elapsed: time.Since(start),
						})

						p.log.Error("job marshal context", "error", err)

						errNack := jb.Nack()
						if errNack != nil {
							p.log.Error("negatively acknowledge failed", "error", errNack)
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
						p.events.Push(events.JobEvent{
							Event:   events.EventJobError,
							ID:      jb.ID(),
							Error:   err,
							Start:   start,
							Elapsed: time.Since(start),
						})
						// RR protocol level error, Nack the job
						errNack := jb.Nack()
						if errNack != nil {
							p.log.Error("negatively acknowledge failed", "error", errNack)
						}

						p.log.Error("job execute failed", "error", err)
						p.putPayload(exec)
						jb = nil
						continue
					}

					// if response is nil or body is nil, just acknowledge the job
					if resp == nil || resp.Body == nil {
						p.putPayload(exec)
						err = jb.Ack()
						if err != nil {
							p.events.Push(events.JobEvent{
								Event:   events.EventJobError,
								ID:      jb.ID(),
								Error:   err,
								Start:   start,
								Elapsed: time.Since(start),
							})
							p.log.Error("acknowledge error, job might be missed", "error", err)
							jb = nil
							continue
						}

						p.events.Push(events.JobEvent{
							Event:   events.EventJobOK,
							ID:      jb.ID(),
							Start:   start,
							Elapsed: time.Since(start),
						})

						jb = nil
						continue
					}

					// handle the response protocol
					err = p.respHandler.Handle(resp.Body, jb)
					if err != nil {
						p.events.Push(events.JobEvent{
							Event:   events.EventJobError,
							ID:      jb.ID(),
							Start:   start,
							Error:   err,
							Elapsed: time.Since(start),
						})
						p.putPayload(exec)
						errNack := jb.Nack()
						if errNack != nil {
							p.log.Error("negatively acknowledge failed, job might be lost", "root error", err, "error nack", errNack)
							jb = nil
							continue
						}

						p.log.Error("job negatively acknowledged", "error", err)
						jb = nil
						continue
					}

					p.events.Push(events.JobEvent{
						Event:   events.EventJobOK,
						ID:      jb.ID(),
						Start:   start,
						Elapsed: time.Since(start),
					})

					// return payload
					p.putPayload(exec)
					jb = nil
				}
			}
		}()
	}
}
