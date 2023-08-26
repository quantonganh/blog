---
title: plugins/docker failed to resolve Keycloak hostname?
date: Wed Jan 27 10:07:19 +07 2021
description:
categories:
    - DevOps
tags:
    - drone
    - ci
    - docker
    - dns
---
After integrating Docker registry with Keycloak, the publishing step failed to authenticate with Docker Registry.

The full error message is:

```
time="2021-01-26T13:44:18.485121053Z" level=error msg="Handler for POST /v1.40/auth returned error: Get https://docker.domain.com/v2/: Get https://sso.domain.com/auth/realms/application/protocol/docker-v2/auth?account=******&client_id=docker&offline_token=true&service=aws-docker-registry: dial tcp: lookup sso.domain.com on 127.0.0.11:53: no such host"
```

`sso.domain.com` is a local hostname which can be resolved on the host. How can I make it resolvable inside the `plugins/docker` container?

I found some similar issues:

- https://discourse.drone.io/t/dns-lookup-fails-inside-plugins-docker-build/501/5
- https://github.com/drone-plugins/drone-docker/issues/193

but they are slightly differences.

Look at this: http://plugins.drone.io/drone-plugins/drone-docker/

`custom_dns` will be passed to the Docker daemon inside `plugins/docker`, something like this:

```shell
/usr/local/bin/dockerd --data-root /var/lib/docker --host=unix:///var/run/docker.sock --dns 10.100.101.5 --dns 8.8.8.8
```

while `add_host` will be passed to the `docker build` steps:

```shell
$ docker build -h
Flag shorthand -h has been deprecated, please use --help

Usage:  docker build [OPTIONS] PATH | URL | -

Build an image from a Dockerfile

Options:
      --add-host list           Add a custom host-to-IP mapping (host:ip)
```

They are the latter steps. Our case failed early when authenticating with docker registry.

I tried to add [`network_mode: host`](https://docs.docker.com/compose/compose-file/compose-file-v3/#network_mode):

```yaml
- name: publish
  image: plugins/docker:19.03
  settings:
    debug: true
    network_mode: host
    registry: docker.domain.com
    repo: docker.domain.com/owner/repo
    tags:
    - ${DRONE_SOURCE_BRANCH}
    username:
      from_secret: docker_username
    password:
      from_secret: docker_password
```

but it didn't help, the error still stands:

```shell
dial tcp: lookup sso.domain.com on 127.0.0.11:53: no such host"
```

Why `plugins/docker` still uses [embedded DNS](https://docs.docker.com/config/containers/container-networking/#dns-services)?

OK, I tried to debug locally by using `drone exec` and still got the same error message.

Then I tried again with a simple example to see what happens:

```shell
steps:
  - name: alpine
    image: alpine:3.13
    network_mode: host
    commands:
      - cat /etc/resolv.conf
```

```shell
$ drone exec
2021/01/27 10:34:12 linter: untrusted repositories cannot configure network_mode

$ drone exec --trusted
[alpine:0] + cat /etc/resolv.conf
[alpine:1] # This file is fetched from the host via vpnkit-bridge
[alpine:2] nameserver 192.168.65.1
```

But wait. Why drone linter does not force me to trust the build when running publish step?
Turned out that `network_mode` is put in wrong place. It's a service configuration, not `plugins/docker`'s settings:

```yaml
- name: publish
  image: plugins/docker:19.03
  network_mode: host
  settings:
    debug: true
    registry: docker.domain.com
    repo: docker.domain.com/owner/repo
    tags:
      - ${DRONE_SOURCE_BRANCH}
    username:
      from_secret: docker_username
    password:
      from_secret: docker_password
```