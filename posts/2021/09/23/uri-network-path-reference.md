---
title: URI network-path reference
date: 2021-09-23
description:
categories:
- Programming
tags:
  - uri
---
#### Problem

I've just updated the code to always prefix the post URI with a slash:

```go
if !strings.HasPrefix(p.URI, "/") {
    p.URI = "/" + p.URI
}
```

and I think that no need to update the HTML template:

```html
<a href="/{{ .URI }}">{{ .Title }}</a>
```

as the redundant slash will be removed. But I was wrong.

Looking at the link address, instead of seeing:

```
http://localhost//2021/09/23/uri-network-path-reference
```

I saw:

```
http://0.0.7.229/09/23/uri-network-path-reference
```

What's going on?

#### Troubleshooting

The first question comes to my mind is what does `href` do if it begins with two slashes?

Take a look at the [RFC 3986](https://datatracker.ietf.org/doc/html/rfc3986#section-4.2):

> A relative reference that begins with two slash characters is termed
a network-path reference

So, `//2021` is termed a network-path relative reference.
But why:

- `http://localhost//2021` becomes `http://0.0.7.229`?
- `http://localhost//2020` becomes `http://0.0.7.228`?
- `http://localhost//2019` becomes `http://0.0.7.227`?

Please pay attention to the "network" in the above quote.
`href` treated `//2021` as a network address:

```
2^8 = 256
2021 = 256 * 7 + 229
```

That's reason why `http://localhost//2021` becomes `http://0.0.7.229`.

#### Conclusion

If you are not using network-path relative reference, make sure that the URI has only one forward slash.

Happy learning.