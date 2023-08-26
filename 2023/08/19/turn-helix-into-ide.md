---
title: "Turning Helix into an IDE with the help of WezTerm and CLI tools"
date: 2023-08-19
description:
categories:
- Development Environment
tags:
- broot
- helix
- lazygit
- tig
- wezterm
---
![Helix as IDE](/2023/08/19/hx-ide.gif)

This post recaps previous series of posts about Helix.

1. [Running code](#running-code)
2. [Jumping to build errors](#jumping-to-build-errors)
3. [Quickly select a command and open it in a new pane](#quickly-select-a-command-and-open-it-in-a-new-pane)
4. [Testing a single function](#testing-a-single-function)
5. [Git integration](#git-integration)
6. [File tree](#file-tree)
7. [Creating snippets](#creating-snippets)
8. [How to debug?](#how-to-debug)
9. [Interactive global search](#interactive-global-search)
10. [Opening file in GitHub](#opening-file-in-github)

#### 1. Running code {#running-code}

In my [previous post](/2023/07/14/run-code-in-helix.md), I shared a method for running code from within Helix by using [this PR](https://github.com/helix-editor/helix/pull/6979).

I later discovered another useful trick [here](https://www.reddit.com/r/HelixEditor/comments/13x9a3z/integrating_fuzzylive_grepping_into_helix_my/). We can use [wezterm cli get-text](https://wezfurlong.org/wezterm/cli/cli/get-text.html) command to extract the filename and line number from the status line:

```sh
status_line=$(wezterm cli get-text | rg -e "(?:NORMAL|INSERT|SELECT)\s+[\x{2800}-\x{28FF}]*\s+(\S*)\s[^│]* (\d+):*.*" -o --replace '$1 $2')
filename=$(echo $status_line | awk '{ print $1}')
line_number=$(echo $status_line | awk '{ print $2}')
```

#### 2. Jumping to build errors {#jumping-to-build-errors}

- [/2023/07/21/jump-to-build-error-helix](/2023/07/21/jump-to-build-error-helix.md)

#### 3. Quickly select a command and open it in a new pane {#quickly-select-a-command-and-open-it-in-a-new-pane}

- [/2023/08/04/quick-select-wezterm](/2023/08/04/quick-select-wezterm.md)

#### 4. Testing a single function {#testing-a-single-function}

As we can retrieve the filename and line number from the status line, we can easily extract the test name and pass it to `cargo test`:

```sh
  "test_single")
    test_name=$(head -$line_number $filename | tail -1 | sed -n 's/^.*fn \([^ ]*\)().*$/\1/p')
    split_pane_down
    case "$extension" in
      "rs")
        run_command="cd $PWD/$(dirname "$basedir"); cargo test $test_name; if [ \$status = 0 ]; wezterm cli activate-pane-direction up; end;"
        ;;
    esac
    echo "$run_command" | $send_to_bottom_pane
    ;;
```

#### 5. Git integration {#git-integration}

While Helix currently [does not support Git](https://github.com/helix-editor/helix/issues/227), we can open [lazygit](https://github.com/jesseduffield/lazygit) in the bottom pane using the following script:

```sh
  "lazygit")
    split_pane_down
    program=$(wezterm cli list | awk -v pane_id="$pane_id" '$3==pane_id { print $6 }')
    if [ "$program" = "lazygit" ]; then
        wezterm cli activate-pane-direction down
    else
        echo "lazygit" | $send_to_bottom_pane
    fi
    ;;
```

and [reload automatically](/2023/07/25/auto-reload-helix.md) after switching back.

##### git blame

Since we can obtain the filename and line number, we can easily send it to [tig](https://jonas.github.io/tig/):

```sh
  "blame")
    split_pane_down
    echo "tig blame $filename +$line_number" | $send_to_bottom_pane
    ;;
```

#### 6. File tree {#file-tree}

- [/2023/08/02/file-tree-workaround-for-helix](/2023/08/02/file-tree-workaround-for-helix.md)

Discover how to open [broot](https://github.com/Canop/broot) from within Helix:

```sh
  "explorer")
    left_pane_id=$(wezterm cli get-pane-direction left)
    if [ -z "${left_pane_id}" ]; then
      left_pane_id=$(wezterm cli split-pane --left --percent 20)
    fi

    left_program=$(wezterm cli list | awk -v pane_id="$left_pane_id" '$3==pane_id { print $6 }')
    if [ "$left_program" != "br" ]; then
      echo "br" | wezterm cli send-text --pane-id $left_pane_id --no-paste
    fi

    wezterm cli activate-pane-direction left
    ;;
```

And when opening a file from `broot`, ensure it opens in the right pane:

```sh
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

#### 7. Creating snippets {#creating-snippets}

- [/2023/07/31/create-snippets-in-helix](/2023/07/31/create-snippets-in-helix.md)

#### 8. How to debug? {#how-to-debug}

- [Debugging Rust in Helix](https://quantonganh.com/2023/08/10/debug-rust-helix.md)
- [Debugging Rust in Helix (Part 2)](https://quantonganh.com/2023/08/11/debug-rust-helix-2.md)

I submitted [this PR](https://github.com/helix-editor/helix/pull/7936) to allow the `target` directory to appear in the completion list in the debugger prompt, and it got merged so quickly.

#### 9. Interactive global search {#interactive-global-search}

I'm implementing a mechanism that involves sending a command using [fzf](https://github.com/junegunn/fzf) (combine with [ripgrep](https://github.com/BurntSushi/ripgrep)) to the bottom pane when a key is pressed:

```sh
  "fzf")
    split_pane_down
    echo "hx-fzf.sh \$(rg --line-number --column --no-heading --smart-case . | fzf --delimiter : --preview 'bat --style=full --color=always --highlight-line {2} {1}' --preview-window '~3,+{2}+3/2' | awk '{ print \$1 }' | cut -d: -f1,2,3)" | $send_to_bottom_pane
    ;;
```

Once a selection is made, the file will be opened by Helix in the top pane:

```sh
selected_file=$1
top_pane_id=$(wezterm cli get-pane-direction Up)
if [ -z "$selected_file" ]; then
    if [ -n "${top_pane_id}" ]; then
        wezterm cli activate-pane-direction --pane-id $top_pane_id Up
        wezterm cli toggle-pane-zoom-state
    fi
    exit 0
fi

if [ -z "${top_pane_id}" ]; then
    top_pane_id=$(wezterm cli split-pane --top)
fi

wezterm cli activate-pane-direction --pane-id $top_pane_id Up

send_to_top_pane="wezterm cli send-text --pane-id $top_pane_id --no-paste"

program=$(wezterm cli list | awk -v pane_id="$top_pane_id" '$3==pane_id { print $6 }')
if [ "$program" = "hx" ]; then
    echo ":open $selected_file\r" | $send_to_top_pane
else
    echo "hx $selected_file" | $send_to_top_pane
fi

wezterm cli toggle-pane-zoom-state
```

I have submitted [this PR](https://github.com/wez/wezterm/pull/4160) to hide the bottom pane after that.

#### 10. Opening file in GitHub {#opening-file-in-github}

As simple it seems:

```sh
  "open")
    gh browse $filename:$line_number  
    ;;
```

You can see my full script [here](https://github.com/quantonganh/dotfiles/blob/main/.local/bin/hx-wezterm.sh) and the Helix configuration below:

```toml
[keys.normal.";"]
b = ":sh hx-wezterm.sh blame"
e = ":sh hx-wezterm.sh explorer"
f = ":sh hx-wezterm.sh fzf"
g = ":sh hx-wezterm.sh lazygit"
o = ":sh hx-wezterm.sh open"
r = ":sh hx-wezterm.sh run"
s = ":sh hx-wezterm.sh test_single"
t = ":sh hx-wezterm.sh test_all"
```