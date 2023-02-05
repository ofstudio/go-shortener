package staticlint

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/stylecheck"
)

// StyleCheckAnalyzers - возвращает список анализаторов из пакета stylecheck.
// См https://staticcheck.io/docs/checks#ST
//
// Возвращаемые анализаторы:
//  1. ST1000 - проверяет, что все пакеты имеют комментарий с описанием пакета.
func StyleCheckAnalyzers() []*analysis.Analyzer {
	var result []*analysis.Analyzer
	for _, a := range stylecheck.Analyzers {
		if a.Analyzer.Name == "ST1000" {
			result = append(result, a.Analyzer)
		}
	}
	return result
}
