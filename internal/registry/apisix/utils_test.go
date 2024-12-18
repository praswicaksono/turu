package apisix_test

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/praswicaksono/turu/internal/registry/apisix"
	"github.com/stretchr/testify/assert"
)

type TestAssertion struct {
	data        func() any
	expectation func(any, error)
}

type TestTable struct {
	assertion map[string]TestAssertion
	test      func(any) (any, error)
}

func TestCreateRoute(t *testing.T) {
	table := TestTable{
		test: func(data any) (any, error) {
			return apisix.CreateRoute(data.(types.ContainerJSON))
		},
		assertion: map[string]TestAssertion{
			"missing_label": {
				data: func() any {
					return types.ContainerJSON{
						ContainerJSONBase: &types.ContainerJSONBase{
							ID:   "test-container",
							Name: "/test-service",
						},
						Config: &container.Config{
							Labels: map[string]string{
								"turu.apisix.uri": "/api",
							},
							ExposedPorts: nat.PortSet{
								nat.Port("80/tcp"): struct{}{},
							},
						},
					}
				},
				expectation: func(obj any, err error) {
					assert.Error(t, err)
					assert.Nil(t, obj)
				},
			},
			"valid": {
				data: func() any {
					return types.ContainerJSON{
						ContainerJSONBase: &types.ContainerJSONBase{
							ID:   "test-container",
							Name: "/test-service",
						},
						Config: &container.Config{
							Labels: map[string]string{
								"turu.apisix.enabled": "true",
								"turu.apisix.host":    "api.example.com",
								"turu.apisix.uri":     "/api",
							},
							ExposedPorts: nat.PortSet{
								nat.Port("80/tcp"): struct{}{},
							},
						},
					}
				},
				expectation: func(obj any, err error) {
					route := obj.(*apisix.Route)
					assert.NoError(t, err)
					assert.NotNil(t, route)
					assert.Equal(t, "test-service", route.ID)
					assert.Equal(t, "test-service", route.Name)
					assert.Equal(t, "api.example.com", route.Host)
					assert.Equal(t, "/api", route.URI)
					assert.Equal(t, 1, len(route.Upstream.Nodes.(map[string]any)))
					assert.Equal(t, "roundrobin", route.Upstream.Type)
				},
			},
		},
	}

	for k, v := range table.assertion {
		t.Run(k, func(t *testing.T) {
			v.expectation(table.test(v.data()))
		})
	}
}

func TestExtractLabelWithPrefix(t *testing.T) {
	table := TestTable{
		test: func(data any) (any, error) {
			return apisix.ExtractLabel(data.(types.ContainerJSON)), nil
		},
		assertion: map[string]TestAssertion{
			"valid": {
				data: func() any {
					return types.ContainerJSON{
						Config: &container.Config{
							Labels: map[string]string{
								"turu.apisix.enabled": "true",
								"turu.apisix.host":    "api.example.com",
								"other.label":         "value",
							},
						},
					}
				},
				expectation: func(obj any, err error) {
					labels := obj.(apisix.ApisixLabel)
					assert.NoError(t, err)
					assert.Equal(t, 2, len(labels))
					assert.Equal(t, "true", labels["turu.apisix.enabled"])
					assert.Equal(t, "api.example.com", labels["turu.apisix.host"])
					assert.NotContains(t, labels, "other.label")
				},
			},
		},
	}

	for k, v := range table.assertion {
		t.Run(k, func(t *testing.T) {
			v.expectation(table.test(v.data()))
		})
	}
}
