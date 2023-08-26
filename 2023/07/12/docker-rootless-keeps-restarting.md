---
title: Docker rootless keeps restarting?
date: 2023-07-12
categories:
- DevOps
tags:
- docker
- rootless
- systemd
description:
---
As some of you may know, this blog is hosted on Raspberry Pi. 
To monitor its status, I wrote a script, which you can find [here](https://github.com/quantonganh/blog/blob/master/scripts/monitor.sh).

Recently, I decided to switch the Docker daemon to run in [rootless mode](https://docs.docker.com/engine/security/rootless/).
However, after making this change, I started receiving notifications indicating that the blog was frequently going offline.

Whenever this happens, I ssh into my Pi and run the command `docker ps` to list the running containers.
Strangely, I noticed that the container was only up for less than a second:

```sh
$ docker ps
CONTAINER ID   IMAGE                                 COMMAND                  CREATED       STATUS                  PORTS                                                                                                                 NAMES
9bd3ad610e97   ghcr.io/quantonganh/blog:master       "./blog -config conf…"   3 hours ago   Up Less than a second   80/tcp                                                     
```

Additionally, I checked the status of the Docker service using the following command:

```sh
● docker.service - Docker Application Container Engine (Rootless)
     Loaded: loaded (/home/pi/.config/systemd/user/docker.service; enabled; vendor preset: enabled)
     Active: active (running) since Wed 2023-07-12 09:33:17 BST; 56s ago
       Docs: https://docs.docker.com/go/rootless/
   Main PID: 2313 (rootlesskit)
      Tasks: 214
        CPU: 13.395s
```

I was surprised by the fact that the Docker service showed an active state only 56s ago.

While analyzing the logs, I paid particular attention to these lines:

```
Jul 12 09:31:06 raspberrypi systemd[1]: Stopping User Manager for UID 1000...
Jul 12 09:31:06 raspberrypi systemd[627]: Stopping libcontainer container e0296500d5f75ef63f66686559c5122ed7bd4da2e88210c19f9ab0b352155554.
Jul 12 09:31:06 raspberrypi systemd[627]: Stopping Docker Application Container Engine (Rootless)...
Jul 12 09:31:06 raspberrypi dockerd-rootless.sh[921]: time="2023-07-12T09:31:06.720404072+01:00" level=info msg="Processing signal 'terminated'"
```

These logs indicated that the `systemd` had stopped the User Manager for my UID and sent a SIGTERM signal to the Docker daemon?
I wondered why this was happening.

After some research, I came across the Arch Linux wiki page on [systemd/User instances](https://wiki.archlinux.org/title/systemd/User#Automatic_start-up_of_systemd_user_instances):

> The systemd user instance is started after the first login of a user and killed after the last session of the user is closed.

This led me to suspect that `docker.service` was being terminated when my SSH session is ended.

To confirm my hypothesis, I checked the linger status using the following command:

```sh
$ loginctl user-status $(whoami)
pi (1000)
           Since: Wed 2023-07-12 09:33:10 BST; 57min ago
           State: active
        Sessions: 14 9 8 *4
          Linger: no
```

Indeed, the output showed that lingering was disabled for my user.
Therefore, I enabled lingering for my user:

```sh
$ loginctl enable-linger $(whoami)
$ loginctl user-status $(whoami)
pi (1000)
           Since: Wed 2023-07-12 09:33:10 BST; 1h 0min ago
           State: active
        Sessions: 14 9 8 *4
          Linger: yes
```

Enabling lingering ensured that the Docker service would continue running even after my SSH session was closed.