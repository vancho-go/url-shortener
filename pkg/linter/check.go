package linter

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func Check() {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers, StandardPassesAnalyzers()...)
	analyzers = append(analyzers, StaticCheckSAAnalyzers()...)
	analyzers = append(analyzers, StaticCheckOtherAnalyzers()...)
	analyzers = append(analyzers, CustomAnalyzers()...)
	analyzers = append(analyzers, NoOSExitInMainFuncAnalyzer)

	multichecker.Main(analyzers...)
}
