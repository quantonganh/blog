---
title: "gocloud - writing data to a bucket: 403"
date: Mon Dec 23 08:58:59 +07 2022
description:
categories:
- DevOps
- Programming
tags:
- gocloud
- gcs
- golang
---
#### Problem

We are writing some integration test using [Go CDK](https://gocloud.dev/).
After writing some data to a bucket:

```go
	writer, err := buckOut.NewWriter(ctx, fileDst, nil)
	if err != nil {
		logger.Errorf("failed to write to fileDst: %v", err)
		return err
	}
	defer writer.Close()
```

we got an error when reading:

```
(code=NotFound): storage: object doesn't exist
```

By reading [the documentation](https://gocloud.dev/howto/blob/#writing), I pay attention to this:

> Closing the writer commits the write to the provider, flushing any buffers, and releases any resources used while writing, so you must always check the error of Close.

So, just check error to see what happens:

```go
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			logger.Errorf("failed to close the writer: %v", closeErr)
		}
	}()
```

Here's the result:

```shell
googleapi: Error 403: prod-iac@infra-prod.iam.gserviceaccount.com does not have storage.objects.create access to the Google Cloud Storage object.
Permission 'storage.objects.create' denied on resource (or it may not exist)., forbidden
```

We have a script to activate staging service account:

```shell
#!/bin/bash

echo $GCP_SERVICE_ACCOUNT_STAGING | base64 -d > /credential.json
gcloud auth activate-service-account --key-file=/credential.json
export GOOGLE_APPLICATION_CREDENTIALS=/credential.json
```

And it's called in the BB pipeline:

```
          script:
            - /gcp-auth staging
            - go test -count=1 -race -v ./...
```

Why it is still using the prod account instead of the staging account?

#### Troubleshooting

First, we need to understand [how Application Default Credentials works](https://cloud.google.com/docs/authentication/application-default-credentials).
As you can see in the above doc, ADC searches for credentials in the following locations:

1. `GOOGLE_APPLICATION_CREDENTIALS` env
2. User credentials setup with Google Cloud CLI
3. The attached service account, as provided by the metadata server

So, looks like ADC is using attached service account which is provided by the metadata server.
Why it is not using `GOOGLE_APPLICATION_CREDENTIALS`?

So, let see what happens when executing a shell script by adding `pstree -p $$` to the end:

```shell
#!/bin/bash

export GOOGLE_APPLICATION_CREDENTIALS=/credential.json
pstree -p $$
```

```shell
$ ./gcp-auth.sh
       \-+= 00407 quanta -fish
         \-+= 19287 quanta bash
           \-+= 30895 quanta /bin/bash ./gcp-auth.sh
             \-+- 30896 quanta pstree -p 30895
               \--- 30897 root ps -axwwo user,pid,ppid,pgid,command
```

```shell
$ echo $GOOGLE_APPLICATION_CREDENTIALS

```
Here you can see that `gcp-ath` is executed in a subshell (`/bin/bash`).
And since the child processes cannot alter the parent's env, `GOOGLE_APPLICATION_CREDENTIALS` is not exported to the parent,
so, ADC fallback to use attached service account from the metadata server.

#### Solution

To run a shell script in the current shell, use `source` or `.`:

```shell
$ source gcp-auth.sh
       \-+= 35696 quanta -fish
         \-+= 35845 quanta bash
           \-+= 37050 quanta pstree -p 35845
             \--- 37051 root ps -axwwo user,pid,ppid,pgid,command
```

```shell
$ echo $GOOGLE_APPLICATION_CREDENTIALS
/tmp/credential.json
```