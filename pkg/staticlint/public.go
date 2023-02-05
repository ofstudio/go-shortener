package staticlint

import (
	"code.gitea.io/gitea-vet/checks"
	"github.com/Abirdcfly/dupword"
	"golang.org/x/tools/go/analysis"
)

// PublicAnalyzers возвращает список "публичных" анализаторов.
//
// Возвращаемые анализаторы:
//  1. dupword - проверяет, что в коде и комментариях нет дублирующихся слов.
//  2. imports - проверяет, что все импорты отсортированы правильным образом.
func PublicAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		dupword.NewAnalyzer(),
		checks.Imports,
	}
}
