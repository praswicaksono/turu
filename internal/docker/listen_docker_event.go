package docker

import (
	"context"

	"github.com/docker/docker/api/types/events"
	"github.com/rs/zerolog/log"
)

type DockerEventHandler func(ctx context.Context, cancelFunc context.CancelFunc, event events.Message)

func (d *Docker) ListenForDockerEvent(ctx context.Context, opt events.ListOptions, handler DockerEventHandler) error {
	eventCtx, eventCancel := context.WithCancel(ctx)

	go func() {
		msg, errs := d.DockerManager.Events(eventCtx, opt)
		for {
			select {
			case err := <-errs:
				log.Error().Err(err).Msg("")
				return
			case event := <-msg:
				go func() {
					handler(ctx, eventCancel, event)
				}()
			}
		}
	}()

	<-eventCtx.Done()

	return nil
}
