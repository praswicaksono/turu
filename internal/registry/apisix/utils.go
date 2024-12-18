package apisix

import (
	"errors"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/gookit/goutil"
	"github.com/praswicaksono/turu/internal/docker"
)

var (
	LABEL_TURU_APISIX_URI  = "turu.apisix.uri"
	LABEL_TURU_APISIX_HOST = "turu.apisix.host"
)

type ApisixLabel map[string]string

func ExtractLabel(cnt types.ContainerJSON) ApisixLabel {
	labels := make(ApisixLabel)

	for k, v := range cnt.Config.Labels {
		if strings.HasPrefix(k, "turu.apisix.") {
			labels[k] = v
		}
	}

	return labels
}

func IsApisixEnabled(cnt types.ContainerJSON) bool {
	if !goutil.Contains(cnt.Config.Labels, LABEL_TURU_APISIX_URI) {
		return false
	}

	if !goutil.Contains(cnt.Config.Labels, LABEL_TURU_APISIX_HOST) {
		return false
	}

	return true
}

func CreateRoute(cnt types.ContainerJSON) (*Route, error) {
	apisixLabels := ExtractLabel(cnt)

	if !IsApisixEnabled(cnt) {
		return nil, errors.New("apisix not enabled")
	}

	name, service := docker.GetContainerOrServiceName(cnt)

	targetHostRule := apisixLabels[LABEL_TURU_APISIX_HOST]
	targetURIRule := "/*"

	if goutil.Contains(apisixLabels, LABEL_TURU_APISIX_URI) {
		targetURIRule = apisixLabels[LABEL_TURU_APISIX_URI]
	}

	lb := docker.GetLoadBalancerURL(name, cnt)
	nodes := map[string]any{}
	for _, v := range lb {
		if v == "" {
			continue
		}
		nodes[v] = 1
	}

	r := &Route{
		BaseInfo: BaseInfo{
			ID: service,
		},
		Name: service,
		URI:  targetURIRule,
		Host: targetHostRule,
		Upstream: &UpstreamDef{
			Nodes: nodes,
			Type:  "roundrobin",
		},
		Status: 1,
	}
	r.Creating()

	return r, nil
}
