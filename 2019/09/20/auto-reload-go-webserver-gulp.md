---
title: Auto reload your Go webserver with Gulp
date: Fri Sep 20 00:07:58 +07 2019
description:
categories:
  - DevOps
  - Programming
tags:
  - web
  - gulp
---
When you developp a webserver with Go, you must compile each time you do an update in your code. Well.. this is redundant. With Gulp you can automatize this task… Indeed, when a go file is modified, a task compile the application in the “bin” folder (“gopath/bin”) then another launch the executable (the webserver).

> https://medium.com/@etiennerouzeaud/autoreload-your-go-webserver-with-gulp-ee5e231d133d

```js
const gulp     = require('gulp'),
      util     = require('gulp-util'),
      notifier = require('node-notifier'),
      child    = require('child_process'),
      os       = require('os'),
      path     = require('path');

var server = 'null'

function build() {
    var build = child.spawn('go', ['install']);

    build.stdout.on('data', (data) => {
        console.log(`stdout: ${data}`);
    });

    build.stderr.on('data', (data) => {
        console.error(`stderr: ${data}`);
    });

    return build;
}

function spawn(done) {
    if (server && server != 'null') {
        server.kill();
    }

    var path_folder = process.cwd().split(path.sep)
    var length = path_folder.length
    var app = path_folder[length - parseInt(1)];

    if (os.platform() == 'win32') {
        server = child.spawn(app + '.exe')
    } else {
        server = child.spawn(app)
    }

    server.stdout.on('data', (data) => {
        console.log(`stdout: ${data}`);
    });

    server.stderr.on('data', (data) => {
        console.log(`stderr: ${data}`);
    });

    done();
}

const serve = gulp.series(build, spawn)
function watch(done) {
    gulp.watch(['*.go', '**/*.go'], serve);
    done();
}

exports.serve = serve
exports.watch = watch
exports.default = gulp.parallel(serve, watch)
```
