---
title: "How to disable ligatures in WezTerm?"
date: 2023-08-01
description:
categories:
- Development Environment
tags:
- helix
- wezterm
---
As you may know, I switched to [WezTerm](https://wezfurlong.org/wezterm/index.html) and [Helix](https://helix-editor.com/) a few months ago.

What surprised me is when I type `!=` in Helix, it is rendered as `≠`.
This behaviour also occurs with other character combinations, for example:

- `<=` becomes rendered as `≤`
- `->` becomes rendered as `→`
- `===` becomes rendered as `≡`

My first guess was that Helix had something to do with the font rendering but I encountered the same issue when typing directly into terminal.

After some research, I found [a relevant link](https://wezfurlong.org/wezterm/faq.html#multiple-characters-being-renderedcombined-as-one-character)
which introduced me to the concept of "ligatures".

It turns out that WezTerm uses [JetBrains Mono](https://www.jetbrains.com/lp/mono/) as the default font, and this font comes with ligatures enabled.
I do not like this feature because I want everything I type to be rendered exactly as it appears on my screen.

To disable ligatures, add the following line to the `~/.wezterm.lua`:

```lua
config.harfbuzz_features = { 'calt=0' }
```