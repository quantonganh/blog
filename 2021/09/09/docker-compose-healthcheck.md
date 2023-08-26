---
title: Docker Compose healthcheck
date: 2021-09-09
description:
categories:
- DevOps
tags:
- docker-compose
- healthcheck
- docker
---
The most important thing when running integration test using docker-compose is ensured that one container is started completely before others.

Sometime [wait-for-it](https://github.com/vishnubob/wait-for-it) is not enough:

```yaml
  cassandra:
    image: bitnami/cassandra:latest
    ports:
      - '7000:7000'
      - '9042:9042'
    volumes:
      - /path/to/init-scripts:/docker-entrypoint-initdb.d

  wait-for-cassandra:
    image: willwill/wait-for-it
    command: cassandra:9042 -t 60
    depends_on:
      - cassandra:
```

We need a way to [check if `cassandra` is really healthy](https://docs.docker.com/compose/compose-file/compose-file-v3/#healthcheck).

Since I [CREATE KEYSPACE](https://cassandra.apache.org/doc/latest/cassandra/cql/ddl.html) in the init script:

```cassandraql
CREATE KEYSPACE IF NOT EXISTS test WITH replication = {'class':'SimpleStrategy', 'replication_factor' : 1};
```

I want to make sure that keyspace 'test' exist before starting others:

```yaml
  cassandra:
    image: bitnami/cassandra:latest
    ports:
      - '7000:7000'
      - '9042:9042'
    volumes:
      - /path/to/init-scripts:/docker-entrypoint-initdb.d
    healthcheck:
      test: [ "CMD", "cqlsh", "-u cassandra", "-p cassandra" ,"-k test" ]
      interval: 5s
      timeout: 10s
      retries: 6

  wait-for-cassandra:
    image: willwill/wait-for-it
    command: cassandra:9042 -t 60
    depends_on:
      cassandra:
        condition: service_healthy
```

Try to bring up the project:

```
ERROR: for wait-for-cassandra  Container "d5a87a961ba1" is unhealthy.
```

Why cassandra is unhealthy? [Is there a way to view docker-compose healthcheck logs](https://stackoverflow.com/questions/42737957/how-to-view-docker-compose-healthcheck-logs)?

Run `docker inspect --format "{{ json .State.Health }}" cassandra_1 | jq` and I got:

```json
    {
      "Start": "2021-09-09T00:16:41.5668221Z",
      "End": "2021-09-09T00:16:42.1117382Z",
      "ExitCode": 1,
      "Output": "Connection error: ('Unable to connect to any servers', {'127.0.0.1': AuthenticationFailed('Failed to authenticate to 127.0.0.1: Error from server: code=0100 [Bad credentials] message=\"Provided username  cassandra and/or password are incorrect\"',)})\n"
    }
```

Why authentication failed?
Turned out that there is an extra space before `cassandra` in the error message:

```
Provided username  cassandra and/or password are incorrect
```

So, either you have to remove that space:

```yaml
    healthcheck:
      test: [ "CMD", "cqlsh", "-ucassandra", "-pcassandra", "-ktest" ]
      interval: 5s
      timeout: 10s
      retries: 6
```

or use the long options:

```yaml
    healthcheck:
      test: [ "CMD", "cqlsh", "--username", "cassandra", "--password", "cassandra", "--keyspace", "test" ]
      interval: 5s
      timeout: 10s
      retries: 6
```

or use the shell form (if the container has a shell `/bin/sh`):

```yaml
    healthcheck:
      test: cqlsh -u cassandra -p cassandra -k test
      interval: 5s
      timeout: 10s
      retries: 6
```