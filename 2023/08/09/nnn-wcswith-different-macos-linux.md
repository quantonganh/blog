---
title: "nnn: wcswidth returns different values for same Unicode code point between macOS and Linux"
date: 2023-08-09
description:
categories:
- Development Environment
tags:
- nnn
- wcswidth
- macOS
---
When attempting to [integrate `nnn` with Helix](/2023/08/02/file-tree-workaround-for-helix.md), I encountered [an interesting issue](https://github.com/jarun/nnn/issues/1692) that resulted in duplicate first characters in the filename:

![nnn duplicate first character in the filename on macOS](/2023/08/09/nnn-duplicate-first-char-macOS.png)

What's interesting is that this issue doesn't occur on Linux:

![nnn on Linux](/2023/08/09/nnn-linux.png)

I've spent several days troubleshooting this problem but haven't been able to identify the root cause.
The only solution I could come up with is [a workaround](https://github.com/jarun/nnn/pull/1711), and I must admit, I'm still not entirely satisfied with it.

If you also use `nnn` and know of a better way to fix this issue without requiring a refresh, please don't hesitate to let me know. Thank you!