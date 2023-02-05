package staticlint

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/staticcheck"
)

// StaticCheckAnalyzers возвращает список всех SA-** анализаторов из пакета staticcheck.
//
// См https://staticcheck.io/docs/checks#SA
func StaticCheckAnalyzers() []*analysis.Analyzer {
	var result []*analysis.Analyzer
	for _, a := range staticcheck.Analyzers {
		result = append(result, a.Analyzer)
	}
	return result
}
