// 文件管理器
// 读
// 写
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndsky1003/crpc"
	"github.com/ndsky1003/crpc/comm"
	"github.com/ndsky1003/crpc/dto"
	"github.com/ndsky1003/crpc/options"
)

var (
	Names   string
	Urls    string
	Secrets string
	Root    string
)

var (
	names   []string
	urls    []string
	secrets []string
	name0   string
	secret0 string
)

// var AdminClient *crpc.Client

func main() {
	flag.StringVar(&Names, "names", "filemgr", "server name, regist in crpc,和地址一一对应,如果只有一个的话,就是通配全部")
	flag.StringVar(&Urls, "urls", "127.0.0.1:18083", "crpc address,eg:127.0.0.1:18083,localhost:18083")
	flag.StringVar(&Secrets, "secrets", "", "和地址一一对应的密钥eg:")
	flag.StringVar(&Root, "root", ".", "root dir")
	flag.Parse()
	names = strings.Split(Names, ",")
	urls = strings.Split(Urls, ",")
	secrets = strings.Split(Secrets, ",")
	if len(names) == 0 {
		panic("names is empty")
	}
	if len(urls) == 0 {
		panic("urls is empty")
	}
	name0 = names[0]
	secret0 = secrets[0]

	for i, url := range urls {
		name, secret := getNameAndSecret(i)
		fmt.Printf("listen:%v,name:%v", url, name)
		tmpclient := crpc.Dial(name, url, options.Client().SetSecret(secret))
		tmpclient.RegisterName("crpc", &msg{})
	}
	select {}
}

func getNameAndSecret(i int) (name string, secret string) {
	if i >= len(names) {
		name = name0
	} else {
		name = names[i]
	}

	if i >= len(secrets) {
		secret = secret0
	} else {
		secret = secrets[i]
	}
	return
}

type msg struct {
}

func (*msg) ListDir(req struct{ Path string }, res *[]*FileInfo) error {
	if strings.HasPrefix(req.Path, ".") {
		return fmt.Errorf("no access in path {{%v}}", req.Path)
	}
	dir := filepath.Join(Root, req.Path)
	if dir == "" {
		dir = "."
	}
	if res == nil {
		res = &[]*FileInfo{}
	}
	fs, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range fs {
		name := file.Name()
		f := &FileInfo{Name: name, IsDir: file.IsDir()}
		if !file.IsDir() {
			f.Ext = filepath.Ext(name)
		}
		*res = append(*res, f)
	}

	return nil
}

func (*msg) Mkdir(req struct{ Path string }, res *int) error {
	if strings.HasPrefix(req.Path, ".") {
		return fmt.Errorf("no access in path {{%v}}", req.Path)
	}
	if strings.HasPrefix(req.Path, "/") {
		return fmt.Errorf("path:%v not startwith:/", req.Path)
	}
	if req.Path == "" {
		return fmt.Errorf("path is empty")
	}
	dir := filepath.Join(Root, req.Path)
	return os.MkdirAll(dir, 0777)
}

func (*msg) SaveFile(req *dto.FileBody, _ *int) error {
	tmp_path := filepath.Join(Root, req.Filename)
	if req.ChunksIndex == 0 {
		if _, err := os.Stat(tmp_path); err == nil {
			return fmt.Errorf("file exist:%v", tmp_path)
		} else {
			if !os.IsNotExist(err) {
				return err
			}
		}
	}
	f, err := comm.GetWriteFile(req.ChunksIndex, tmp_path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(req.Data)
	if err != nil {
		return err
	}
	return nil
}

type FileInfo struct {
	Name  string
	IsDir bool
	Ext   string
}

func (this *FileInfo) String() string {
	return fmt.Sprintf("%+v", *this)
}
