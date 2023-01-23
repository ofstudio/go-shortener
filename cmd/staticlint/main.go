package main

import (
	"github.com/ofstudio/go-shortener/pkg/staticlint"
	"github.com/ofstudio/go-shortener/pkg/staticlint/osexit"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	// 1. Стандартные статические анализаторы пакета `golang.org/x/tools/go/analysis/passes`
	// 2. Все анализаторы SA из пакета `staticcheck.io`
	analyzers := append(staticlint.MultiCheckAnalyzers(), staticlint.StaticCheckAnalyzers()...)

	// 3. Не менее одного анализатора остальных классов пакета `staticcheck.io`
	analyzers = append(analyzers, staticlint.StyleCheckAnalyzers()...)

	// 4. Два или более любых публичных анализаторов на ваш выбор
	analyzers = append(analyzers, staticlint.PublicAnalyzers()...)

	// 5. Собственный анализатор, запрещающий использовать прямой вызов `os.Exit`
	//    в функции `main` пакета `main`
	analyzers = append(analyzers, osexit.Analyzer)

	multichecker.Main(analyzers...)
}
