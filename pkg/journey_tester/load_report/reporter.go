// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package load_report

import (
	"fmt"
	"os"
	"time"

	"4d63.com/optional"
	journey_tester_report "github.com/trustap/journey_tester/pkg/journey_tester/report"
)

func New(printer Printer, numberOfParallelActiveUsers, timeout int) *LoadReporter {
	return &LoadReporter{
		printer:                     printer,
		StartTime:                   time.Now(),
		Journeys:                    []*LoadJourneyLog{},
		NumberOfParallelActiveUsers: numberOfParallelActiveUsers,
		Timeout:                     timeout,
		HasError:                    false,
	}
}

type LoadReporter struct {
	StartTime                   time.Time
	EndTime                     time.Time
	NumberOfParallelActiveUsers int
	Timeout                     int
	Journeys                    []*LoadJourneyLog
	HasError                    bool
	FrameworkError              *error
	printer                     Printer
	NumberJourneysStarted       int
}

type LoadJourneyLog struct {
	Desc     string
	Filename string
	Steps    []*TimedStepLog
	HasError bool
	printer  Printer
}

type TimedStepLog struct {
	Desc     string
	Failure  error
	Call     APICall
	Duration time.Duration
}

type APICall struct {
	Request  optional.Optional[string]
	Response optional.Optional[string]
}

func (j *LoadJourneyLog) LogStep(desc string, duration time.Duration, call *APICall, err error) {
	result := StepResultPassed
	if err != nil {
		result = StepResultFailed
	}
	j.printer.PrintStepResult(desc, duration, result)

	sl := &TimedStepLog{
		Desc:     desc,
		Failure:  err,
		Call:     APICall{},
		Duration: duration,
	}
	if call != nil {
		sl.Call = *call
	}
	j.Steps = append(j.Steps, sl)
}

func (j *LoadJourneyLog) LogSkippedStep(desc string) {
	j.printer.PrintStepResult(desc, 0, StepResultSkipped)

	j.Steps = append(j.Steps, &TimedStepLog{
		Desc:     desc,
		Failure:  nil,
		Call:     APICall{},
		Duration: 0,
	})
}

func NewJourneyLog(printer Printer, desc string, filename string) *LoadJourneyLog {
	jl := LoadJourneyLog{
		Desc:     desc,
		Filename: journey_tester_report.RemoveAbsoluteDirectories(filename),
		Steps:    []*TimedStepLog{},
		printer:  printer,
	}
	return &jl
}

func (r *LoadReporter) AddJourneyLogs(jls []*LoadJourneyLog) {
	for _, jl := range jls {
		r.Journeys = append(r.Journeys, jl)
		if jl.HasError {
			r.HasError = true
		}
	}
}

func (r *LoadReporter) AddJourneysQuantity(numJourneysStarted int) {
	r.NumberJourneysStarted = numJourneysStarted
}

func (r *LoadReporter) OutputReport(printer Printer, reportOutputDir string) error {
	r.EndTime = time.Now()

	err := os.Mkdir(reportOutputDir, journey_tester_report.FileModeDirOwnerOnly)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating report output folder: %w", err)
	}

	reportDataFile := reportOutputDir + "/load_report.json"
	completedJourneys := len(r.Journeys)
	printer.PrintResultSummary(
		completedJourneys,
		r.NumberJourneysStarted,
		r.HasError,
		r.FrameworkError != nil,
	)
	printer.PrintMsg(fmt.Sprintf("writing report data to file: %v", reportDataFile))
	err = WriteReportDataToFile(r, reportDataFile)
	if err != nil {
		return fmt.Errorf("error writing report data to file: %w", err)
	}
	return nil
}
