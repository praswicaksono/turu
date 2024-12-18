package docker

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/gookit/goutil"
)

type LoadBalancerURL []string

func GetContainerOrServiceName(cnt types.ContainerJSON) (string, string) {
	var (
		name    string
		service string
	)
	if goutil.Contains(cnt.Config.Labels, "com.docker.compose.project") {
		name = fmt.Sprintf(
			"%s-%s",
			cnt.Config.Labels["com.docker.compose.project"],
			cnt.Config.Labels["com.docker.compose.service"],
		)
		service = name
	} else {
		name = strings.Replace(cnt.Name, "/", "", 1)
		service = name
		if goutil.Contains(cnt.Config.Labels, "turu.service") {
			service = cnt.Config.Labels["turu.service"]
		}
	}

	return name, service
}

func GetRegistry(cnt types.ContainerJSON) string {
	if goutil.Contains(cnt.Config.Labels, "turu.registry") {
		return cnt.Config.Labels["turu.registry"]
	}

	return ""
}

func GetLoadBalancerURL(name string, cnt types.ContainerJSON) LoadBalancerURL {

	lb := make(LoadBalancerURL, len(cnt.Config.ExposedPorts))
	i := 0
	for k := range cnt.Config.ExposedPorts {
		lb = append(lb[:i], fmt.Sprintf("%s:%s", name, k.Port()))
		i++
	}

	return lb
}
