---
title: How to show all HTTP2 headers using tshark?
date: Tue May  5 08:44:55 +07 2020
description:
categories:
    - DevOps
    - Networking
tags:
    - grpc
    - http2
    - tshark
---
After sniffing with tcpdump, how can I show all HTTP2 header using tshark?

First, find the frame number based on method:

```
$ tshark -r grpc.pcapng -Y 'http2.headers.path contains "getBook"'
214 923.033174    127.0.0.1 → 127.0.0.1    HTTP2 150 HEADERS[1]: POST /book.BookInfo/getBook
```

and then show the packet details:

```
tshark -r grpc.pcapng -Y "frame.number == 214" -V

HyperText Transfer Protocol 2
    Stream: HEADERS, Stream ID: 1, Length 85, POST /book.BookInfo/getBook
        Length: 85
        Type: HEADERS (1)
        Flags: 0x04
            .... ...0 = End Stream: False
            .... .1.. = End Headers: True
            .... 0... = Padded: False
            ..0. .... = Priority: False
            00.0 ..0. = Unused: 0x00
        0... .... .... .... .... .... .... .... = Reserved: 0x0
        .000 0000 0000 0000 0000 0000 0000 0001 = Stream Identifier: 1
        [Pad Length: 0]
        Header Block Fragment: 8386459162339faaf74e7eb92a94ec4c54dd39faff418b08…
        [Header Length: 216]
        [Header Count: 8]
        Header: :method: POST
            Name Length: 7
            Name: :method
            Value Length: 4
            Value: POST
            :method: POST
            [Unescaped: POST]
            Representation: Indexed Header Field
            Index: 3
        Header: :scheme: http
            Name Length: 7
            Name: :scheme
            Value Length: 4
            Value: http
            :scheme: http
            [Unescaped: http]
            Representation: Indexed Header Field
            Index: 6
        Header: :path: /book.BookInfo/getBook
            Name Length: 5
            Name: :path
            Value Length: 22
            Value: /book.BookInfo/getBook
            :path: /book.BookInfo/getBook
            [Unescaped: /book.BookInfo/getBook]
            Representation: Literal Header Field with Incremental Indexing - Indexed Name
            Index: 5
        Header: :authority: 127.0.0.1:50051
            Name Length: 10
            Name: :authority
            Value Length: 15
            Value: 127.0.0.1:50051
            :authority: 127.0.0.1:50051
            [Unescaped: 127.0.0.1:50051]
            Representation: Literal Header Field with Incremental Indexing - Indexed Name
            Index: 1
        Header: content-type: application/grpc
            Name Length: 12
            Name: content-type
            Value Length: 16
            Value: application/grpc
            content-type: application/grpc
            [Unescaped: application/grpc]
            Representation: Literal Header Field with Incremental Indexing - Indexed Name
            Index: 31
        Header: user-agent: grpc-go/1.24.0
            Name Length: 10
            Name: user-agent
            Value Length: 14
            Value: grpc-go/1.24.0
            user-agent: grpc-go/1.24.0
            [Unescaped: grpc-go/1.24.0]
            Representation: Literal Header Field with Incremental Indexing - Indexed Name
            Index: 58
        Header: te: trailers
            Name Length: 2
            Name: te
            Value Length: 8
            Value: trailers
            [Unescaped: trailers]
            Representation: Literal Header Field with Incremental Indexing - New Name
        Header: grpc-client: evans
            Name Length: 11
            Name: grpc-client
            Value Length: 5
            Value: evans
            [Unescaped: evans]
            Representation: Literal Header Field with Incremental Indexing - New Name
```

Notice that, in gRPC, all requests are HTTP POST with content-type is `application/grpc`.
