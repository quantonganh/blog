---
title: How I run code in Helix editor?
date: 2023-07-14
description:
categories:
- Development Environment
tags:
- helix
- wezterm
---
In my [previous post](../../05/06/how-I-setup-new-macbook), I mentioned that I've made the switch to [WezTerm](https://wezfurlong.org/wezterm/) and [Helix](https://helix-editor.com/),
and today I want to share how I run my code in this setup.

Previously, I would split Wezterm into two panes and run the code in a new pane.
However, I found it cumbersome to type different commands for each programming language (Go, C, Racket, and so on).
While auto-completion helped to some extent, or using `control-p` to run the previous command, it still took up a little time.

That's when I started exploring the possibility of running my code directly from Helix based on the file extension.
Unfortunately, Helix didn't have a variable for the current filename, I have to compile [a specific PR](https://github.com/helix-editor/helix/pull/6979) to use `%val{filename}`.

To streamline the process, I created a script that executes different commands based on the file extension:

```sh
#!/bin/sh

pane_id=$(wezterm cli get-pane-direction down)
if [ -z "${pane_id}"]; then
  pane_id=$(wezterm cli split-pane)
fi

filename="$1"
basedir=$(dirname "$filename")
basename=$(basename "$filename")
basename_without_extension="${basename%.*}"
extension="${filename##*.}"

case "$extension" in
  "c")
    run_command="clang -Wall -g -O3 $filename -o $basedir/$basename_without_extension && $basedir/$basename_without_extension"
    ;;
  "go")
    run_command="go run $filename"
    ;;
  "rkt")
    run_command="racket $filename"
    ;;
esac

echo "${run_command}" | wezterm cli send-text --pane-id $pane_id --no-paste
```

Then I made the following addition to the Helix's config file:


```toml
[keys.normal.";"]
r = ":sh run.sh %val{filename}"
```

With this setup, you can now simply press `;`, followed by `r` to run your code directly from Helix.

Additionally, I enabled auto-save by adding the following lines to `~/.config/helix/config.toml`

```toml
[editor]
auto-save=true
```
