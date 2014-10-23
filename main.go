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
func genToken(bucket string, overwrite bool, key string) string {
	policy := rs.PutPolicy{
		Scope:      bucket,
		ReturnBody: `{"bucket": $(bucket),"key": $(key)}`,
	}
	if overwrite {
		policy.Scope = policy.Scope + ":" + key
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

type args struct {
	bucketName  string
	bucketURL   string
	fileSlice   []string
	key         string
	autoName    bool
	autoMD5Name bool
	overwrite   bool
	saveDir     string
	uptoken     string
}

func parse_args() args {
	// 保存名称
	saveName := flag.String("n", "", "Save name")
	saveDir := flag.String("d", "", "Save dirname")
	autoName := flag.Bool("a", true, "Auto named saved files")
	autoMD5Name := flag.Bool("md5", false, "Auto named saved files use MD5 value")
	overwrite := flag.Bool("w", false, "Overwrite exists files")
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
		os.Exit(1)
	}

	// 配置 accesskey, secretkey
	conf.ACCESS_KEY = accessKey
	conf.SECRET_KEY = secretKey

	return args{
		bucketName:  bucketName,
		bucketURL:   bucketURL,
		fileSlice:   fileSlice,
		key:         key,
		autoName:    *autoName,
		autoMD5Name: *autoMD5Name,
		overwrite:   *overwrite,
		saveDir:     *saveDir,
		uptoken:     "",
	}
}

func main() {
	a := parse_args()
	if !a.overwrite {
		// 生成上传 token
		a.uptoken = genToken(a.bucketName, a.overwrite, a.key)
	}

	// 上传文件
	for _, file := range a.fileSlice {
		if a.autoMD5Name && a.key == "" {
			a.key = autoMD5FileName(file)
		} else if a.autoName && a.key == "" {
			a.key = autoFileName(file)
		}
		if a.saveDir != "" {
			a.key = path.Join(a.saveDir, a.key)
		}
		if a.overwrite {
			a.uptoken = genToken(a.bucketName, a.overwrite, a.key)
		}

		// 上传文件
		ret, err := uploadFile(file, a.key, a.uptoken)
		if err != nil {
			fmt.Printf("Upload file %s faied: %s\n", file, err)
		} else {
			fmt.Printf("Upload file %s successed: %s\n",
				file,
				a.bucketURL+ret.Key,
			)
		}
	}
}
