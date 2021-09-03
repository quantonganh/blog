---
title: A && B || C is not the same as if-then-else
date: Sat Dec 26 16:06:55 +07 2020
description:
categories:
  - DevOps
tags:
  - bash
  - shell
---
I have been thought that if-then-else statement can be written in one line by using `A && B || C` but I was wrong.
    
To speed up my Drone CI time, I configured the static-check step to only run when there are some `.go` files have been changed:

```shell script
git --no-pager diff --name-only ${DRONE_COMMIT_LINK##*/} | grep -q "\.go$" && golangci-lint run --deadline 2m -v ./... || true
```

I was thought that `A && B || C` is the same as:

```shell script
if [[ A ]]; then
    B
else
    C
fi
```

But it's not. The pipeline is still success even that step is failed:

```shell script
+ git --no-pager diff --name-only 1098f1323249f4560a706501a9a76139bb149551...cac8967252fe374561e37b81c035d428d3fff163 | grep -q "\.go$" && golangci-lint run --deadline 2m -v ./... || true
fatal: Invalid symmetric difference expression 1098f1323249f4560a706501a9a76139bb149551...cac8967252fe374561e37b81c035d428d3fff163
```

Actually, `A && B || C` means that:

- run A first
- if the exit status of A is zero -> run B
    - if the exit status of B is zero -> stop
    - if the exit status of B is non-zero -> run C
- if the exit status of A is non-zero -> run C

So, `A && B || C` will be the same as `if [[ A ]]; then B; else C; fi` if B is always success.
My problem is `golangci-lint` returned an error but `true` is run -> the pipeline still success.

It can be fixed by switching back to use `if`:

```shell script
if git --no-pager diff --name-only ${DRONE_COMMIT_LINK##*/} | grep -q "\.go$"; then golangci-lint run --deadline 2m -v ./...; fi
```