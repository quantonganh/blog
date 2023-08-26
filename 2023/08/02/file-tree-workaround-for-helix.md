---
title: "File tree workaround for Helix"
date: 2023-08-02
description:
categories:
- Development Environment
tags:
- helix
- wezterm
- nnn
---
![nnn as a file tree workaround for Helix](/2023/08/02/nnn-hx.jpeg)

Many people have expressed that the absence of a file tree is what prevents them from using Helix.

If, for some reasons, you cannot use [the PR #5768](https://github.com/helix-editor/helix/pull/5768),
here's a workaround to create a file tree using [WezTerm](https://wezfurlong.org/wezterm/index.html) and [nnn](https://github.com/jarun/nnn)

The idea here is to implement [a custom opener](https://github.com/jarun/nnn/wiki/Advanced-use-cases#cli-only-opener) in `nnn.`
When navigating to a file, it will be opened using `hx` in the right pane. 
When opening another file, it checks if there is a running instance of `hx` in the right pane and [sends](https://wezfurlong.org/wezterm/cli/cli/send-text.html) the `:open $1\r` command to that pane.

Here's the implementation:

```sh
#!/usr/bin/env sh

fpath="$1"

pane_id=$(wezterm cli get-pane-direction right)
if [ -z "${pane_id}" ]; then
  pane_id=$(wezterm cli split-pane --right --percent 80)
fi

program=$(wezterm cli list | awk -v pane_id="$pane_id" '$3==pane_id { print $6 }')
if [ "$program" = "hx" ]; then
  echo ":open ${fpath}\r" | wezterm cli send-text --pane-id $pane_id --no-paste
else
  echo "hx ${fpath}" | wezterm cli send-text --pane-id $pane_id --no-paste
fi

wezterm cli activate-pane-direction --pane-id $pane_id right
```

```sh
$ export NNN_OPENER=nnn-hx.sh
$ nnn -c
```

This allows you to efficiently manage your files and work with Helix using WezTerm and `nnn`.
