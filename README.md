qn-cli
======

Qiniu upload client by Go


Usage
------

```
$ go build -o qn_cli main.go

$ export QINIU_BUCKET_NAME="<QINIU_BUCKET_NAME>"
$ export QINIU_BUCKET_URL="<QINIU_BUCKET_URL>"
$ export QINIU_ACCESS_KEY="<QINIU_ACCESS_KEY>"
$ export QINIU_SECRET_KEY="<QINIU_SECRET_KEY>"

$ ./qn_cli -s test/README2.md README.md
http://tmp-images.qiniudn.com/test/README2.md

$ ./qn_cli --help
Usage of ./qn_cli:
  -s="--save": save name
```
