Source code for my blog at https://quantonganh.com

My passion for sharing knowledge led me to create my blog from the ground up. Writing about my experiences and discoveries in the world of programming became my integral part of my learning process. This journey has not only deepened my understanding of programming but also enabled me to contribute to the broader community by offering insights, solutions and inspiration. Remember, knowledge is a torch; the more you share it, the brighter it shines for others to follow.

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