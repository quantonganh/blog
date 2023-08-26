---
title: A simple terminal UI for ChatGPT
date: 2023-04-01
categories:
- Programming
tags:
- golang
- tview
- chatgpt
description:
---
When I started learning Linux, I immediately fell in love with the terminal.
I still remember the feeling of typing commands into the terminal, pressing Enter and seeing if I got the expected result.
As a system administrator and DevOps engineer, I always wanted to do everything on the terminal.

The reasons are simple: it allowed me know what was happening behind the scenes,
and it was faster than switching to other tools, and then back to the terminal to continue working.

These reasons still stand today. I'm a fan of [lazygit](https://github.com/jesseduffield/lazygit)
and always want to write something that uses [gocui](https://github.com/jroimartin/gocui).
After investigating some other TUI libraries, I decided to give [tview](https://github.com/rivo/tview) a try for my project,
which can be found at: https://github.com/quantonganh/chatgpt

By being attentive during your interaction with ChatGPT, you will notice that a conversation title will be generated 
on the left-hand side following the first answer. However, the API for this feature is not available. 
Hence, my strategy involves [asking ChatGPT to produce the title](https://github.com/quantonganh/chatgpt/blob/215ff7e5361599eb6b2658917580bcb79a1b72eb/main.go#L353) for us:

```go
			if list.GetItemCount() == 0 || isNewChat {
				go func() {
					resp, err := createChatCompletion([]Message{
						{
							Role:    roleUser,
							Content: prefixSuggestTitle + content,
						},
					}, false)
```

And same as the version on the web, I also want to allow to edit the title.
But I was having trouble showing an [InputField](https://github.com/rivo/tview/wiki/InputField) 
at the position of a list item when the `e` key is pressed.

Looking at the [Modal](https://github.com/rivo/tview/wiki/Modal) feature in `tview`, an idea came to my mind.

My main layout is something like this:

```go
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(button, 3, 1, false).
				AddItem(list, 0, 1, false), 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(textView, 0, 1, false).
				AddItem(textArea, 5, 1, false), 0, 3, false), 0, 1, false).
		AddItem(help, 1, 1, false)
```

I plan to create a modal on top of this layout, with all items nil except for the one at the list item's position:

```go
	modal := func(p tview.Primitive, currentIndex int) tview.Primitive {
		return tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(nil, 4+(currentIndex*2), 1, false).
					AddItem(p, 1, 1, true).
					AddItem(nil, 0, 1, false), 0, 1, true).
				AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(nil, 0, 1, false).
					AddItem(nil, 5, 1, false), 0, 3, false), 0, 1, true).
			AddItem(nil, 1, 1, false)
	}
```

The interesting part is:

```go
					AddItem(nil, 4+(currentIndex*2), 1, false).
					AddItem(p, 1, 1, true).
```

The "New chat" button has height of 3, so if I want to show an InputField at the position of the first item in the list, it must be at 4.
Since we have a blank line between each item, the formula to calculate the fixed size is:

```
4+(currentIndex*2)
```