package main

import (
	"musthave-metrics/cmd/staticlint/osexitanalyzer"
	"regexp"

	"github.com/alexkohler/nakedret/v2"
	"github.com/breml/bidichk/pkg/bidichk"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	mychecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		structtag.Analyzer,
		osexitanalyzer.Analyzer,
		bidichk.NewAnalyzer(),
		nakedret.NakedReturnAnalyzer(3, true),
	}

	re := regexp.MustCompile(`(SA*|S1*|ST1*)`)
	for _, v := range staticcheck.Analyzers {
		if re.MatchString(v.Analyzer.Name) {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
