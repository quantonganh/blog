---
title: Unable to obtain ACME certificate for domains
date: Mon Dec 28 10:30:54 +07 2020
description:
categories:
    - DevOps
tags:
    - traefik
    - letsencrypt
---
Lets Encrypt tells me that my domain contains an invalid character. What is it?

`remark42` is configured like this:

```yaml
  remark42:
    image: umputun/remark42:arm64
    container_name: "remark42"
    restart: always
    labels:
      - traefik.enable=true
      - traefik.http.routers.remark42.rule=Host(`${REMARK_URL}`)
      - traefik.http.routers.remark42.entrypoints=https
      - traefik.http.routers.remark42.tls.certresolver=le
      - traefik.http.services.remark42.loadbalancer.server.port=8080
```

Running `docker-compose up -d remark42`, and I saw the following error in the `traefik` logs:

```
time="2020-12-28T02:42:29Z" level=error msg="Unable to obtain ACME certificate for domains \"https://remark42.domain.com\": 
unable to generate a certificate for the domains [https://remark42.domain.com]: 
acme: error: 400 :: POST :: https://acme-v02.api.letsencrypt.org/acme/new-order :: urn:ietf:params:acme:error:rejectedIdentifier :: 
Error creating new order :: Cannot issue for \"https://remark42.domain.com\": Domain name contains an invalid character, 
url: " providerName=le.acme routerName=remark42@docker rule="Host(`https://remark42.domain.com`)"
```

What's going on? What character is invalid?

Looked at that error more closely, I figured out that the culprit is `https://`. `Host` rule should only contains the domain:

- https://doc.traefik.io/traefik/https/acme/
- https://doc.traefik.io/traefik/routing/routers/#rule
- https://doc.traefik.io/traefik/user-guides/docker-compose/acme-tls/