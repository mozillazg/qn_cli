/*
七牛本地上传客户端
$ qn_cli --help
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/qiniu/api/conf"
	"github.com/qiniu/api/resumable/io"
	"github.com/qiniu/api/rs"
)

// 生成上传 token
func genToken(bucket string) string {
	policy := rs.PutPolicy{
		Scope:      bucket,
		ReturnBody: `{"bucket": $(bucket),"key": $(key)}`,
	}
	return policy.Token(nil)
}

// 上传本地文件
func uploadFile(localFile, key, uptoken string) (ret io.PutRet, err error) {
	var extra = &io.PutExtra{}

	if key == "" {
		err = io.PutFileWithoutKey(nil, &ret, uptoken, localFile, extra)
	} else {
		err = io.PutFile(nil, &ret, uptoken, key, localFile, extra)
	}

	return
}

func autoFileName(p string) string {
	_, name := path.Split(p)
	return name
}

func main() {
	// 保存名称
	saveName := flag.String("n", "", "Save name")
	saveDir := flag.String("d", "", "Save dirname")
	autoName := flag.Bool("a", false, "Auto named saved files")
	flag.Parse()
	files := flag.Args()

	bucketName := os.Getenv("QINIU_BUCKET_NAME")
	bucketURL := os.Getenv("QINIU_BUCKET_URL")
	accessKey := os.Getenv("QINIU_ACCESS_KEY")
	secretKey := os.Getenv("QINIU_SECRET_KEY")
	key := *saveName

	if len(files) == 0 {
		flag.PrintDefaults()
		fmt.Println("need files: qn_cli FILE [FILE ...]")
		return
	}

	// 配置 accesskey, secretkey
	conf.ACCESS_KEY = accessKey
	conf.SECRET_KEY = secretKey
	// 生成上传 token
	uptoken := genToken(bucketName)

	// 上传文件
	for _, file := range files {
		if *autoName {
			key = autoFileName(file)
		}
		if *saveDir != "" {
			key = path.Join(*saveDir, key)
		}
		ret, err := uploadFile(file, key, uptoken)
		if err != nil {
			fmt.Printf("Upload file %s faied: %s\n", file, err)
		} else {
			fmt.Printf("Upload file %s successed: %s\n", file, bucketURL+ret.Key)
		}
	}
}
