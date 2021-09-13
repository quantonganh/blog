---
title: Do not use jq when working with large number
date: 2021-09-07
description:
categories:
- Programming
tags:
- jq
- json
- javascript
- bigint
---
We use [etcd](https://etcd.io/) to store application configuration.
On production, config is loaded from json file by using [Python](https://pypi.org/project/etcd3/).

I'm wondering if we can use [jq](https://github.com/stedolan/jq) to do that. So, I tried something like this:

```shell
host=${ETCD_HOST:-127.0.0.1}
port=${ETCD_PORT:-2379}
for key in $(jq -r 'keys[]' "$1"); do
  value=$(jq -r ".$key" -c "$1")
  ETCDCTL_API=3 etcdctl --endpoints="$host:$port" put "$key" "$value"
done
```

Config is loaded into `etcd` but some values are not the same as in the JSON file.
What is going on?

Looking more closely I found out that `jq` is rounding large number:

```shell
$ echo 947295729583939162 | jq '.'
947295729583939200
```

The reason is `jq` converts the JSON numbers to [IEEE 754](https://en.wikipedia.org/wiki/IEEE_754) 64-bit values and the largest number is 2^53 (9,007,199,254,740,992), same as JavaScript:

```shell
$ node -e "console.log(Number.MAX_SAFE_INTEGER);"
9007199254740991
```

Please take a look at [FAQ](https://github.com/stedolan/jq/wiki/FAQ#numbers) for more details.

