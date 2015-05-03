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
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
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
func autoFileName(p string) (string, string, string) {
	dirname, name := path.Split(p)
	ext := path.Ext(name)
	return dirname, name, ext
}
func autoMD5FileName(p string) string {
	dirname, oldName, ext := autoFileName(p)
	now := int(time.Now().Nanosecond())
	hash := md5.Sum([]byte(
		strconv.Itoa(now),
	))
	newName := dirname + oldName + "_" + hex.EncodeToString(hash[:]) + ext
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
	verbose     bool
}

func parse_args() args {
	// 保存名称
	saveName := flag.String("n", "", "Save name")
	saveDir := flag.String("d", "", "Save dirname")
	autoName := flag.Bool("a", true, "Auto named saved files")
	autoMD5Name := flag.Bool("md5", false, "Auto named saved files use MD5 value")
	overwrite := flag.Bool("w", false, "Overwrite exists files")
	verbose := flag.Bool("v", false, "Verbose mode")
	flag.Parse()
	files := flag.Args()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

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

			for _, path := range matches {
				// 遍历目录
				err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						log.Print(err)
						return nil
					}
					if info.IsDir() {
						return nil
					}
					fileSlice = append(fileSlice, path)
					return nil
				})
				if err != nil {
					log.Print(err)
				}
			}
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
		verbose:     *verbose,
	}
}

func main() {
	a := parse_args()
	if !a.overwrite {
		// 生成上传 token
		a.uptoken = genToken(a.bucketName, a.overwrite, a.key)
	}
	// 定义任务组
	var wg sync.WaitGroup

	// 上传文件
	for _, file := range a.fileSlice {
		// 增加一个任务
		wg.Add(1)
		// 使用 goroutine 异步执行上传任务
		go func(file string) {
			defer wg.Done() // 标记任务完成
			key := a.key
			uptoken := a.uptoken

			if a.autoMD5Name && key == "" {
				key = autoMD5FileName(file)
			} else if a.autoName && key == "" {
				key = file
			}
			if a.saveDir != "" {
				key = path.Join(a.saveDir, key)
			}
			if a.overwrite {
				uptoken = genToken(a.bucketName, a.overwrite, key)
			}

			// 上传文件
			ret, err := uploadFile(file, key, uptoken)
			if err != nil {
				if a.verbose {
					fmt.Printf("Upload file %s faied: %s\n", file, err)
				}
			} else {
				url := a.bucketURL + ret.Key
				if a.verbose {
					fmt.Printf("Upload file %s successed: %s\n",
						file,
						url,
					)
				} else {
					fmt.Println(url)
				}
			}
		}(file)
	}

	// 等待所有任务完成
	wg.Wait()
}
