---
title: How to perform integration testing in Go?
date: 2020-09-29
description:
categories:
    - DevOps
    - Programming
tags:
    - golang
    - drone
    - integration-test
---
Integration testing can be triggered by using Drone [downstream](http://plugins.drone.io/drone-plugins/drone-downstream/) plugin:

```yaml
steps:
- name: trigger
  image: plugins/downstream:linux-amd64
  settings:
    params:
    - COMMIT_BRANCH=${DRONE_COMMIT_BRANCH}
    repositories:
    - repo/integration-test@${DRONE_COMMIT_BRANCH}
    server: https://drone.example.com
    token:
      from_secret: drone_token
```

It can be separated with unit tests by using [build tags](https://golang.org/pkg/go/build/#hdr-Build_Constraints):

```
// +build integration

package webserver_test
```

Then we can write code to perform integration test as usual.

In the integration-test, clone the upstream repo:

```yaml
- name: clone-upstream
  image: alpine/git:v2.24.3
  commands:
  - git clone -b $${COMMIT_BRANCH:-main} https://github.com/owner/upstream.git /upstream
  volumes:
  - name: upstream
    path: /upstream
```

and run integration tests:

```yaml
- name: test-something
  image: golang:1.14.7
  commands:
  - cd /path/to/webserver
  - go test -tags integration -v -run TestSomething
  volumes:
  - name: upstream
    path: /upstream
```