---
title: Reload Helix automatically after switching from lazygit
date: 2023-07-25
description:
categories:
- Development Environment
tags:
- helix
- wezterm
---
This is a continuation of the Helix series.
In the previous posts, I shared methods to:

- [Run the code from within Helix](../14/run-code-in-helix)
- [Jump to the build errors from the terminal output](../21/jump-to-build-error-helix)

I also have a script to call [lazygit](https://github.com/jesseduffield/lazygit) from within Helix:

```sh
#!/bin/sh

source wezterm-split-pane.sh

program=$(wezterm cli list | awk -v pane_id="$pane_id" '$3==pane_id { print $6 }')
if [ "$program" = "lazygit" ]; then
    wezterm cli activate-pane-direction down
else
    echo "lazygit" | wezterm cli send-text --pane-id $pane_id --no-paste
fi
```

Howerver, there is an issue with [git gutter is not updated after committing](https://github.com/helix-editor/helix/issues/4994).

To address this, I've come up with an automatic workaround using WezTerm's support for [custom events](https://wezfurlong.org/wezterm/config/lua/wezterm/on.html#custom-events).
I created an event to reload Helix if we are coming back from `lazygit`:

```lua
wezterm.on('reload-helix', function(window, pane)
  local top_process = basename(pane:get_foreground_process_name())
  if top_process == 'hx' then
    local bottom_pane = pane:tab():get_pane_direction('Down')
    if bottom_pane ~= nil then
      local bottom_process = basename(bottom_pane:get_foreground_process_name())
      if bottom_process == 'lazygit' then
        local action = wezterm.action.SendString(':reload-all\r\n')
        window:perform_action(action, pane);
      end
    end
  end
end)
```

then I just need to trigger [multiple](https://wezfurlong.org/wezterm/config/lua/keyassignment/Multiple.html) actions when switching panes:

```lua
config.keys = {
  {
    key = '[',
    mods = 'CMD',
    action = act.Multiple {
      act.ActivatePaneDirection 'Up',
      act.EmitEvent 'reload-helix',
    }
  },
```

With this solution, Helix will automatically reload when switching from `lazygit`, making the workflow more seamless and efficient.