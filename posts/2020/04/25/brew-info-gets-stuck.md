---
title: brew info gets stuck
date: Sat Apr 25 08:44:15 +07 2020
description:
categories:
    - DevOps
tags:
    - saltstack
    - brew
---
I [managed my Mac with SaltStack](https://github.com/quantonganh/salt-osx).

For some reasons, it gets stuck when running `brew` state:

```
[INFO    ] Executing command '/usr/local/bin/brew info --json=v1 --installed' as user 'quanta' in directory '/Users/quanta'
```

As usual, whenever you get a problem, let's enable debug mode to see what happens:

```
brew info --json=v1 --installed -d
```

Now I can see that it stucked at [drone/drone](https://github.com/drone/homebrew-drone) repo.
By just untap this repo, and it's solved.
