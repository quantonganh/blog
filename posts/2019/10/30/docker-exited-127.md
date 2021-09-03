---
title: Why my golang docker container exits immediately (code 127)?
date: Wed Oct 30 10:31:36 +07 2019
description:
categories:
    - DevOps
tags:
    - docker
---
To trim the binary size, I used `LDFLAGS='-w -s'`, pack with `upx`, then build from scratch. The thing is when starting, it exited immediately with code 127. Why?

My Dockerfile:

```
FROM scratch

WORKDIR /app

COPY build/linux/<binary> .

ENTRYPOINT [ "/app/<binary>" ]
```

When starting:

```
0fbce782a9bd        quantonganh/<binary>:T276-dockerize                              "/app/<binary>"           6 seconds ago       Exited (127) 4 seconds ago                                           relaxed_thompson
```

and there is no logs at all:

```
❯ docker logs 0fbce782a9bd
```

IIRC, 127 means that "The supplied command cannot be found", but I'm sure my binary exists, so what's the problem?

To get into this container, we need `/bin/bash`, so I switched from `scratch` to `alpine` as a temporary:

```
FROM alpine:3.10

RUN apk add --no-cache bash
...
```

then I tried to run that binary:

```
❯ docker run -it --entrypoint /bin/bash <image> /app/<binary>
/app/<binary>: /app/<binary>: cannot execute binary file
```

Why "cannot execute"?

Let's see what the file type is:

```
❯ docker run -it --entrypoint /bin/bash <image> -s
bash-5.0# apk add file
bash-5.0# file /app/<binary>
/app/binary: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, Go BuildID=9goElUU1Hmb_TUJeOeZr/vWwqHpXL1j9jP4lLwfaC/JmUJczr0kCVHg1u9gMuT/orDzu6T6STfr9G_YODTe, stripped
```

What? Why "dynamically linked"?

Turned out that I forgot to specify `CGO_ENABLED=0` when building, so it was dynamically linked:

```
bash-5.0# ldd /app/<binary>
	/lib64/ld-linux-x86-64.so.2 (0x7f1f8b561000)
	libpthread.so.0 => /lib64/ld-linux-x86-64.so.2 (0x7f1f8b561000)
	libc.so.6 => /lib64/ld-linux-x86-64.so.2 (0x7f1f8b561000)
```

These does not exist, so the container exited with code 127 immediately.
