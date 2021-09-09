---
title: Let's Encrypt too many certificates already issued
date: 2021-02-04
description: Why LE issued too many certificates for one domain?
categories:
  - DevOps
tags:
  - letsencrypt
  - traefik
  - docker
---
Traefik is configured to use Let's Encrypt to generate certificate for my blog (and other services) automatically. One day after restarting, I cannot access to my blog via HTTPS anymore (NET::ERR_CERT_AUTHORITY_INVALID). Why?

By looking at the Traefik logs, I found this:

> time="2021-02-04T01:54:33Z" level=error msg="Unable to obtain ACME certificate for domains \"quantonganh.com\": unable to generate a certificate for the domains [quantonganh.com]: acme: error: 429 :: POST :: https://acme-v02.api.letsencrypt.org/acme/new-order :: urn:ietf:params:acme:error:rateLimited :: Error creating new order :: too many certificates already issued for exact set of domains: quantonganh.com: see https://letsencrypt.org/docs/rate-limits/, url: " providerName=le.acme routerName=blog-secured@docker rule="Host(`quantonganh.com`)"

Reading documentation, I know that the main limit is 50 per week. But wait, how can I hit this limit if I only have some domains?
Traefik configuration:

```yaml
volumes:
  - ./traefik:/etc/traefik:ro
  - /var/run/docker.sock:/var/run/docker.sock:ro
```

What! By mounting the Traefik directory as read-only, how can it update its storage (`acme.json`).
So, it will reissue certificates every time on a restart. That's the reason why I hit this limit.

There is one more thing I can do: since I have some subdomains, why not combine them into a single certificate?
https://doc.traefik.io/traefik/https/acme/

```yaml
labels:
  - traefik.http.routers.blog-secured.tls.domains[0].main=${BLOG_DOMAIN}
  - traefik.http.routers.blog-secured.tls.domains[0].sans=*.${BLOG_DOMAIN}
```

As the docs mentioned, there is no way to reset this limit. Now I have to wait until it expires after a week :D.