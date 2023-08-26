---
title: "Open same file in already running instance of Helix"
date: 2023-08-02
description:
categories:
- Development Environment
tags:
- helix
- wezterm
---
Recently, I came across [this issue](https://github.com/helix-editor/helix/issues/2177#issuecomment-1278123316):

> I find that I have occasionally (frequently!) made the mistake of two instances of helix in different terminals editing the same file. 
> This causes me a problem, because I will then have changes in one of the instances overwrite changes that are made in another. 
> The worst case is when I accidentally switch back to an instance that has an older version of the file open.

after some consideration, I believe this issue can be solved by using [WezTerm](https://wezfurlong.org/wezterm/index.html)

The proposed solution involves the following steps:

- Use the command [wezterm cli list](https://wezfurlong.org/wezterm/cli/cli/list.html) to list the panes.
- Check if there is any running instance of `hx` that shares the same current working directory as the file you want to open
- If there is a matching instance, send the command `:open $file_path\r` to that pane to open the file
- If no matching instance is found, call `hx` as usual

Here's the script I've come up with:

```sh
#!/usr/bin/env sh

tty=$(tty)
hostname=$(hostname | tr '[:upper:]' '[:lower:]')
pwd=$(pwd)
file_path=$1

pane_id=$(wezterm cli list --format json | jq --arg tty "$tty" --arg opening_cwd "file://$hostname$pwd" --arg file_path "file://$hostname$file_path" -r '.[] | .cwd as $running_cwd | select((.tty_name != $tty) and (.title | startswith("hx")) and (($opening_cwd | contains($running_cwd)) or ($file_path | contains($running_cwd)))) | .pane_id')
if [ -z "$pane_id" ]; then
    ~/.cargo/bin/hx $1
else
    if [ -n "$file_path" ]; then
        echo ":open ${file_path}\r" | wezterm cli send-text --pane-id $pane_id --no-paste
    else
        echo ":open ${pwd}\r" | wezterm cli send-text --pane-id $pane_id --no-paste
    fi
    wezterm cli activate-pane --pane-id $pane_id
fi
```
