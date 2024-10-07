# serving-webhook

## Setup

Install dependencies:

```shell
go mod tidy
```

Install mockgen:

```shell
go install go.uber.org/mock/mockgen@v0.2.0
go get go.uber.org/mock/mockgen/model@v0.2.0
```

Run tests:

```shell
make test
```

Run on local machine:

```shell
make server
```

## Overview

This repo provides webhooks for uploading files to S3 or GCS and storing prediction task information:

1. After the model prediction is finished, if the prediction results are images or files (specified by using
   [kservehelper](https://github.com/yangwenz/kserve-helper)), the webhook will be called to upload these
   files to external storage and return the file URLs.
2. The task information is stored in redis or redis cluster. The task info includes "model name", "status",
   "running time", "outputs", etc.

## API Definition

The key APIs:

|    API     |         Description          | Method |                       Input                       |
:----------:|:----------------------------:|:------:|:-------------------------------------------------:
|  /upload   | The API for uploading files  |  POST  |         Content-Type: multipart/form-data         |
|   /task    |      Create a new task       |  POST  | {"id": "<TASK_ID>", "model_name": "<MODEL_NAME>"} |
| /task/{ID} |   Get the task information   |  GET   |                        NA                         |
|   /task    | Update an existing task info |  PUT   |      {"id": "", "status": "succeeded", ...}       |

## Parameter Settings

Here are the key parameters:

|       Parameter       |                 Description                 | Sample value  |
:---------------------:|:-------------------------------------------:|:-------------:
|  HTTP_SERVER_ADDRESS  | The TCP address for the server to listen on | 0.0.0.0:12000 |
|     REDIS_ADDRESS     |     The redis server address for Asynq      | 0.0.0.0:6379  |
|  REDIS_CLUSTER_MODE   |        Whether it is a redis cluster        |     False     |
|  REDIS_KEY_DURATION   |        The duration of a task record        |      12h      |
|      AWS_REGION       |          The region of a S3 bucket          |   us-east-2   |
|      AWS_BUCKET       |             The S3 bucket name              |  test-bucket  |
| AWS_S3_USE_ACCELERATE |       Whether to use S3 acceleration        |     False     |
|   AWS_ACCESS_KEY_ID   |            The AWS access key ID            |     xxxxx     |
| AWS_SECRET_ACCESS_KEY |          The AWS secret access key          |     xxxxx     | 

If `REDIS_ADDRESS` is empty, the webhook will not include task related APIs.
