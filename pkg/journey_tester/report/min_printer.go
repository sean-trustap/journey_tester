// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package report

import (
	"fmt"
)

func NewMinimalPrinter() *MinimalPrinter {
	return &MinimalPrinter{}
}

type MinimalPrinter struct{}

func (p *MinimalPrinter) PrintPhase(name string) {
	fmt.Printf("\nrunning %s...\n\n", name)
}

func (p *MinimalPrinter) PrintMsg(msg string) {
}

func (p *MinimalPrinter) PrintError(err string) {
}

func (p *MinimalPrinter) PrintStepMsg(msg string) {
}

func (p *MinimalPrinter) PrintStepResult(step string, result StepResult) {
}

func (p *MinimalPrinter) PrintFrameworkError(err error) {
	fmt.Printf("fatal error: %v\n", err)
}

func (p *MinimalPrinter) PrintJourneyStart(name string) {
}

func (p *MinimalPrinter) PrintJourneyResult(fpath, name string, passed bool) {
	result := "  "
	if passed {
		result = "ok"
	}
	fmt.Printf("\t%s %s", result, name)
	if !passed {
		fmt.Printf(" (%s)", fpath)
	}
	fmt.Println()
}

func (p *MinimalPrinter) PrintResultSummary(hadTestErr, hadFrameworkErr bool) {
	fmt.Println()
	if hadFrameworkErr || hadTestErr {
		fmt.Println("failures:")
	} else {
		fmt.Println("all tests passed")
	}
	fmt.Println()
}

func (p *MinimalPrinter) PrintFailure(failNum int, journey string, step *StepLog) {
	fmt.Printf("\t%d. %s: %s\n", failNum, journey, step.Step)
	fmt.Printf("\t\t%s\n\n", step.ErrorMsg)
}

func (p *MinimalPrinter) PrintReportPath(path string) {
	fmt.Printf("Test report was generated to '%s'\n", path)
}
