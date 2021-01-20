Source code for my blog at https://quantonganh.com

- Install:

```shell
$ go get -v github.com/quantonganh/blog/cmd/blog
```

- Create a root directory for your blog:

```shell
$ mkdir -p some/dir
```

- Copy the HTML templates, assets:

```shell
$ cp -r /path/to/blog/http/html/templates some/dir/
$ cp -r /path/to/blog/http/assets some/dir/
```

- Create a `config.yaml` file in `some/dir`:

```yaml
templates:
  dir: templates
```

- Create your first blog post `test.md` in `some/dir/posts`:

```yaml
---
title: Test
date: Thu Sep 19 21:48:39 +07 2019
description: Just a test
tags:
  - test
---
Test.
```

- Run:

```shell
$ cd some/dir
$ /path/to/gopath/bin/blog
```

Then open a browser and access your blog at http://localhost.