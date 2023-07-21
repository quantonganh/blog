---
title: "Helix: How to jump to the build error from the terminal output?"
date: 2023-07-21
description:
categories:
- Development Environment
tags:
- helix
- wezterm
---
In my [previous post](../14/run-code-in-helix), I shared how I run code in Helix, and today I've taken things a step further by integrating WezTerm into Helix.
Now, after running the code, if I encountered errors or warnings, I can simply click on the corresponding file to open it in Helix at the specific line and column number.

Previously, I had to either look at the line number and navigate to that specific line or manually copy the file path and open it in Helix, which was somewhat cumbersome.

That's when I wondered if I could convert the file path into a hyperlink, and have it open in Helix with just a click.

To generate a [hyperlink](https://wezfurlong.org/wezterm/config/lua/config/hyperlink_rules.html) from terminal output,
I made the following addition to the WezTerm config file:

```lua
config.hyperlink_rules = wezterm.default_hyperlink_rules()

table.insert(config.hyperlink_rules, {
  regex = '^/[^/\r\n]+(?:/[^/\r\n]+)*:\\d+:\\d+',
  format = "$0",
})
```

Additionally, to open the file in the Helix editor, I needed a way to find the pane that is on top of the current pane:

```lua
local hx_pane = pane:tab():get_pane_direction("Up")
```

If it does not exist, I will [create](https://wezfurlong.org/wezterm/config/lua/keyassignment/SplitPane.html) a new one:

```lua
    if hx_pane == nil then
      local action = wezterm.action{
        SplitPane={
          direction = direction,
          command = { args = { 'hx', uri } }
        };
      };
      window:perform_action(action, pane);
```

Otherwise I can [send](https://wezfurlong.org/wezterm/config/lua/keyassignment/SendString.html) a command to that pane to open the file:

```lua
    local action = wezterm.action.SendString(":open " .. uri .. "\r\n")
    window:perform_action(action, hx_pane);
```

The full configuration looks like this:

```lua
config.set_environment_variables = {
  PATH = '/Users/quantong/.cargo/bin:'
      .. '/opt/homebrew/bin:'
      .. os.getenv('PATH')
}

wezterm.on('open-uri', function(window, pane, uri)
  local user = os.getenv('USER')
  local start, _ = string.find(uri, '/Users/' .. user)
  if start == 1 then
    local direction = 'Up'
    local hx_pane = pane:tab():get_pane_direction(direction)
    if hx_pane == nil then
      local action = wezterm.action{
        SplitPane={
          direction = direction,
          command = { args = { 'hx', uri } }
        };
      };
      window:perform_action(action, pane);
    else
      local action = wezterm.action.SendString(':open ' .. uri .. '\r\n')
      window:perform_action(action, hx_pane);
    end
    -- prevent the default action from opening in a browser
    return false
  end
  -- otherwise, by not specifying a return value, we allow later
  -- handlers and ultimately the default action to caused the
  -- URI to be opened in the browser
end)

config.hyperlink_rules = wezterm.default_hyperlink_rules()

table.insert(config.hyperlink_rules, {
  regex = '^/[^/\r\n]+(?:/[^/\r\n]+)*:\\d+:\\d+',
  format = '$0',
})
```

In the above code, I've utilized a trick to open only the local file paths (starting with `/Users/<my_username>`) in Helix.
Other links like https://github.com/ will still be opened in the browser.

I'm delighted with the combination of WezTerm and Helix, as it enhances my workflow and productivity.