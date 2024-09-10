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
	Name    string
	Urls    string
	Secrets string
	Root    string
)

// var AdminClient *crpc.Client

func main() {
	flag.StringVar(&Name, "name", "filemgr", "server name, regist in crpc")
	flag.StringVar(&Urls, "urls", "127.0.0.1:18083", "crpc address,eg:127.0.0.1:18083,localhost:18083")
	flag.StringVar(&Urls, "secrets", "", "和地址一一对应的密钥eg:")
	flag.StringVar(&Root, "root", ".", "root dir")
	flag.Parse()
	urls := strings.Split(Urls, ",")
	secrets := strings.Split(Secrets, ",")
	if len(urls) != len(secrets) {
		panic(fmt.Sprintf("len(urls) != len(secrets),len(urls):%v,len(secrets):%v", len(urls), len(secrets)))
	}
	for i, url := range urls {
		fmt.Println("listen:", url, Name)
		tmpclient := crpc.Dial(Name, url, options.Client().SetSecret(secrets[i]))
		tmpclient.RegisterName("crpc", &msg{})
	}
	select {}
	// m := &msg{}
	// var res []*FileInfo
	// err := m.ListDir(struct{ Path string }{Path: ""}, &res)
	// fmt.Println(res, err)
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
