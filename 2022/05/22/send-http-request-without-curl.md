---
title: How to send an HTTP request without using curl?
date: Sun May 22 09:16:50 +07 2022
description:
categories:
- Networking
tags:
- bash
- http
---
#### Problem

We are using [JWT validation](https://www.krakend.io/docs/authorization/jwt-validation/).
For some reasons, when testing on staging, we got 401 error:

```
[GIN] 2022/05/20 - 14:20:57 | 401 |    2.588128ms |       127.0.0.1 | POST     "/v1/endpoint"
```

#### Troubleshooting

After looking at the [source code](https://github.com/devopsfaith/krakend-jose/blob/master/gin/jose.go#L129-L154),
we need to set the [operation_debug](https://www.krakend.io/docs/authorization/jwt-validation/#jwt-validation-settings) to true
to see what caused that error:

```
2022/05/20 08:31:26 KRAKEND ERROR: [ENDPOINT: /v1/endpoint][JWTValidator] Unable to validate the token: should have a JSON content type for JWKS endpoint
```

The thing is when testing locally or use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) we see that the `Content-Type` is set correctly:

```
$ http get localhost:50050/.well-known/jwks.json
HTTP/1.1 200 OK
Content-Length: 419
Content-Type: application/json
Date: Fri, 20 May 2022 08:54:56 GMT
```

So, what could be the reason?

After [ssh to that pod](https://kubernetes.io/docs/tasks/debug/debug-application/get-shell-running-container/),
I realized that `curl`, `wget`, `nc`, ... is not installed. And I don't have permission to do that.
I was wondering if I can send an HTTP request without using any external tool?

#### Solution

Reading [Redirections](https://www.gnu.org/software/bash/manual/html_node/Redirections.html#:~:text=Redirection%20allows%20commands'%20file%20handles,the%20current%20shell%20execution%20environment.) section in the Bash manual,
I saw this:

> /dev/tcp/host/port
> 
> If host is a valid hostname or Internet address, and port is an integer port number or service name, Bash attempts to open the corresponding TCP socket.

So, looks like I can use this Bash's feature to open a TCP socket to the target host.

First, we need to open a file descriptor for reading and writing on the specified TCP socket:

```shell
$ exec 3<>/dev/tcp/s-auth/50050
```

Then send an HTTP request to that socket:

```shell
$ echo -e "GET /.well-known/jwks.json HTTP/1.1\nHost: s-auth\nConnection: close\n\n" >&3
```

You can see this request header by using `curl`:

```shell
$ curl -v localhost:50050/.well-known/jwks.json
*   Trying ::1:50050...
* Connected to localhost (::1) port 50050 (#0)
> GET /.well-known/jwks.json HTTP/1.1
> Host: localhost:50050
> User-Agent: curl/7.77.0
> Accept: */*
```

And read the response:

```shell
$ cat <&3
HTTP/1.1 502 Bad Gateway
content-length: 87
content-type: text/plain
date: Fri, 20 May 2022 09:33:33 GMT
server: envoy
x-envoy-upstream-service-time: 13

upstream connect error or disconnect/reset before headers. reset reason: protocol error
```

I will write another blog post to troubleshoot this error.