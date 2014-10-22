/*
七牛本地上传客户端
$ qn_cli --help
*/
package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

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

// 自动生成文件名
func autoFileName(p string) string {
	_, name := path.Split(p)
	return name
}
func autoMD5FileName(p string) string {
	oldName := autoFileName(p)
	now := int(time.Now().Nanosecond())
	hash := md5.Sum([]byte(
		strconv.Itoa(now),
	))
	newName := hex.EncodeToString(hash[:]) + "_" + oldName
	return newName
}

func main() {
	// 保存名称
	saveName := flag.String("n", "", "Save name")
	saveDir := flag.String("d", "", "Save dirname")
	autoName := flag.Bool("a", true, "Auto named saved files")
	autoMD5Name := flag.Bool("m", false, "Auto named saved files use MD5 value")
	flag.Parse()
	files := flag.Args()

	bucketName := os.Getenv("QINIU_BUCKET_NAME")
	bucketURL := os.Getenv("QINIU_BUCKET_URL")
	accessKey := os.Getenv("QINIU_ACCESS_KEY")
	secretKey := os.Getenv("QINIU_SECRET_KEY")
	key := *saveName
	fileSlice := []string{}

	// 支持通配符
	for _, file := range files {
		matches, err := filepath.Glob(file)
		if err == nil {
			fileSlice = append(fileSlice, matches...)
		}
	}
	if len(fileSlice) == 0 {
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
	for _, file := range fileSlice {
		if *autoName && key == "" {
			key = autoFileName(file)
		}
		if *autoMD5Name && key == "" {
			key = autoMD5FileName(file)
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
