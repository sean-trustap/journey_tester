// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package load_report

import (
	"time"
)

type Printer interface {
	PrintMsg(msg string)
	PrintStepResult(step string, duration time.Duration, result StepResult)
	PrintStepMsg(msg string)
	PrintTimeout(cancelledJourneys int)
	PrintResultSummary(numCompletedJourneys, numStartedJourneys int, hasError, frameworkError bool)
	PrintJourneysParallelStart(name string)
	PrintJourneyStart(activeUserNumber int)
}

type StepResult int

const (
	StepResultSkipped StepResult = iota
	StepResultPassed
	StepResultFailed
)
