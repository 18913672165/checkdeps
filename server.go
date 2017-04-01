package main

import (
	"flag"
	"log"
	"os"
	"path"
	"strings"
)

var (
	from = flag.String("f", "", "分析的模块名")
	to   = flag.String("o", "", "输出路径")
)

func main() {
	var fromPath string = *from
	if fromPath == "" {
		tmpFrom, err := os.Getwd()
		if err != nil {
			log.Fatal("未指定模块路径且获取当前路径失败")
		}
		log.Println("未指定模块路径,使用当前路径")
		tmpPath := strings.Split(tmpFrom, path.Join(baseDir, "src")+"/")
		if len(tmpPath) == 1 || tmpPath[1] == tmpFrom {
			log.Fatal("错误的模块路径")
		}
		fromPath = tmpPath[1]
	}
	root := path.Join(baseDir, "src", fromPath)
	if err := os.Chdir(root); err != nil {
		log.Fatal("错误的路径名")
	}
	if _, err := buildContext.Import(fromPath, root, 0); err != nil {
		log.Fatalf("无效的模块路径:%s", fromPath)
	}
	var dir string
	var base string
	dir, base = path.Split(*to)
	var filePath string
	var err error
	if dir == "" && base == "" {
		filePath, err = os.Getwd()
		dir, base = path.Split(filePath)
	}

	var respType string
	respType = path.Ext(base)
	if respType == "" {
		respType = "png"
	} else {
		respType = respType[1:]
	}
	if respType != "png" && respType != "csv" {
		log.Fatal("暂不支持的输出格式")
	}
	filename := filePath + "." + respType
	err = processPackage(root, fromPath)
	if err != nil {
		log.Fatalf("error:%v", err)
	}
	err = getDeps(respType, filePath, base)
	if err != nil {
		log.Fatalf("error:%v", err)
	}
	log.Printf("检查成功,可查看文件:%s", filename)
}
