---
title: "WezTerm: quickly select a command and open it in a new pane"
date: 2023-08-04
description:
categories:
- Development Environment
tags:
- helix
- wezterm
---
In my previous post, [jump to build errors with WezTerm](../../07/21/jump-to-build-error-helix), I demostrated how to navigate to build error using [hyperlinks](https://wezfurlong.org/wezterm/hyperlinks.html#implicit-hyperlinks).
One limitation of that approach is that it requires a mouse click.

In this post, I explore how to achieve the same functionality using a shortcut key in WezTerm, 
thanks to [QuickSelectArgs](https://wezfurlong.org/wezterm/config/lua/keyassignment/QuickSelectArgs.html) feature.

WezTerm's `QuickSelectArgs` allows us to bind a shortcut key to select specific patterns and perform actions on the selected text.
Let's say we have the following warning:

```sh
For more information about this error, try `rustc --explain E0382`.
warning: `ownership` (bin "ownership") generated 1 warning
error: could not compile `ownership` (bin "ownership") due to previous error; 1 warning emitted
```

Out goal is to quickly select `rustc --explain E0382` and open it in a new [pane](https://wezfurlong.org/wezterm/config/lua/keyassignment/SplitPane.html).

Here's what I've come up with to accomplish this:

```lua
  {
    key = 's',
    mods = 'CMD|SHIFT',
    action = wezterm.action.QuickSelectArgs {
      label = 'open url',
      patterns = {
        'rustc --explain E\\d+',
      },
      action = wezterm.action_callback(function(window, pane)
        local selection = window:get_selection_text_for_pane(pane)
        wezterm.log_info('opening: ' .. selection)
        if startswith(selection, "http") then
          wezterm.open_with(selection)
        elseif startswith(selection, "rustc --explain") then
          local action = wezterm.action{
            SplitPane={
              direction = 'Right',
              command = {
                args = {
                  'rustc',
                  '--explain',
                  selection:match("(%S+)$"),
                },
              },
            };
          };
          window:perform_action(action, pane);
        else
          selection = "$EDITOR:" .. selection
          return open_with_hx(window, pane, selection)
        end
      end),
    },
  },
```

This script works perfectly.

Now, let's take it up a notch and make the output even fancier by using [mdcat](https://github.com/swsnr/mdcat):

```lua
          local action = wezterm.action{
            SplitPane={
              direction = 'Right',
              command = {
                args = {
                  'rustc',
                  '--explain',
                  selection:match("(%S+)$"),
                  '|',
                  'mdcat',
                  '-p',
                },
              },
            };
          };
```

but the pipe operator is wrapped with single quotes, leading to an error:

```sh
error: Unrecognized option: 'p'

⚠️  Process "rustc --explain E0382 '|' mdcat -p" in domain "local" didn't exit cleanly
Exited with code 1
This message is shown because exit_behavior="Hold"
```

The reason behind this error is [WezTerm doesn't use the shell to launch the programs](https://github.com/wez/wezterm/discussions/4089#discussioncomment-6629409),
therefore, I need to run it as a sub-command using `sh -c`, like this:

```lua
          local action = wezterm.action{
            SplitPane={
              direction = 'Right',
              command = {
                args = {
                  '/bin/sh',
                  '-c',
                  'rustc --explain ' .. selection:match("(%S+)$") .. ' | mdcat -p',
                },
              },
            };
          };
```

With this setup, whenever I want to explore an error further, I can quickly select the `rustc --explain Exxxx` command 
and spawn it in a newly created pane or tab.
