qn\_cli
======

Qiniu upload client by Go


Installation
------------

```
$ go get -u github.com/qiniu/api
$ go get -u github.com/mozillazg/qn_cli
```


Usage
------

```
$ export QINIU_BUCKET_NAME="<QINIU_BUCKET_NAME>"
$ export QINIU_BUCKET_URL="<QINIU_BUCKET_URL>"
$ export QINIU_ACCESS_KEY="<QINIU_ACCESS_KEY>"
$ export QINIU_SECRET_KEY="<QINIU_SECRET_KEY>"

$ qn_cli 1234.txt
http://tmp-images.qiniudn.com/1234.txt
$ qn_cli *.txt
http://tmp-images.qiniudn.com/1234.txt
http://tmp-images.qiniudn.com/2345.txt
$ qn_cli -n test/124.txt 1234.txt
http://tmp-images.qiniudn.com/test/124.txt
$ qn_cli -d test *.txt
http://tmp-images.qiniudn.com/test/1234.txt
http://tmp-images.qiniudn.com/test/2345.txt


$ qn_cli --help
Usage of qn_cli:
  -a=true: Auto named saved files
  -d="": Save dirname
  -md5=false: Auto named saved files use MD5 value
  -n="": Save name
  -v=false: Verbose mode
  -w=false: Overwrite exists files
```
