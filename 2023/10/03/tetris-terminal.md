---
title: "Learning Rust by building Tetris: my favorite childhood game"
date: 2023-10-03
description:
categories:
- Programming
tags:
- rust
- tetris
---
![Play tetris 2-player mode in the terminal](/2023/10/03/tetris-2-player.gif)

After completing the [Rust book](https://doc.rust-lang.org/book/) and working through [rustlings](https://github.com/rust-lang/rustlings), I found myself standing at the crossroads, wondering where to go next. It was then that I had an idea - a project that would allow me to apply my newfound knowledge and create something meaningful. My favorite childhood game, Tetris, became the inspiration for my next coding adventure.

Throughout [this project](https://github.com/quantonganh/tetris-tui/), I gained invaluable insights, including:

- Crafting a cross platform CLI application with [crossterm](https://docs.rs/crossterm/latest/crossterm/)
- Establishing communication between two machines through [TcpStream](https://doc.rust-lang.org/stable/std/net/struct.TcpStream.html)
- Achieving thread synchronization using [Barrier](https://doc.rust-lang.org/stable/std/sync/struct.Barrier.html)
- Effectively passing data between threads utilizing [mpsc::channel](https://doc.rust-lang.org/stable/std/sync/mpsc/fn.channel.html)
- ...

PS: I am currently seeking a new remote software engineering opportunity with a focus on backend development. My flexibility extends accommodating time zones within a range of +/- 3 hours from ICT (Indochina Time). If you have any information regarding companies or positions that are actively hiring for such roles, I would greately approciate it if you can kindly leave a comment or get in touch. Your assistance is sincerely valued. Thank you.
