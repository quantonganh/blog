---
title: "go/packages.Load: no packages found"
date: Tue Oct 22 05:23:53 +07 2019
description: Go to definition does not work in a mono-repo
categories:
    - Programming
tags:
    - go
    - gopls
---
I have a mono-repo, structure like this:

```
repo
├── service1
├── service2
├── go.mod
```

I often open the root folder from the command line by using `code .`. After that I open a file in `service1`, and cannot go to definition. Here's the logs from `gopls`:

```
[Info  - 5:49:28 AM] 2019/10/22 05:49:28 20.675656ms for GOROOT=/usr/local/Cellar/go/1.13/libexec GOPATH=/Users/quanta/go GO111MODULE=auto PWD=/Users/quanta/go/src/github.com/owner/repo go "list" "-e" "-json" "-compiled=true" "-test=true" "-export=false" "-deps=true" "-find=false" "--" "/Users/quanta/go/src/github.com/owner/repo/a/b/c", stderr: <<go: directory a/b/c is outside main module
>>

[Error - 5:49:32 AM] Request textDocument/definition failed.
  Message: go/packages.Load: no packages found for /Users/quanta/go/src/github.com/owner/repo/a/b/c/file.go
  Code: 0 
```

- https://github.com/golang/go/issues/32667
- https://github.com/golang/go/issues/32394
- https://github.com/microsoft/vscode-go/issues/2490

So, `gopls` only works if we open vscode at the module root (`service1`, `service2` in this case). If you open vscode at the project root, you have to use File -> Add Folder to Workspace to create a workspace with multiple folders.
