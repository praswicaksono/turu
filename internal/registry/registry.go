package registry

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types"
	"github.com/gookit/goutil"
	"github.com/praswicaksono/turu/internal/docker"
	"github.com/praswicaksono/turu/internal/registry/apisix"
	"github.com/rs/zerolog/log"
)

type RegistryCollection = map[string]Registry

var availableRegistry = RegistryCollection{
	"apisix-yaml": &apisix.RegistryYaml{},
	"apisix-etcd": &apisix.RegistryEtcd{},
}

type Registry interface {
	Register(ctx context.Context, c types.ContainerJSON) error
	Deregister(ctx context.Context, c types.ContainerJSON) error
	Construct(ctx context.Context)
}

func isValidRegistry(ctx context.Context, p string) error {
	if !goutil.Contains(availableRegistry, p) {
		err := errors.New("invalid turu registry, skiping registation")
		log.Ctx(ctx).Warn().Msg(err.Error())
		return err
	}

	return nil
}
func HandleContainerCreateEvent(ctx context.Context, cnt types.ContainerJSON) error {
	p := docker.GetRegistry(cnt)

	// if registry label not found skip then process
	err := isValidRegistry(ctx, p)
	if err != nil {
		return nil
	}

	availableRegistry[p].Construct(ctx)
	err = availableRegistry[p].Register(ctx, cnt)
	if err != nil {
		return err
	}

	log.Ctx(ctx).Info().Msg("container successfully registered")
	return nil
}

func HandleContainerKillEvent(ctx context.Context, cnt types.ContainerJSON) error {
	p := docker.GetRegistry(cnt)

	// if registry label not found skip then process
	err := isValidRegistry(ctx, p)
	if err != nil {
		return nil
	}

	availableRegistry[p].Construct(ctx)
	err = availableRegistry[p].Deregister(ctx, cnt)
	if err != nil {
		return err
	}

	log.Ctx(ctx).Info().Msg("container successfully deregistered")
	return nil
}
