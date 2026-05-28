// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package report

import (
	"fmt"
	"log"
)

func NewLogPrinter() *LogPrinter {
	return &LogPrinter{}
}

type LogPrinter struct{}

func (p *LogPrinter) PrintPhase(name string) {
	log.Printf("\n\n\t%s phase\n\n", name)
}

func (p *LogPrinter) PrintMsg(msg string) {
	log.Printf("%s\n", msg)
}

func (p *LogPrinter) PrintStepMsg(msg string) {
	log.Printf("\t%s\n", msg)
}

func (p *LogPrinter) PrintError(err string) {
	log.Printf("%v\n", err)
}

func (p *LogPrinter) PrintStepResult(step string, result StepResult) {
	var msg string
	switch result {
	case StepResultPassed:
		msg = fmt.Sprintf("%spassed%s", greenColor, noColor)
	case StepResultFailed:
		msg = fmt.Sprintf("%sfailed%s", redColor, noColor)
	case StepResultSkipped:
		msg = "skipped"
	}
	log.Printf("\tstep: %s -> %s\n", step, msg)
}

const (
	redColor   = "\033[0;31m"
	greenColor = "\033[0;32m"
	noColor    = "\033[0m"
)

func (p *LogPrinter) PrintFrameworkError(err error) {
	log.Printf("fatal error: %s%v%s\n", redColor, err, noColor)
}

func (p *LogPrinter) PrintJourneyStart(name string) {
	log.Printf("starting journey: %v\n", name)
}

func (p *LogPrinter) PrintJourneyResult(fpath, name string, passed bool) {
}

func (p *LogPrinter) PrintResultSummary(hadTestErr, hadFrameworkErr bool) {
	log.Printf("\n\n\ttest run finished\n\n")

	msg := "all tests have passed :)"
	if hadFrameworkErr {
		msg = "result: framework error"
	}
	if hadTestErr {
		msg = "result: some tests have failed"
	}
	log.Printf("\n\n\t%s\n\n", msg)
}

func (p *LogPrinter) PrintFailure(failNum int, journey string, step *StepLog) {
	log.Printf("\n\n\t*** failure number %v ***\n\n", failNum)
	log.Printf("failed step: %v", step.Step)
	log.Printf("failure message: %v", step.ErrorMsg)
	log.Printf("failure request / response:\n\n")
	if step.Call.Request.IsPresent() {
		log.Print(step.Call.Request.ElseZero())
	}
	if step.Call.Response.IsPresent() {
		log.Print(step.Call.Response.ElseZero())
	}
}

func (p *LogPrinter) PrintReportPath(path string) {
	log.Printf("\ngenerated a test report: file://%s\n", path)
}
