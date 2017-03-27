package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	host      = flag.String("host", "", "主机名")
	port      = flag.Int("port", 12346, "端口")
	staticDir = flag.String("staticDir", os.Getenv("HOME"), "静态文件存放路径")
)

type errMsg struct {
	Reason string
}

func checkdeps(w http.ResponseWriter, req *http.Request) {
	//从query获取模块名
	pkg := req.URL.Query().Get("pkg")
	//检查query是否为空
	if pkg == "" {
		reason := "query pkg can't be null"
		resp, _ := json.Marshal(&errMsg{Reason: reason})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(resp)
		return
	}
	//检查pkg是否有效
	root := path.Join(baseDir, "src", pkg)
	if err := os.Chdir(root); err != nil {
		reason := "query pkg is an not path"
		resp, _ := json.Marshal(&errMsg{Reason: reason})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(resp)
		return
	}
	_, err := buildContext.Import(pkg, root, 0)
	if err != nil {
		reason := fmt.Sprintf("failed to find package %s", pkg)
		resp, _ := json.Marshal(&errMsg{Reason: reason})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(resp)
		return
	}
	//获取客户端所需类型,支持img和csv,默认是img
	respType := req.URL.Query().Get("type")
	if respType != "img" && respType != "csv" {
		respType = "img"
	}
	//这边可以检查是否已经生成过
	processPackage(root, pkg)
	filename := getDeps(respType)
	f, _ := os.Open(filename)
	defer f.Close()
	if respType == "csv" {
		w.Header().Set("Content-Type", "text/csv")
	} else {
		w.Header().Set("Content-Type", "image/png")
	}
	content, _ := ioutil.ReadAll(f)
	w.Write(content)
}

func main() {
	http.HandleFunc("/checkdeps", checkdeps)
	http.HandleFunc("/cc", test)
	log.Printf("服务器启动,在%s:%d", *host, *port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), nil)
}
