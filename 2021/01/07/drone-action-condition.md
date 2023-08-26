---
title: How to run a pipeline step only when pushing to a new branch?
date: Thu Jan  7 09:55:59 +07 2021
description:
categories:
    - DevOps
tags:
    - drone
    - ci
    - monorepo
    - integration-test
    - golang
---
GitHub sends push hook when pushing to a new branch and merging a PR. How can we distinguish between them?

We are running [integration test](/2020/09/29/integration-test-golang) by triggering a downstream build when a PR is merged into specific branches.
In the downstream repository, we pull all the docker images which has been built and pushed from upstream.
The thing is we also used the [drone-convert-pathschanged](https://github.com/meltwater/drone-convert-pathschanged) to [only publish what modules has been changed](../../2020/06/drone-trigger-based-on-modified-dir.md).

So, what happens when pushing to a new `release-*` branch?
No code changed -> publish steps won't run -> the docker images does not exist -> integration test failed.
How can we solve this situation?

GitHub created two hooks when pushing to a new branch:

- X-GitHub-Event: create
- X-GitHub-Event: push

Since Drone ignores the "create" event, how can we know that a branch has been created in a "push" event?
By looking at the [Webhook payloads](https://docs.github.com/en/free-pro-team@latest/developers/webhooks-and-events/webhook-events-and-payloads#push) docs, we can check if "Created" is true or false.
Turn out that there is [another way](https://github.com/drone/go-scm/pull/92#issuecomment-755852095) to check.

Created: https://github.com/drone/drone/pull/3052

A condition for a step to run when pushing to a new branch will be something like this:

```yaml
  when:
    action:
    - create
    branch:
    - release-*
    event:
    - push
```

It can be distinguished with the step is run when a PR is merged into `release-*`:

```yaml
  when:
    action:
      exclude:
      - create
    branch:
    - release-*
    event:
    - push
```