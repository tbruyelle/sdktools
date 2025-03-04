package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"

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
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "NewBasicManager" {
					for i, arg := range call.Args {
						switch arg := arg.(type) {
						case *ast.CompositeLit:
							name := arg.Type.(*ast.SelectorExpr).X.(*ast.Ident).Name
							fmt.Println(i, name)
						case *ast.CallExpr:
							name := arg.Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name
							fmt.Println(i, name)
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
	return nil
}

func main() {
	err := Fetch(os.Args[1])
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
