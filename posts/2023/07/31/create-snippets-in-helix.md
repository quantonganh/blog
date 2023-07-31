---
title: "How to create snippets in Helix?"
date: 2023-07-31
description:
categories:
- Development Environment
tags:
- helix
- lsp
- golang
---
[Helix currently does not support snippets](https://github.com/helix-editor/helix/issues/395).
However, there is a workaround to insert snippets using `:insert-output` to run a shell command.
Still, this might not feel like the most natural way to work with snippets.

For instance, in GoLand, a similar feature called [live templates](https://www.jetbrains.com/help/go/using-live-templates.html#live_templates_types)
allows you to type `err` and expand it into the following:

```go
if err != nil {
  
}
```

I wonder if I can archieve the same functionality in Helix by creating a simple language server 
using the [Language Server Protocol](https://microsoft.github.io/language-server-protocol/).

After exploring different [SDKs for the LSP](https://microsoft.github.io/language-server-protocol/implementors/sdks/) 
I decided to give [go-lsp](https://github.com/TobiasYin/go-lsp) a try, as I am most familiar with Golang.

[Here](https://github.com/quantonganh/snippets-ls)'s what I came up with. Initially, I encountered and error during the first attempt:

```json
{
  "jsonrpc": "",
  "id": 0,
  "result": {
    "capabilities": {
      "textDocumentSync": 1,
      "completionProvider": {
        "triggerCharacters": [
          "."
        ]
      }
    }
  },
  "error": null
}
```

```
2023-07-31T17:12:56.162 helix_lsp::transport [ERROR] snippets-ls err: <- Parse(Error("data did not match any variant of untagged enum ServerMessage", line: 0, column: 0))
```

However, I was able to resolve the issue by [passing the jsonrpc version in the response](https://github.com/TobiasYin/go-lsp/pull/6).

To test this feature, you can create a configuration file like this:

```yaml
main: |-
  func main() {
      $1
  }

err: |-
  if err != nil {
      $1
  }
```

Then add the following into `~/.config/helix/languages.toml`:

```toml
[[language]]
name = "go"
formatter = { command = "goimports"}
language-servers = ["gopls", "snippets-ls"]

[language-server.snippets-ls]
args = ["-config", "/Users/quantong/.config/helix/go-snippets.yaml"]
command="snippets-ls"
```

Now, when we open a `.go` file in Helix, the editor forks a process to start the language server:

```sh
$ pstree -p 818
-+= 00001 root /sbin/launchd
 \-+= 00440 quantong /Applications/WezTerm.app/Contents/MacOS/wezterm-gui
   \-+= 00487 quantong -fish
     \-+= 00818 quantong hx .
       |--- 00855 quantong /opt/homebrew/bin/gopls
       \--- 00856 quantong /Users/quantong/go/bin/snippets-ls -config /Users/quantong/.config/helix/go-snippets.yaml
```

Next, you can type `err` and press tab, Helix will send a `textDocument/completion` request to the language server:

```json
{
  "jsonrpc": "2.0",
  "method": "textDocument/completion",
  "params": {
    "position": {
      "character": 2,
      "line": 25
    },
    "textDocument": {
      "uri": "file:///Users/quantong/Code/personal/snippets-ls/main.go"
    }
  },
  "id": 3
}
```

and the server will respond with the result:

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": [
    {
      "label": "main",
      "kind": 15,
      "insertText": "func main() {\n    $1\n}"
    },
    {
      "label": "err",
      "kind": 15,
      "insertText": "if err != nil {\n    $1\n}"
    }
  ],
  "error": null
}
```

With this setup, you should be able to use snippets in Helix just like in GoLand, enhancing your coding productivity.