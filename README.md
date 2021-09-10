
# s3tools

## build

```shell
make build
```

## introduce

config file:

```yaml
credential:
  access_key_id: 'your access key id'  # AccessKeyId
  access_key_secret: 'your acdess key secret' # AccessKeySecret
  endpoint: 'http://127.0.0.1:20000'  # s3 endpoint
  region: z0  # s3 region

# 要上传到哪个 bucket
bucket: liujie
# 随机从里面取一个大小，生成这么大的文件上传
file_size_list: [ '10K', '512K', '10M', '20M' ]
# 上传时随机取一个 deleteAfterDays
delete_after_days: [ '1', '2', '4', '7', '10' ]
# 上传的 worker 数量（并发数）
workers: 2

```

Print the help:

```shell
$ s3tools -h

Usage of s3tools:
  -f string
        the config path (default "s3tools.yaml")
  -n int
        total files count (default 1)
```

upload files:

```shell
$ s3tools -f s3tools.yaml -n 20
```
