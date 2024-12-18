package apisix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/gookit/goutil"
	"github.com/gookit/goutil/maputil"
	"github.com/praswicaksono/turu/internal/conf"
	"github.com/praswicaksono/turu/internal/docker"
	"github.com/rs/zerolog/log"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type RegistryEtcd struct {
	ec *clientv3.Client
	c  *concurrency.Session
}

func (p *RegistryEtcd) createEtcdClient() *clientv3.Client {
	cnf := clientv3.Config{
		Endpoints:   conf.TuruConfig.Config.ApisixEtcd.Endpoint,
		DialTimeout: conf.TuruConfig.Config.ApisixEtcd.Timeout,
	}

	if conf.TuruConfig.Config.ApisixEtcd.Username != nil && conf.TuruConfig.Config.ApisixEtcd.Password != nil {
		cnf.Username = *conf.TuruConfig.Config.ApisixEtcd.Username
		cnf.Password = *conf.TuruConfig.Config.ApisixEtcd.Password
	}

	if conf.TuruConfig.Config.ApisixEtcd.MTLS != nil {
		if conf.TuruConfig.Config.ApisixEtcd.MTLS.CA != "" && conf.TuruConfig.Config.ApisixEtcd.MTLS.Cert != "" && conf.TuruConfig.Config.ApisixEtcd.MTLS.Key != "" {
			tlsInfo := transport.TLSInfo{
				CertFile:      conf.TuruConfig.Config.ApisixEtcd.MTLS.Cert,
				KeyFile:       conf.TuruConfig.Config.ApisixEtcd.MTLS.Key,
				TrustedCAFile: conf.TuruConfig.Config.ApisixEtcd.MTLS.CA,
			}
			tlsConfig, err := tlsInfo.ClientConfig()
			if err != nil {
				log.Fatal().Err(err).Msg("")
			}

			cnf.TLS = tlsConfig
		}
	}

	cli, err := clientv3.New(cnf)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	return cli
}

func (p *RegistryEtcd) lock(name string) (*concurrency.Mutex, error) {
	var err error
	p.c, err = concurrency.NewSession(p.ec)
	if err != nil {
		return nil, err
	}

	return concurrency.NewMutex(p.c, name), nil
}

func (p *RegistryEtcd) Construct(ctx context.Context) {
	if p.ec == nil {
		p.ec = p.createEtcdClient()
	}
}

func (p *RegistryEtcd) Register(ctx context.Context, c types.ContainerJSON) error {
	_, servicename := docker.GetContainerOrServiceName(c)

	lock, err := p.lock(fmt.Sprintf("/turu-apisix-etcd-create-%s/", servicename))
	if err != nil {
		return err
	}
	defer p.c.Close()

	// check if lock already by another process, skip if already locked
	err = lock.TryLock(ctx)
	if err != nil {
		return err
	}

	key := "/apisix/routes/" + servicename

	res, err := p.ec.Get(ctx, key)
	if err != nil {
		return err
	}

	newRoute, err := CreateRoute(c)
	if err != nil {
		return err
	}

	// add new route if not exist
	if res.Count == 0 {
		j, err := json.Marshal(newRoute)
		if err != nil {
			return err
		}
		_, err = p.ec.Put(ctx, key, string(j))
		if err != nil {
			return err
		}
		return nil
	}

	// if route exist, merge upsteam nodes
	body := res.Kvs[0].Value

	var currentRoute Route
	err = json.Unmarshal(body, &currentRoute)
	if err != nil {
		return err
	}

	currentRoute.Upstream.Nodes = maputil.Merge1level(currentRoute.Upstream.Nodes.(map[string]any), newRoute.Upstream.Nodes.(map[string]any))

	j, err := json.Marshal(currentRoute)
	if err != nil {
		return err
	}
	_, err = p.ec.Put(ctx, key, string(j))
	if err != nil {
		return err
	}

	return nil
}

func (p *RegistryEtcd) Deregister(ctx context.Context, c types.ContainerJSON) error {
	name, servicename := docker.GetContainerOrServiceName(c)

	lock, err := p.lock(fmt.Sprintf("/turu-apisix-etcd-remove-%s/", servicename))
	if err != nil {
		return err
	}
	defer p.c.Close()

	// check if lock already by another process, skip if already locked
	err = lock.TryLock(ctx)
	if err != nil {
		return err
	}

	key := "/apisix/routes/" + servicename

	res, err := p.ec.Get(ctx, key)
	if err != nil {
		return err
	}

	if res.Count == 0 {
		return errors.New("route not found, nothing deregistered")
	}

	body := res.Kvs[0].Value

	var currentRoute Route
	err = json.Unmarshal(body, &currentRoute)
	if err != nil {
		return err
	}

	lb := docker.GetLoadBalancerURL(name, c)

	var nodes = make(map[string]int)
	currNodes := currentRoute.Upstream.Nodes.(map[string]any)
	for k := range currNodes {
		if !goutil.Contains(lb, k) {
			nodes[k] = 1
		}
	}

	// if there is no node left, delete the route
	if len(nodes) == 0 {
		_, err = p.ec.Delete(ctx, key)
		if err != nil {
			return err
		}
		return nil
	}

	// otherwise, update the route with new node list
	currentRoute.Upstream.Nodes = nodes

	j, err := json.Marshal(currentRoute)
	if err != nil {
		return err
	}

	_, err = p.ec.Put(ctx, key, string(j))
	if err != nil {
		return err
	}

	return nil
}
