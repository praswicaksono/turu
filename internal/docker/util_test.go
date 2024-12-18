package docker_test

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/praswicaksono/turu/internal/docker"
	"github.com/stretchr/testify/assert"
)

// Container with compose labels returns swarm mode with project-service name format
func TestGetContainerNameWithComposeLabels(t *testing.T) {
	container := types.ContainerJSON{
		Config: &container.Config{
			Labels: map[string]string{
				"com.docker.compose.project": "myproject",
				"com.docker.compose.service": "webapp",
			},
		},
		ContainerJSONBase: &types.ContainerJSONBase{
			Name: "/container-name",
		},
	}

	name, service := docker.GetContainerOrServiceName(container)

	assert.Equal(t, "myproject-webapp", name)
	assert.Equal(t, "myproject-webapp", service)
}

// Container with empty compose project label
func TestGetContainerNameWithoutComposeLabels(t *testing.T) {
	container := types.ContainerJSON{
		Config: &container.Config{
			Labels: map[string]string{
				"turu.service": "myservice",
			},
		},
		ContainerJSONBase: &types.ContainerJSONBase{
			Name: "/container-name",
		},
	}

	name, service := docker.GetContainerOrServiceName(container)

	assert.Equal(t, "container-name", name)
	assert.Equal(t, "myservice", service)
}

// Container without turu.service label returns empty service string
func TestGetContainerNameWithoutTuruServiceLabel(t *testing.T) {
	container := types.ContainerJSON{
		Config: &container.Config{
			Labels: map[string]string{},
		},
		ContainerJSONBase: &types.ContainerJSONBase{
			Name: "/container-name",
		},
	}

	name, service := docker.GetContainerOrServiceName(container)

	assert.Equal(t, "container-name", name)
	assert.Equal(t, "container-name", service)
}

// Container with single exposed port returns URL with correct name and port
func TestGetLoadBalancerURLWithSinglePort(t *testing.T) {
	containerName := "test-container"

	container := types.ContainerJSON{
		Config: &container.Config{
			ExposedPorts: nat.PortSet{
				"8080/tcp": struct{}{},
			},
		},
	}

	result := docker.GetLoadBalancerURL(containerName, container)

	assert.Equal(t, 1, len(result))

	ex := docker.LoadBalancerURL{"test-container:8080"}
	for k, v := range result {
		assert.Equal(t, ex[k], v)
	}
}

func TestGetRegistryReturnsLabelValue(t *testing.T) {
	container := types.ContainerJSON{
		Config: &container.Config{
			Labels: map[string]string{
				"turu.registry": "apisix",
			},
		},
	}

	r := docker.GetRegistry(container)

	assert.Equal(t, "apisix", r)
}
