---
title: How I setup a new Macbook in 2023
date: 2023-05-06
categories:
- Development Environment
tags:
- macOS
- homebrew-bundle
- mackup
- wezterm
- helix
description:
---
I purchased my first MacBook Pro in 2013 prior to transitioning to remote work. 
In my job, we were using [saltstack](https://github.com/saltstack/salt) to build [an automation tool](https://www.robotinfra.com/).
In order to delve deeper into Salt, I decided to configure my Macbook using it, and I documented the process [here](https://github.com/quantonganh/salt-osx).

I also utilized this method to set up another Macbook for my older sister.

In 2018, due to work requirements, 8GB of memory proved insufficient to run Windows on VirtualBox, prompting me to acquire a second hand MacBook with 16GB of RAM.
And I used [Migration Assistant](https://support.apple.com/en-vn/HT204350) to transfer to a new Mac.

Last year, I had to replace battery and speakers. And after that, unfortunately, the keyboard malfunctioned, and I had to purchase a new trackpad cable.

A few months ago, the new battery died after the warranty period expired, necessitating another purchase.
At that point, I told myself: if any further issues arose, I would invest in a new MacBook.

And [my chance has come](../20/unresponsive-macbook-pro) in the last month.

To ensure that I didn't transfer unnecessary software to the new Mac, 
I conducted a thorough search and found [homebrew-bundle](https://github.com/Homebrew/homebrew-bundle) and [mackup](https://github.com/lra/mackup).

Additionally, I had the desire to explore something new, so I decided to transition:

- from [iTerm2](https://iterm2.com/) to [WezTerm](https://wezfurlong.org/wezterm/)
- from Vim to [Helix](https://helix-editor.com/)

You can find my dotfiles [here](https://github.com/quantonganh/dotfiles).
