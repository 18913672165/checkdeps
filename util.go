package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
)

var (
	pkgs         map[string]*build.Package //模块路径和模块对象的关系
	deps         map[string][]string       //每个模块所依赖的模块,去重前
	depReal      map[string][]string       //每个模块所依赖的模块,去重后
	depby        map[string]int            //被依赖
	ids          map[string]int            //模块名和dot label的编号关系
	nextId       int                       //dot label唯一编号,递增
	buildContext = build.Default           //获取模块的依赖和自身路径
	baseDir      string
)

func init() {
	pkgs = make(map[string]*build.Package)
	deps = make(map[string][]string)
	depReal = make(map[string][]string)
	ids = make(map[string]int)
	flag.Parse()
	baseDir = os.Getenv("GOPATH")
	if baseDir == "" {
		log.Fatal("ENV GOPATH must be set")
	}
}

func getDeps(respType string) (fileName string) {
	for k, v := range pkgs {
		tmp := strings.Split(k, "/vendor/")
		newk := ""
		if len(tmp) == 2 {
			newk = tmp[1]
		} else {
			newk = tmp[0]
		}
		pkgs[newk] = v
		deps[newk] = deps[k]
		if newk != k {
			delete(pkgs, k)
			delete(deps, k)
		}
	}
	for k, v := range deps {
		sort.Strings(v)
		tmplen := len(v)
		ret := []string{}
		for i := 0; i < tmplen; i++ {
			if i > 0 && v[i-1] == v[i] {
				continue
			}
			if v[i] == k {
				continue
			}
			ret = append(ret, v[i])
		}
		deps[k] = ret
	}
	pkgKeys := []string{}
	for k := range pkgs {
		pkgKeys = append(pkgKeys, k)
	}
	for _, pkgName := range pkgKeys {
		depReal[pkgName] = countDepnum(pkgName)
	}
	for k, v := range depReal {
		sort.Strings(v)
		tmplen := len(v)
		ret := []string{}
		for i := 0; i < tmplen; i++ {
			if i > 0 && v[i-1] == v[i] {
				continue
			}
			if v[i] == k {
				continue
			}
			ret = append(ret, v[i])
		}
		depReal[k] = ret
	}
	sort.Strings(pkgKeys)
	if respType == "csv" {
		fileName = path.Join(*staticDir, "out.cv")
		file, _ := os.Create(fileName)
		defer file.Close()
		tmplist := [][]string{}
		for k, v := range depReal {
			tmplist = append(tmplist, []string{k, strconv.Itoa(len(v))})
		}
		for i := 0; i < len(tmplist)-1; i++ {
			for j := i + 1; j < len(tmplist); j++ {
				cur, _ := strconv.Atoi(tmplist[j][1])
				pri, _ := strconv.Atoi(tmplist[i][1])
				if cur < pri {
					tmp := tmplist[i]
					tmplist[i] = tmplist[j]
					tmplist[j] = tmp
				}
			}
		}
		for _, v := range tmplist {
			file.WriteString(v[0] + "\t" + v[1] + "\n")
		}

		file.Sync()
		file.Close()
		return
	}
	fileTmp := path.Join(*staticDir, "out.tmp")
	file, _ := os.Create(fileTmp)
	defer file.Close()
	file.WriteString(fmt.Sprint("digraph godep {\n"))
	for _, pkgName := range pkgKeys {
		pkg := pkgs[pkgName]
		pkgId := getId(pkgName)
		var color string = "paleturquoise"
		file.WriteString(fmt.Sprintf("_%d [label=\"%s deps:%d\" style=\"filled\" color=\"%s\"];\n", pkgId, pkgName, len(depReal[pkgName]), color))
		if pkg.Goroot {
			continue
		}
		for _, imp := range deps[pkgName] {
			impPkg := pkgs[imp]
			if impPkg == nil {
				continue
			}
			impId := getId(imp)
			file.WriteString(fmt.Sprintf("_%d -> _%d;\n", pkgId, impId))

		}

	}
	file.WriteString(fmt.Sprint("}\n"))
	fileName = path.Join(*staticDir, "out.png")
	fileReal, _ := os.Create(fileName)
	defer fileReal.Close()
	cmd := exec.Command("dot", "-Tpng", fileTmp)
	val, _ := cmd.Output()
	fileReal.Write(val)
	return

}

func countDepnum(pkgName string) []string {
	var rst []string
	if deplist, ok := deps[pkgName]; ok {
		for _, i := range deplist {
			tmp := countDepnum(i)
			rst = append(rst, i)
			for _, j := range tmp {
				rst = append(rst, j)
			}
		}
		if len(deplist) == 0 {
			rst = append(rst, pkgName)
		}
	} else {
		rst = append(rst, pkgName)
	}
	return rst

}

func processPackage(root string, pkgName string) error {
	pkg, err := buildContext.Import(pkgName, root, 0)
	if err != nil {
		return fmt.Errorf("failed to import %s:%s", pkgName, err)
	}
	pkgs[pkg.ImportPath] = pkg
	if pkg.Goroot {
		return nil
	}
	imps := getImports(pkg)
	for _, imp := range imps {
		checkPkg, _ := buildContext.Import(imp, root, 0)
		if _, ok := pkgs[imp]; ok {
			deps[pkg.ImportPath] = append(deps[pkg.ImportPath], imp)
			continue
		}
		if _, ok := pkgs[checkPkg.ImportPath]; ok {
			deps[pkg.ImportPath] = append(deps[pkg.ImportPath], imp)
			continue
		}
		if !checkPkg.Goroot {
			processPackage(root, imp)
			deps[pkg.ImportPath] = append(deps[pkg.ImportPath], imp)
		}
	}
	return nil
}

func getImports(pkg *build.Package) []string {
	allImports := pkg.Imports
	var imports []string
	found := make(map[string]struct{})
	for _, imp := range allImports {
		if imp == pkg.ImportPath {
			continue
		}
		if _, ok := found[imp]; ok {
			continue
		}
		found[imp] = struct{}{}
		imports = append(imports, imp)
	}
	return imports

}

func getId(name string) int {
	id, ok := ids[name]
	if !ok {
		id = nextId
		nextId++
		ids[name] = id
	}
	return id
}
