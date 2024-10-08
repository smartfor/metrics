package staticlint

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var NoExitAnalyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "Check for direct os.Exit calls in main package and in main function",
	Run:  runNoExit,
}

const failMessage = "не используйте вызов os.Exit напрямую в main функции пакета main"

// runNoExit выполняет анализ и проверяет, что в функции main пакета main нет прямых вызовов os.Exit.
func runNoExit(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if fn.Name.Name != "main" {
					continue
				}

				checkMainFunction(pass, fn)
			}
		}
	}

	return nil, nil
}

func checkMainFunction(pass *analysis.Pass, fn *ast.FuncDecl) {
	ast.Inspect(fn, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if sel.Sel.Name == "Exit" {
					pass.Reportf(
						call.Pos(),
						failMessage,
					)
				}
			}
		}

		return true
	})
}
