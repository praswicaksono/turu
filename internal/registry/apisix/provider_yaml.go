package apisix

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/goccy/go-yaml"
	"github.com/gookit/goutil"
	"github.com/gookit/goutil/maputil"
	"github.com/praswicaksono/turu/internal/conf"
	"github.com/praswicaksono/turu/internal/docker"
)

type RegistryYaml struct {
	m *sync.Mutex
}

func (p *RegistryYaml) Construct(ctx context.Context) {
	if p.m == nil {
		p.m = &sync.Mutex{}
	}
}

func (p *RegistryYaml) readConfig(path string) (*Config, error) {
	var cfg Config

	if path == "" {
		return nil, errors.New("apisix-yaml.path could not be empty")
	}

	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (p *RegistryYaml) writeConfig(path string, cfg *Config) error {
	s, err := yaml.MarshalWithOptions(cfg)

	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(fmt.Sprintf("%s\n\n#END", s)), 0766)
}

func (p *RegistryYaml) Register(ctx context.Context, c types.ContainerJSON) error {
	p.m.Lock()
	defer p.m.Unlock()

	path := conf.TuruConfig.Config.ApisixYaml.Path

	cfg, err := p.readConfig(path)

	if err != nil {
		return err
	}

	r, err := CreateRoute(c)
	if err != nil {
		return err
	}

	rs := cfg.Routes[:0]

	// loop current routes and merge node if exist
	for _, v := range cfg.Routes {
		if v.ID == r.ID {
			mn := maputil.Merge1level(v.Upstream.Nodes.(map[string]any), r.Upstream.Nodes.(map[string]any))
			v.Upstream.Nodes = mn
			rs = append(rs, v)
		} else {
			rs = append(rs, *r)
		}
	}

	if len(rs) == 0 {
		rs = append(rs, *r)
	}

	cfg.Routes = rs

	return p.writeConfig(path, cfg)
}

func (p *RegistryYaml) Deregister(ctx context.Context, c types.ContainerJSON) error {
	p.m.Lock()
	defer p.m.Unlock()

	path := conf.TuruConfig.Config.ApisixYaml.Path

	cfg, err := p.readConfig(path)

	if err != nil {
		return err
	}

	name, _ := docker.GetContainerOrServiceName(c)
	lb := docker.GetLoadBalancerURL(name, c)

	rs := cfg.Routes[:0]

	// search node, if found exclude from node list
	for _, x := range cfg.Routes {
		var nodes = make(map[string]int)
		currNodes := x.Upstream.Nodes.(map[string]any)
		for k := range currNodes {
			if !goutil.Contains(lb, k) {
				nodes[k] = 1
			}
		}

		if len(nodes) > 0 {
			x.Upstream.Nodes = nodes
			rs = append(rs, x)
		}
	}

	cfg.Routes = rs

	return p.writeConfig(path, cfg)
}
