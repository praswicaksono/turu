package docker

import (
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
)

type Docker struct {
	DockerManager *client.Client
}

func NewClientWithOpts(ops ...client.Opt) *Docker {
	cli, err := client.NewClientWithOpts(ops...)
	if err != nil {
		log.Fatal().Err(err).Msg(err.Error())
	}

	return &Docker{
		DockerManager: cli,
	}
}

func (d *Docker) Close() {
	d.DockerManager.Close()
}
