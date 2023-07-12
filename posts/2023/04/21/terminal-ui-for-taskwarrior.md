---
title: A terminal UI for Taskwarrior
date: 2023-04-21
categories:
- Programming
tags:
- golang
- tview
- taskwarrior
description:
---
When I first started using Linux, I came across [Taskwarrior](https://taskwarrior.org/), a command line tool for managing to-do lists.
However, I never got around to using it extensively, mainly because there was no Android app available at the time, making it difficult to sync tasks from my phone.

I did experiment with [taskwarrior-pomodoro](https://github.com/coddingtonbear/taskwarrior-pomodoro) but unfortunately, it wasn't as effective as I had hoped.

Recently, though, I stumbled upon [foreground](https://github.com/bgregos/foreground), a new discovery that reignited my interest in Taskwarrior, 
so I decided to give it another chance.

To streamline the setup process, I wrote an Ansible playbook (available [here](https://github.com/quantonganh/ansible-pi/blob/main/taskd/tasks/main.yml)) to install taskd on my Raspberry Pi.

Motivated by this progress, I took things a step further and developed a terminal-based user interface (TUI) for Taskwarriror.
You can find the project [here](https://github.com/quantonganh/taskwarrior-tui).

I truly believe that what we've built will prove to be useful and enjoyable to use.