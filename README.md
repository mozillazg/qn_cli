qn-cli
======

Qiniu upload client by Go


Installation
------------

```
$ go get -u github.com/qiniu/api
$ go build -o qn_cli main.go
```


Usage
------

```
$ export QINIU_BUCKET_NAME="<QINIU_BUCKET_NAME>"
$ export QINIU_BUCKET_URL="<QINIU_BUCKET_URL>"
$ export QINIU_ACCESS_KEY="<QINIU_ACCESS_KEY>"
$ export QINIU_SECRET_KEY="<QINIU_SECRET_KEY>"

$ ./qn_cli *.txt
Upload file 1234.txt successed: http://tmp-images.qiniudn.com/1234.txt
$ ./qn_cli -n test/124.txt 1234.txt
Upload file 1234.txt successed: http://tmp-images.qiniudn.com/test/124.txt

$ ./qn_cli --help
Usage of ./qn_cli:
  -a=true: Auto named saved files
  -d="": Save dirname
  -md5=false: Auto named saved files use MD5 value
  -n="": Save name
  -w=false: Overwrite exists files
```
