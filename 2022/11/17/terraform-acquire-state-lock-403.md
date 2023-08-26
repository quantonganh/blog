---
title: "Terraform failed to acquire state lock: 403: Access denied., forbidden"
date: Thu Nov 17 10:12:50 +07 2022
description:
categories:
- DevOps
- Programming
tags:
- terraform
- gcs
- golang
---
#### Problem

We stored Terraform state in [gcs](https://developer.hashicorp.com/terraform/language/settings/backends/gcs).
For some reasons, we got this error randomly when running on BitBucket pipelines:

```terraform
╷
│ Error: Error acquiring the state lock
│ 
│ Error message: 2 errors occurred:
│ 	* writing
│ "gs://bucket/path/to/default.tflock"
│ failed: googleapi: Error 403: Access denied., forbidden
│ 	* storage: object doesn't exist
│ 
│ 
│ 
│ Terraform acquires a state lock to protect the state from being written
│ by multiple users at the same time. Please resolve the issue above and try
│ again. For most commands, you can disable locking with the "-lock=false"
│ flag, but this is not recommended.
╵
```

It can be successful after re-running 3, 4 times.

#### Troubleshooting

I tried to reproduce this issue on my local, but I cannot.
Please notice that if the state has been locked by other pipeline, we will get something different:

```terraform
│ Error: Error acquiring the state lock
│ 
│ Error message: writing
│ "gs://bucket/path/to/default.tflock"
│ failed: googleapi: Error 403: Access denied., forbidden
│ Lock Info:
│   ID:        1668249141203541
│   Path:      gs://bucket/path/to/default.tflock
│   Operation: OperationTypePlan
│   Who:       root@de3c91cdae61
│   Version:   1.4.0
│   Created:   2022-11-12 10:32:21.161498675 +0000 UTC
│   Info: 
```

And if the service account do not have permission on that bucket, we will get another error:

```terraform
╷
│ Error: Error acquiring the state lock
│
│ Error message: 2 errors occurred:
│ 	* writing "gs://bucket/path/to/default.tflock" failed: googleapi: Error 403:
│ username@project.iam.gserviceaccount.com does not have storage.objects.create access to the Google Cloud Storage object. Permission 'storage.objects.create' denied
│ on resource (or it may not exist)., forbidden
│ 	* googleapi: got HTTP response code 403 with body: <?xml version='1.0' encoding='UTF-8'?><Error><Code>AccessDenied</Code><Message>Access
│ denied.</Message><Details>username@project.iam.gserviceaccount.com does not have storage.objects.get access to the Google Cloud Storage object. Permission
│ 'storage.objects.get' denied on resource (or it may not exist).</Details></Error>
```

Moreover, I cannot think of a reason why a service account don't have permission randomly.

By re-reading the [documentation](https://developer.hashicorp.com/terraform/language/settings/backends/gcs#authentication),
I pay attention to this:

> IAM Changes to buckets are eventually consistent and may take upto a few minutes to take effect. Terraform will return 403 errors till it is eventually consistent.

I suspected that something changed IAM permissions but our DevOps team confirmed that there is no such cron job.

Here are the other things I tried, but it didn't help:

- turn on [debug](https://developer.hashicorp.com/terraform/internals/debugging) by using `TF_LOG=DEBUG`
- update `terraform` to the latest version 1.3.4

I tried to take a deeper look at the source code::

- [terraform/internal/backend/remote-state/gcs/client.go](https://github.com/hashicorp/terraform/blob/d43ec0f30f56973702a8a8b1ddd016831612921b/internal/backend/remote-state/gcs/client.go#L106)
- [google-cloud-go/storage/writer.go](https://github.com/googleapis/google-cloud-go/blob/4766d3e1004119b93c6bd352024b5bf3404252eb/storage/writer.go#L196)
- [google-cloud-go/storage/http_client.go](https://github.com/googleapis/google-cloud-go/blob/c987e969e4f97feebc4a88308dcb90b29d89a1ec/storage/http_client.go#L1054)
- [google-api-go-client/storage/v1/storage-gen.go#L9997](https://github.com/googleapis/google-api-go-client/blob/ee25e29fd586cde25a006504d0059194a90f19ac/storage/v1/storage-gen.go#L9997)
- [google-api-go-client/storage/v1/storage-gen.go#L9948](https://github.com/googleapis/google-api-go-client/blob/ee25e29fd586cde25a006504d0059194a90f19ac/storage/v1/storage-gen.go#L9948)

and add some debug code to print the credential:

```go
	credential, err := google.FindDefaultCredentials(context.Background())
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] google-api-go-client/storage/v1: doRequest: credential: %+v", credential)
	log.Printf("[TRACE] google-api-go-client/storage/v1: doRequest: credential.JSON: %s", string(credential.JSON))
	var data map[string]interface{}
	if err := json.Unmarshal(credential.JSON, &data); err == nil {
		log.Printf("[TRACE] google-api-go-client/storage/v1: doRequest: credential: %v", data)
	}
```

and I know that [our runners are running on Google Compute Engine](https://github.com/golang/oauth2/blob/fd043fe589d2d1486b6af56f44a691e819752a23/google/default.go#L148).

By re-reading [documentation](https://developer.hashicorp.com/terraform/language/settings/backends/gcs#authentication) one more time:

> If you are running terraform on Google Cloud, you can configure that instance or cluster to use a Google Service Account.
> This will allow Terraform to authenticate to Google Cloud without having to bake in a separate credential/authentication file.
> Make sure that the scope of the VM/Cluster is set to cloud-platform.

I found out that one of our runners does not have that access scope:

```
    "scopes": [
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/trace.append"
    ]
  }
```

#### Solution

Edit the VM and make sure that it has `cloud-platform` scope:

```
    "scopes": [
      "https://www.googleapis.com/auth/cloud-platform",
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/trace.append"
    ]
```

Happy debugging!