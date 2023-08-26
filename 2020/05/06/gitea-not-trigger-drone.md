---
title: Drone build is not triggered after pushing code to Gitea?
date: Wed May  6 14:11:01 +07 2020
description:
categories:
    - DevOps
tags:
    - gitea
    - drone
    - fstab
---
I pushed code to Gitea and nothing happens in Drone. Why?

But if I go to [Settings -> Webhooks](https://git-tea.dynu.net/quanta/blog/settings/hooks/1), then click "Test Delivery" -> build pipeline will be executed. Why?

First, look at the "Recent Deliveries" to see if a webhook is triggerd here when you pushing code.
In my case, it's not. So, looks like the problem is on Gitea side.

By pay close attention to the git output when pushing:

```
Counting objects: 4, done.
Delta compression using up to 4 threads.
Compressing objects: 100% (4/4), done.
Writing objects: 100% (4/4), 2.02 KiB | 2.02 MiB/s, done.
Total 4 (delta 2), reused 0 (delta 0)
hint: The 'hooks/pre-receive' hook was ignored because it's not set as executable.
hint: You can disable this warning with `git config advice.ignoredHook false`.
hint: The 'hooks/update' hook was ignored because it's not set as executable.
hint: You can disable this warning with `git config advice.ignoredHook false`.
hint: The 'hooks/post-receive' hook was ignored because it's not set as executable.
hint: You can disable this warning with `git config advice.ignoredHook false`.
To gitea.pi:quanta/blog.git
 + d29f63d...47e9ed3 master -> master (forced update)
```

Why hooks are not set as executable?

`docker inspect gitea`:

```
        "Mounts": [
            {
                "Type": "volume",
                "Name": "data_gitea",
                "Source": "/data/docker/volumes/data_gitea/_data",
```

```
root@raspberrypi:~# ls -l /data/docker/volumes/data_gitea/_data/git/repositories/quanta/blog.git/hooks/
total 68
-rwxr-xr-x 1 pi pi  303 Oct 11  2019 post-receive
-rwxr-xr-x 1 pi pi  303 Oct 11  2019 pre-receive
-rwxr-xr-x 1 pi pi  283 Oct 11  2019 update
```

What's going on?

So, hooks scripts are executable. Why Gitea said that it's not?
Ah, Gitea volumes are mounted on `/data`:

```
root@raspberrypi:~# grep sda1 /proc/mounts
/dev/sda1 /data ext4 rw,nosuid,nodev,noexec,relatime 0 0
```

The culprit is `noexec` which prevent the direct execution of binaries on the mounted filesystem.
I have to remount `/dev/sda1` with `exec` flag:

```
UUID=25a735b1-76bb-467f-beb1-f7be00a83d3d	/data	ext4	defaults,auto,users,exec,rw,nofail 0 0
```
