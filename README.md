# s3tools

s3tools 用于 s3 数据上传测试。

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
# 随机从里面取一个大小，生成指定大小的文件上传。单位支持 `G` `M` `K`
file_size_list: [ '512', '10K', '512K', '10M', '20M' ]
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
