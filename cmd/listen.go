package cmd

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/praswicaksono/turu/internal/docker"
	"github.com/praswicaksono/turu/internal/registry"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen docker event and register to service discovery",
	Run: func(cmd *cobra.Command, args []string) {
		client := docker.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		defer client.Close()

		ctx := context.Background()

		client.ListenForDockerEvent(
			ctx,
			events.ListOptions{
				Filters: filters.NewArgs(
					filters.KeyValuePair{Key: "type", Value: "container"},
					filters.KeyValuePair{Key: "event", Value: "start"},
					filters.KeyValuePair{Key: "event", Value: "kill"},
				),
			},
			func(ctx context.Context, cancelFunc context.CancelFunc, event events.Message) {
				res, err := client.DockerManager.ContainerInspect(ctx, event.Actor.ID)
				if err != nil {
					log.Error().Err(err).Stack().Msg("")
					return
				}

				ctx = log.With().
					Str("event", fmt.Sprintf("%s-%s", string(event.Type), event.Action)).
					Str("container_id", res.ID).
					Str("name", res.Name).
					Logger().WithContext(ctx)

				switch event.Action {
				case "start":
					err = registry.HandleContainerCreateEvent(ctx, res)
				case "kill":
					err = registry.HandleContainerKillEvent(ctx, res)
				}

				if err != nil {
					log.Ctx(ctx).Error().Stack().Err(err).Msg("")
				}
			},
		)
	},
}
