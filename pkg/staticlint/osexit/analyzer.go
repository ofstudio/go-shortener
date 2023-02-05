package osexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer - анализатор, запрещающий использовать вызов os.Exit в функции main пакета main
var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check for os.Exit() calls in main function main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	// Не анализируем не-main пакеты и тесты.
	if pass.Pkg.Name() != "main" || strings.HasSuffix(pass.Pkg.Path(), ".test") {
		return nil, nil
	}

	// Перебираем все файлы пакета
	for _, file := range pass.Files {
		// Анализируем AST файла
		ast.Inspect(file, func(node ast.Node) bool {
			// Только вызовы функций
			if node, ok := node.(*ast.CallExpr); ok {
				if selector, ok := node.Fun.(*ast.SelectorExpr); ok {
					// Только функция Exit
					if selector.Sel.Name == "Exit" {
						if ident, ok := selector.X.(*ast.Ident); ok {
							// Только из пакета os
							if ident.Name == "os" {
								pass.Reportf(node.Pos(), "os.Exit call in main function of main package")
							}
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
