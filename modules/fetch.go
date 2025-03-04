package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
)

func Fetch(srcFile string) error {
	fset := token.NewFileSet()
	bz, err := os.ReadFile(srcFile)
	if err != nil {
		return err
	}
	f, err := parser.ParseFile(fset, srcFile, bz, 0)
	if err != nil {
		return err
	}
	mods := make(map[string]string)
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "NewBasicManager" {
					for i, arg := range call.Args {
						switch arg := arg.(type) {
						case *ast.CompositeLit:
							name := arg.Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
							mods[name] = ""
						case *ast.CallExpr:
							name := arg.Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name
							mods[name] = ""
						default:
							spew.Dump(i, arg)
						}
					}
					return false
				}
			}
		}
		return true
	})
	for _, imp := range f.Imports {
		path := imp.Path.Value
		path = path[1 : len(path)-1] // remove quotes
		key := filepath.Base(path)
		if imp.Name != nil {
			key = imp.Name.Name
		}
		if _, ok := mods[key]; ok {
			mods[key] = path
		}
	}
	for k, v := range mods {
		println(k, v)
	}
	return nil
}

func main() {
	err := Fetch(os.Args[1])
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
