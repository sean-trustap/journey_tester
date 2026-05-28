// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package report

type Printer interface {
	PrintPhase(name string)
	PrintMsg(msg string)
	PrintError(err string)
	PrintStepMsg(msg string)
	PrintStepResult(step string, result StepResult)
	PrintFrameworkError(err error)
	PrintJourneyStart(name string)
	PrintJourneyResult(fpath, name string, passed bool)
	PrintResultSummary(hadTestErr, hadFrameworkErr bool)
	PrintFailure(failNum int, journey string, step *StepLog)
	PrintReportPath(path string)
}

type StepResult int

const (
	StepResultSkipped StepResult = iota
	StepResultPassed
	StepResultFailed
)
