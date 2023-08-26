---
title: How a gRPC message uses length-prefixed framing?
date: Wed May  6 09:18:44 +07 2020
description:
categories:
    - DevOps
    - Networking
tags:
    - grpc
    - protocol-buffers
    - length-prefixed-framing
---
Once we have the encoded data to send to the other end, we need to package the data in a way that other end can easily extract the information. gRPC uses a message framing technique called length-prefixed framing.

Filter gRPC messages:

```
$ tshark -r grpc.pcapng -Y 'grpc'
```

Find out a `stream: DATA`:

```
216 923.033355    127.0.0.1 → 127.0.0.1    GRPC 108 DATA[1] (GRPC) (PROTOBUF)
```

and show the packet details:

```
tshark -r grpc.pcapng -Y "frame.number == 216" -V
```

```
HyperText Transfer Protocol 2
    GRPC Message: /book.BookInfo/getBook, Request
        Compressed Flag: Not Compressed (0)
        Message Length: 38
        Message Data: 38 bytes
```

- first byte: indicate whether data is compressed or not
- next 4 bytes: size of the encoding binary message (38)
