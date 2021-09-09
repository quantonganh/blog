---
title: condition form of depends_on in docker-compose version 3
date: 2021-03-05
description: How to control startup order in Compose?
categories:
- DevOps
tags:
- docker-compose
- healthcheck
- docker
---
As version 3 no longer supports the `condition` form of `depends_on`, what is the alternative way to wait for a container to be started completely?
  
From [1.27.0](https://github.com/docker/compose/releases/tag/1.27.0), 2.x and 3.x are merged with [COMPOSE_SPEC](https://github.com/compose-spec/compose-spec/blob/master/spec.md) schema.

[version](https://github.com/compose-spec/compose-spec/blob/master/spec.md#version-top-level-element) is now optional. So, you can just remove it and specify a [condition](https://github.com/compose-spec/compose-spec/blob/master/spec.md#depends_on) as before:

```
services:
  web:
    build: .
    depends_on:
      redis:
        condition: service_healthy
  redis:
    image: redis
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 1s
      timeout: 3s
      retries: 30
```
