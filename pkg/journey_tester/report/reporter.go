// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package report

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"4d63.com/optional"
)

func New(printer Printer) *Reporter {
	return &Reporter{
		printer:   printer,
		StartTime: time.Now(),
		Journeys:  []*JourneyLog{},
		HasError:  false,
	}
}

type Reporter struct {
	StartTime      time.Time
	EndTime        time.Time
	Journeys       []*JourneyLog
	HasError       bool
	FrameworkError *error
	printer        Printer
}

type JourneyLog struct {
	Desc     string
	Filename string
	Steps    []*StepLog
	HasError bool
	printer  Printer

	DurationS  float64
	DurationMs float64
}

type StepLog struct {
	Step       string
	ErrorMsg   string
	Call       APICall
	HasError   bool
	WasSkipped bool

	OptionalTiming *Timing
}

type Timing struct {
	Started time.Time
	Ended   time.Time

	DurationS  float64
	DurationMs float64
}

type APICall struct {
	Request  optional.Optional[string]
	Response optional.Optional[string]
}

func (j *JourneyLog) LogSkippedStep(step string) {
	j.printer.PrintStepResult(step, StepResultSkipped)

	j.Steps = append(j.Steps, &StepLog{
		Step:       step,
		HasError:   false,
		ErrorMsg:   "",
		WasSkipped: true,
	})
}

func (j *JourneyLog) LogStep(step string, call *APICall, err error) {
	j.logStep(step, call, optional.Empty[*Timing](), err)
}

func (j *JourneyLog) LogStepWithTiming(step string, call *APICall, timing *Timing, err error) {
	j.logStep(step, call, optional.Of(timing), err)
}

func (j *JourneyLog) logStep(step string, call *APICall, timing optional.Optional[*Timing], err error) {
	var errMessage string
	result := StepResultPassed
	if err != nil {
		result = StepResultFailed

		j.HasError = true
		errMessage = err.Error()
	}
	j.printer.PrintStepResult(step, result)

	var stepTiming *Timing
	if t, ok := timing.Get(); ok {
		stepTiming = t

		j.DurationS += t.DurationS
		j.DurationMs += t.DurationMs
	}

	sl := &StepLog{
		Step:           step,
		Call:           APICall{},
		HasError:       err != nil,
		ErrorMsg:       errMessage,
		WasSkipped:     false,
		OptionalTiming: stepTiming,
	}
	if call != nil {
		sl.Call = *call
	}
	j.Steps = append(j.Steps, sl)
}

func (r *Reporter) LogFrameworkError(err error) {
	r.printer.PrintFrameworkError(err)
	r.FrameworkError = &err
}

func NewJourneyLog(printer Printer, desc string, filename string) *JourneyLog {
	jl := JourneyLog{
		Desc:     desc,
		Filename: RemoveAbsoluteDirectories(filename),
		Steps:    []*StepLog{},
		printer:  printer,
	}
	return &jl
}

func (r *Reporter) AddJourneyLogs(jls []*JourneyLog) {
	for _, jl := range jls {
		r.Journeys = append(r.Journeys, jl)
		if jl.HasError {
			r.HasError = true
		}
	}
}

func (r *Reporter) OutputReport(printer Printer, reportAssetsDir, reportOutputDir string) error {
	r.EndTime = time.Now()
	WriteReportToStdOut(printer, r)

	err := os.Mkdir(reportOutputDir, FileModeDirOwnerOnly)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating report output folder: %w", err)
	}

	reportDataFile := reportOutputDir + "/report.json"
	printer.PrintMsg(fmt.Sprintf("writing report data to file: %v", reportDataFile))
	err = WriteReportDataToFile(r, reportDataFile)
	if err != nil {
		return fmt.Errorf("error writing report data to file: %w", err)
	}
	err = GenerateReports(printer, r, reportAssetsDir, reportOutputDir)
	if err != nil {
		return fmt.Errorf("error generating report: %w", err)
	}
	return nil
}

const FileModeDirOwnerOnly = 0o755

func RemoveAbsoluteDirectories(filename string) string {
	wd, err := os.Getwd()
	if err != nil {
		return filename
	}
	rel, err := filepath.Rel(wd, filename)
	if err != nil {
		return filename
	}
	return rel
}
