package linter

import (
	"flag"
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// NoOSExitInMainFuncAnalyzer проверяет наличие вызова os.Exit в функции main.
var NoOSExitInMainFuncAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "forbids calls to os.Exit in the main function",
	Run:  run,
	URL:  "", // не используется
	Flags: *flag.NewFlagSet(
		"",
		flag.ContinueOnError), // не используется
	RunDespiteErrors: false, // не используется
	Requires:         nil,   // не используется
	ResultType:       nil,   // не используется
	FactTypes:        nil,   // не используется
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Пропускаем файлы, не принадлежащие пакету main
		if pass.Pkg.Name() != "main" {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			// Ищем объявление функции
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true // продолжаем поиск, если текущий узел не объявление функции
			}

			// Проверяем на функцию main
			if funcDecl.Name.Name == "main" {
				// проверяем, содержит ли функция main вызов os.Exit
				ast.Inspect(funcDecl, func(n ast.Node) bool {
					callExpr, ok := n.(*ast.CallExpr)
					if !ok {
						return true // не вызов функции
					}
					selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
					if !ok {
						return true // не вызов метода
					}
					pkgIdent, ok := selExpr.X.(*ast.Ident)
					if !ok {
						return true // не идентификатор пакета
					}
					if pkgIdent.Name == "os" && selExpr.Sel.Name == "Exit" {
						// найден вызов os.Exit в функции main
						pass.Reportf(callExpr.Pos(), "call to os.Exit in function main is forbidden")
					}
					return true
				})
			}
			return true
		})
	}
	return nil, nil
}
