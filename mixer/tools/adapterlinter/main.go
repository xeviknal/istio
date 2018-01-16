package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"sort"
)
var exitCode int

var invalidImportPaths = map[string]string {
	// importing log is bad; instead use env.logger.
	"log":"instead use env.logger",
}

func main() {
	flag.Parse()
	var reports Reports
	if flag.NArg() == 0 {
		reports = doDir(".")
	} else {
		reports = doAllDirs(flag.Args())
	}

	for _, r := range reports {
		reportErr(r.msg)
	}
	os.Exit(exitCode)
}

// error formats the error to standard error, adding program
// identification and a newline
func reportErr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	exitCode = 2
}

func adptLintErrf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "adapterlinter: "+format+"\n", args...)
	exitCode = 2
}

func doAllDirs(args []string) Reports {
	reports := make(Reports, 0)
	for _, name := range args {
		// Is it a directory?
		if fi, err := os.Stat(name); err == nil && fi.IsDir() {
			for _, r := range doDir(name) {
				reports = append (reports, r)
			}
		} else {
			adptLintErrf("not a directory: %s", name)
		}
	}
	sort.Sort(reports)
	return reports
}

func doDir(name string) Reports {
	notests := func(info os.FileInfo) bool {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") &&
			!strings.HasSuffix(info.Name(), "_test.go") {
			return true
		}
		return false
	}
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, name, notests, parser.Mode(0))
	if err != nil {
		adptLintErrf("%v", err)
		return nil
	}
	reports := make(Reports, 0)
	for _, pkg := range pkgs {
		for _, r := range doPackage(fs, pkg) {
			reports = append (reports, r)
		}
	}
	sort.Sort(reports)
	return reports
}

func doPackage(fs *token.FileSet, pkg *ast.Package) Reports {
	v := newVisitor(fs)
	for _, file := range pkg.Files {
		ast.Walk(&v, file)
	}
	return v.reports
}

func newVisitor(fs *token.FileSet) visitor {
	return visitor{
		debugTab: 0,
		fs: fs,
	}
}

type visitor struct {
	debugTab int
	reports Reports
	fs *token.FileSet
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		v.debugTab = v.debugTab - 1
		return nil
	}

	switch d := node.(type) {
	case *ast.GoStmt:
		v.reports = append(v.reports,
			Report{
				d.Pos(),
				fmt.Sprintf("%v:%v:go routines are not allowed inside adapters; " +
					"instead use env.ScheduleWork or env.ScheduleDaemon.",
					v.fs.Position(d.Pos()).Filename, v.fs.Position(d.Pos()).Line),

			})
	case *ast.BasicLit:
		//fmt.Printf("**%s%s\n", strings.Repeat("\t", int(v.debugTab)), d.Value)
	case *ast.ImportSpec:
		if d.Path != nil {
			p := strings.Trim(d.Path.Value, "\"")
			for badImp, alternate := range invalidImportPaths {
				if p == badImp {
					v.reports = append(v.reports,
						Report{
							d.Path.Pos(),
							fmt.Sprintf("%v:%v:\"%s\" import is not allowed; %s.",
								v.fs.Position(d.Path.Pos()).Filename, v.fs.Position(d.Path.Pos()).Line, badImp, alternate),
						})
				}
			}
		}
	}

	//fmt.Printf("%s%T\n", strings.Repeat("\t", int(v.debugTab)), node)
	v.debugTab = v.debugTab + 1

	return v
}

type Report struct {
	pos  token.Pos
	msg  string
}

type Reports []Report

func (l Reports) Len() int           { return len(l) }
func (l Reports) Less(i, j int) bool { return l[i].pos < l[j].pos }
func (l Reports) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
