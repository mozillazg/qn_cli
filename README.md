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

$ ./qn_cli -n test/124.txt 1234.txt
Upload file 1234.txt successed: http://tmp-images.qiniudn.com/test/124.txt

$ ./qn_cli --help
Usage of ./qn_cli:
  -a=false: Auto named saved files
  -d="": Save dirname
  -n="": Save name
```
