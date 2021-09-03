---
title: Troubleshooting slow network file transfer
date: 2021-06-25
description: Which switch port connected to a server?
categories:
- Networking
tags:
- iperf3
- tcpdump
---
We have a [Hadoop](https://hadoop.apache.org/) data cluster. Recently, my colleague noticed that data transfer between servers was slow. Not all are slow; only half of them have problems.

The first tool that comes to mind is [iperf3](https://github.com/esnet/iperf). After installing, I run `iperf3 -s` then `iperf3 -c`. While other servers have bandwidth > 900 Mbits/sec, the slow ones are just:

```shell
[ ID] Interval           Transfer     Bandwidth       Retr  Cwnd
[  4]   0.00-1.00   sec  11.9 MBytes  99.4 Mbits/sec    0    170 KBytes
[  4]   1.00-2.00   sec  11.1 MBytes  93.3 Mbits/sec    0    204 KBytes
[  4]   2.00-3.00   sec  10.9 MBytes  91.7 Mbits/sec    0    204 KBytes
[  4]   3.00-4.00   sec  8.26 MBytes  69.3 Mbits/sec  102    154 KBytes
[  4]   4.00-5.00   sec  10.9 MBytes  91.8 Mbits/sec    3    157 KBytes
[  4]   5.00-6.00   sec  10.9 MBytes  91.7 Mbits/sec    0    165 KBytes
[  4]   6.00-7.00   sec  10.9 MBytes  91.7 Mbits/sec    0    168 KBytes
[  4]   7.00-8.00   sec  10.9 MBytes  91.7 Mbits/sec    0    173 KBytes
[  4]   8.00-9.00   sec  10.9 MBytes  91.7 Mbits/sec    0    173 KBytes
[  4]   9.00-10.00  sec  10.9 MBytes  91.7 Mbits/sec    0    173 KBytes
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bandwidth       Retr
[  4]   0.00-10.00  sec   108 MBytes  90.4 Mbits/sec  105             sender
[  4]   0.00-10.00  sec   107 MBytes  89.5 Mbits/sec                  receiver
```

Hell, why is that?

I wonder is there any way to know which switch port connected to these servers? A quick Google search lead me to: https://www.lazysystemadmin.com/2011/09/find-out-which-switch-port-connected.html

```shell
# tcpdump -v -i eno1 -s 1500 -c 1 '(ether[20:2]=0x2000 or ether[12:2]=0x88cc)'
```

- `ether[20:2]=0x2000`
  - `ether`: Ethernet protocol
  - `20:2`: filter on the byte 20 and get 2 byte value
  - `0x2000`: Ethernet type of [CDP](https://wiki.wireshark.org/CDP)
    
- `ether[12:2]=0x88cc`
    - `ether`: Ethernet protocol
    - `12:2`: filter on the byte 12 and get 2 byte value
    - `0x88cc`: Ethernet type of [LLDP](https://wiki.wireshark.org/LinkLayerDiscoveryProtocol)
  
Normal server:

```shell
        Platform (0x06), value length: 20 bytes: 'cisco WS-C3560G-48TS'
        Port-ID (0x03), value length: 18 bytes: 'GigabitEthernet0/6'
```

Slow one:

```shell
        Platform (0x06), value length: 19 bytes: 'cisco WS-C3560-48TS'
        Port-ID (0x03), value length: 15 bytes: 'FastEthernet0/5'
```

Do you see the difference?