# 💤 Turu

Turu is simple service registrator for docker containers. It will listen docker container event and register service to selected service registry.

## Install

Build from source

```bash
git clone https://github.com/FiloSottile/mkcert && cd mkcert
go build -o turu
```

## Usage

Make sure docker already running and docker socket is accessible. If you intend to use turu on docker container make sure to mount docker socket to container.

```bash
./turu listen
```

## Configuration

Turu utilize golang viper to load configuration. It read configuration from environment variables with prefix `TURU_`. And load configuration from this following file:

- turu.yaml
- $HOME/turu.yaml
- /etc/turu/turu.yaml

## Supported Registry

Turu will read docker label to determine which registry to use. Currently turu support this following registry:

- apisix-yaml
- apisix-etcd

How to specify docker label

```bash
docker run -d -l turu.service=whoami -l turu.registry=apisix-etcd --name whoami-1 --network turu traefik/whoami
```

### - apisix-yaml configuration

`turu.yaml` configuration

```yaml
config:
  apisix-yaml:
    path: path-to-yaml-file
```

docker label configuration

```txt
turu.apisix.uri=/*
turu.apisix.host=example.com
```

example docker container

```bash
docker run -d -l turu.service=whoami -l turu.registry=apisix-yaml -l turu.apisix.uri=/* -l turu.apisix.host=example.com --name whoami-1 --network turu traefik/whoami
```

It also support docker-compose you need to put label in `compose.yaml`

Turu will automatically detect exposed port, you dont have to bind it to local port but make sure apisix and your container are in same network.

**NOTE**: For docker-compose it will registry only 1 service node, since load balance will be handled by docker compose.

### - apisix-etcd configuration

`turu.yaml` configuration

```yaml
config:
  apisix-etcd:
    endpoint: http://127.0.0.1:2379
    timeout: 5s
    username: optional
    password: optional
    mtls:
      cert: path-to-certificate
      key: path-to-key
      ca: path-to-ca
```

example docker container

```bash
docker run -d -l turu.service=whoami -l turu.registry=apisix-yaml -l turu.apisix.uri=/* -l turu.apisix.host=example.com --name whoami-1 --network turu traefik/whoami
```

It also support docker-compose you need to put label in `compose.yaml`

Turu will automatically detect exposed port, you dont have to bind it to local port but make sure apisix and your container are in same network.

**NOTE**: For docker-compose it will registry only 1 service node, since load balance will be handled by docker compose.
