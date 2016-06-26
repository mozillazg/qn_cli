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
	"io"
	"io/ioutil"
	"log"
	"net/http"
	url2 "net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/kodocli"
)

func Version() (v string) {
	return "0.4.1"
}

var ignorePaths = []string{
	".git", ".hg", ".svn", ".module-cache", ".bin",
}
var tempDir = ""

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%s", *s)
}
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// 生成上传 token
func genUpToken(a *args, c *kodo.Client, key string) string {
	policy := kodo.PutPolicy{
		Scope: a.bucketName,
		// ReturnBody: `{"bucket": $(bucket),"key": $(key)}`,
		DetectMime: 1,
	}
	if key != "" {
		policy.SaveKey = key
		if a.overwrite {
			policy.Scope = policy.Scope + ":" + key
			policy.InsertOnly = 0
		}
	}
	return c.MakeUptoken(&policy)
}

// 上传本地文件
func uploadFile(
	uploader kodocli.Uploader, ctx context.Context, localFile, key, uptoken string) (ret *kodocli.PutRet, err error) {
	ret = &kodocli.PutRet{}
	if key == "" {
		err = uploader.PutFileWithoutKey(ctx, ret, uptoken, localFile, nil)
	} else {
		err = uploader.PutFile(ctx, ret, uptoken, key, localFile, nil)
	}
	return
}

// 自动生成文件名
func autoFileName(p string) (string, string, string) {
	dirname, name := path.Split(p)
	ext := path.Ext(name)
	if tempDir != "" && dirname == tempDir {
		dirname = ""
	}
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

func walkFiles(files []string, ignorePaths []string) (fileSlice []string) {
	for _, file := range files {
		if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
			if pathx, err := downloadFile(file); err == nil {
				file = pathx
			} else {
				continue
			}
		}

		matches, err := filepath.Glob(file)
		_ = "breakpoint"
		if err == nil {

			for _, path := range matches {
				// 遍历目录
				err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						log.Print(err)
						return nil
					}

					// ignore ignorePaths
					for _, i := range ignorePaths {
						p := filepath.Base(path)
						if m, _ := filepath.Match(i, p); m {
							if info.IsDir() {
								return filepath.SkipDir
							}
							return nil
						}
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

	return
}

func finalURL(bucketURL, key string) (url string) {
	u, _ := url2.Parse(bucketURL + key)
	url = u.String()
	return
}

func downloadFile(url string) (pathx string, err error) {
	defer func() {
		if err != nil {
			log.Fatal(err)
		}
	}()

	c := &http.Client{}
	resp, err := c.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	dir, err := ioutil.TempDir("", "qn_cli")
	tempDir = dir
	if err != nil {
		return
	}

	u2, _ := url2.Parse(url)
	name := filepath.Base(u2.Path)
	pathx = filepath.Join(dir, name)
	file, err := os.Create(pathx)
	if err != nil {
		return
	}
	defer file.Close()
	io.Copy(file, resp.Body)
	return
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
	verbose     bool
	maxTasks    int
}

func parseArgs() *args {
	// 保存名称
	saveName := flag.String("n", "", "Save name")
	saveDir := flag.String("d", "", "Save dirname")
	autoName := flag.Bool("a", true, "Auto named saved files")
	autoMD5Name := flag.Bool("md5", false, "Auto named saved files use MD5 value")
	overwrite := flag.Bool("w", true, "Overwrite exists files")
	verbose := flag.Bool("v", false, "Verbose mode")
	version := flag.Bool("V", false, "Version info")
	maxTasks := flag.Int("max-tasks", 5, "Max upload tasks")
	var ignores stringSlice
	flag.Var(&ignores, "i", "ignores")

	flag.Parse()
	if *version {
		fmt.Println("qn_cli " + Version())
		os.Exit(0)
	}

	files := flag.Args()

	bucketName := os.Getenv("QINIU_BUCKET_NAME")
	bucketURL := os.Getenv("QINIU_BUCKET_URL")
	accessKey := os.Getenv("QINIU_ACCESS_KEY")
	secretKey := os.Getenv("QINIU_SECRET_KEY")
	if *verbose {
		fmt.Printf("bucketName: %s\n", bucketName)
		fmt.Printf("bucketURL: %s\n", bucketURL)
		fmt.Printf("accessKey: %s\n", accessKey)
		fmt.Printf("secretKey: %s\n", secretKey)
	}

	key := *saveName
	// 支持通配符
	fileSlice := walkFiles(files, ignorePaths)

	if len(fileSlice) == 0 {
		flag.PrintDefaults()
		fmt.Println("need files: qn_cli FILE [FILE ...]")
		os.Exit(1)
	}

	// 配置 accessKey, secretKey
	kodo.SetMac(accessKey, secretKey)
	if len(ignores) != 0 {
		ignorePaths = append(ignorePaths, ignores...)
	}

	return &args{
		bucketName:  bucketName,
		bucketURL:   bucketURL,
		fileSlice:   fileSlice,
		key:         key,
		autoName:    *autoName,
		autoMD5Name: *autoMD5Name,
		overwrite:   *overwrite,
		saveDir:     *saveDir,
		verbose:     *verbose,
		maxTasks:    *maxTasks,
	}
}

func main() {
	a := parseArgs()
	if a.verbose {
		fmt.Println(a)
	}

	// 定义任务组
	var wg sync.WaitGroup
	cts := 1

	// 上传文件
	for _, file := range a.fileSlice {
		// 正在上传的任务数超出了限制，等待上传完成
		if cts > a.maxTasks {
			wg.Wait()
		}

		// 增加一个任务
		wg.Add(1)
		cts++

		// 使用 goroutine 异步执行上传任务
		go func(file string) {
			defer wg.Done()          // 标记任务完成
			defer func() { cts-- }() // 正在进行的任务数减一

			key := a.key
			zone := 0
			c := kodo.New(zone, nil)
			uploader := kodocli.NewUploader(zone, nil)
			ctx := context.Background()

			if a.autoMD5Name && key == "" {
				key = autoMD5FileName(file)
			} else if a.autoName && key == "" {
				if tempDir != "" && strings.HasPrefix(file, tempDir) {
					key = strings.Split(file, tempDir)[1]
				} else {
					key = file
				}
			}
			if a.saveDir != "" {
				key = path.Join(a.saveDir, key)
			}
			token := genUpToken(a, c, key)

			// 上传文件
			ret, err := uploadFile(uploader, ctx, file, key, token)
			if err != nil {
				if a.verbose {
					fmt.Printf("%s: %s ✕\n", file, err)
				} else {
					fmt.Printf("%s ✕\n", file)
				}
				log.Fatal(err)
			} else {
				url := finalURL(a.bucketURL, ret.Key)
				if a.verbose {
					fmt.Printf("%s: %s ✓\n", file, url)
				} else {
					fmt.Printf("%s\n", url)
				}
			}
		}(file)
	}

	// 等待所有任务完成
	wg.Wait()
}
