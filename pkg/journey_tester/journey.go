// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package journey_tester

import (
	"fmt"
	"math"
	"net/http"
	"net/http/cookiejar"
	"runtime"
	"time"

	"4d63.com/optional"
	journey_tester_context "github.com/trustap/journey_tester/pkg/journey_tester/context"
	"github.com/trustap/journey_tester/pkg/journey_tester/load_report"
	"github.com/trustap/journey_tester/pkg/journey_tester/report"
	"github.com/trustap/journey_tester/pkg/journey_tester/runner"
)

func New(desc string, groups ...string) *Journey {
	return newJourney(desc, true, groups...)
}

type Journey struct {
	Desc     string
	Steps    []*Step
	Groups   []string
	FilePath string
	Parallel bool
}

func newJourney(desc string, parallel bool, groups ...string) *Journey {
	var filepath string
	// `runtime.Caller(0)` is the current function, `runtime.Caller(1)` is
	// the calling function.
	_, file, _, ok := runtime.Caller(2)
	if ok {
		filepath = file
	}
	return &Journey{
		Desc:     desc,
		Steps:    []*Step{},
		Groups:   groups,
		FilePath: filepath,
		Parallel: parallel,
	}
}

func NewNonParallel(desc string, groups ...string) *Journey {
	return newJourney(desc, false, groups...)
}

func (j *Journey) AddStep(desc string, conf StepFunc) {
	j.Steps = append(j.Steps, &Step{
		Type: StepTypeRequest,
		Desc: desc,
		Conf: conf,
	})
}

func (j *Journey) AddCleanupStep(desc string, conf StepFunc) {
	j.Steps = append(j.Steps, &Step{
		Type:          StepTypeRequest,
		Desc:          desc,
		Conf:          conf,
		IsCleanUpStep: true,
	})
}

func (j *Journey) AddPolledStep(desc string, timeout string, conf StepFunc) {
	j.Steps = append(j.Steps, &Step{
		Type:            StepTypeRequest,
		Desc:            desc,
		Conf:            conf,
		PollTimeoutConf: optional.Of(timeout),
	})
}

func (j *Journey) AddWaitStep(desc string, duration time.Duration) {
	j.Steps = append(j.Steps, &Step{
		Type:     StepTypeWait,
		Desc:     desc,
		WaitTime: &duration,
	})
}

func runJourney(p report.Printer, setupSteps []*Step, c *journey_tester_context.Context, j *Journey) (*report.JourneyLog, error) {
	p.PrintJourneyStart(j.Desc)
	jl := report.NewJourneyLog(p, j.Desc, j.FilePath)

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create cookie jar: %v", cookieJar)
	}

	client := &http.Client{Jar: cookieJar}

	journeyFailed := false
	for _, step := range append(setupSteps, j.Steps...) {
		switch step.Type {
		case StepTypeWait:
			if step.WaitTime != nil {
				time.Sleep(*step.WaitTime)
			} else {
				jl.LogStep(step.Desc, nil, fmt.Errorf("wait step specified but no wait time provided"))
			}

		case StepTypeRequest:
			conf, err := getConfig(step, c)
			if err != nil {
				if journeyFailed {
					jl.LogSkippedStep(step.Desc)
				} else {
					jl.LogStep(step.Desc, nil, err)
					journeyFailed = true
				}
				continue
			}
			if journeyFailed && !step.IsCleanUpStep {
				jl.LogSkippedStep(step.Desc)
			} else {
				var call *runner.APICall
				var err error

				started := time.Now()
				if step.PollTimeoutConf.IsPresent() {
					call, err = callAPIUntilCallSucceeds(p, client, c, conf, c.Conf.GetDuration(step.PollTimeoutConf.ElseZero()))
				} else {
					call, err = runner.CallAPI(client, c, conf)
				}
				ended := time.Now()
				duration := ended.Sub(started)

				timing := &report.Timing{
					Started:    started,
					Ended:      ended,
					DurationS:  round(float64(duration.Milliseconds()), 3, 1),
					DurationMs: round(float64(duration.Microseconds()), 3, 1),
				}
				jl.LogStepWithTiming(step.Desc, toReporterCall(call), timing, err)
				if err != nil {
					journeyFailed = true
				}
			}

		default:
			jl.LogStep(step.Desc, nil, fmt.Errorf("unknown step type: %v", step.Type))
		}
	}

	p.PrintJourneyResult(j.FilePath, j.Desc, !journeyFailed)

	return jl, nil
}

func round(n float64, decPlaces, keep int) float64 {
	return math.Floor(n/math.Pow10(decPlaces-keep)) / math.Pow10(keep)
}

func callAPIUntilCallSucceeds(
	p report.Printer,
	client *http.Client,
	c *journey_tester_context.Context,
	step runner.Step,
	timeout time.Duration,
) (*runner.APICall, error) {
	p.PrintStepMsg("pre-step: polling until call succeeds")

	end := time.Now().Add(timeout)
	for {
		p.PrintStepMsg("\tchecking...")
		apiCall, err := runner.CallAPI(client, c, step)
		if err == nil {
			p.PrintStepMsg("\tdone: call succeed")
			return apiCall, nil
		}

		if time.Now().After(end) {
			return apiCall, fmt.Errorf("call timed out: %w", err)
		}

		time.Sleep(time.Second)
	}
}

func loadCallAPIUntilCallSucceeds(
	client *http.Client,
	c *journey_tester_context.Context,
	step runner.Step,
	timeout time.Duration,
) (*runner.APICall, error) {
	end := time.Now().Add(timeout)
	var apiCall *runner.APICall
	var err error
	for time.Now().Before(end) {
		time.Sleep(time.Second)
		apiCall, err = runner.CallAPI(client, c, step)
		if err == nil {
			return apiCall, nil
		}
	}
	return apiCall, fmt.Errorf("call timed out: %w", err)
}

// `getConfig` can fail if we attempt to get a value from the context that
// doesn't exist. In this failure case we normally return an error, however,
// this makes the test config quite verbose. To work around this, we panic and
// recover instead.
func getConfig(step *Step, c *journey_tester_context.Context) (conf runner.Step, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error getting step config: %v", rec)
		}
	}()
	conf = step.Conf(c)
	return conf, err
}

func getTimeout(timeout string, c *journey_tester_context.Context) (to time.Duration, err error) {
	defer func() {
		rec := recover()
		if rec != nil {
			err = fmt.Errorf("error getting timeout value: %v", rec)
		}
	}()
	to = time.Second *
		time.Duration(c.Conf.GetInt(timeout))
	return to, err
}

func toReporterCall(runnerCall *runner.APICall) *report.APICall {
	var call *report.APICall
	if runnerCall != nil {
		call = &report.APICall{
			Request:  runnerCall.Request,
			Response: runnerCall.Response,
		}
	}
	return call
}

func toLoadReporterCall(runnerCall *runner.APICall) *load_report.APICall {
	call := &load_report.APICall{
		Request:  runnerCall.Request,
		Response: runnerCall.Response,
	}
	return call
}

// `runSteps` returns a `*LoadJourneyLog` containing the log information
// of each journey step run. Each item in the `LoadJourneyLog.Steps` array
// will hold a value of either `TimedStepLog` for a step that passed or`nil`
// for a step that were skipped/wait.
func runSteps(
	lp load_report.Printer,
	c *journey_tester_context.Context,
	steps []*Step,
	j *Journey,
) (*load_report.LoadJourneyLog, error) {
	jl := load_report.NewJourneyLog(lp, j.Desc, j.FilePath)
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create cookie jar: %v", cookieJar)
	}

	client := &http.Client{Jar: cookieJar}

	journeyFailed := false
	for _, step := range steps {
		var duration time.Duration
		switch step.Type {
		case StepTypeWait:
			if step.WaitTime != nil {
				time.Sleep(*step.WaitTime)
			} else {
				jl.LogStep(
					step.Desc,
					duration,
					nil,
					fmt.Errorf("wait step specified but no wait time provided"),
				)
			}

		case StepTypeRequest:
			if journeyFailed && !step.IsCleanUpStep {
				continue
			}

			conf, err := getConfig(step, c)
			if err != nil {
				if journeyFailed {
					jl.LogSkippedStep(step.Desc)
				} else {
					jl.LogStep(
						step.Desc,
						duration,
						&load_report.APICall{},
						err,
					)
					journeyFailed = true
				}
				continue
			}

			if step.PollTimeoutConf.IsPresent() {
				var call *runner.APICall
				var err error
				to, err := getTimeout(step.PollTimeoutConf.ElseZero(), c)
				if err != nil {
					jl.LogStep(step.Desc, duration, nil, err)
					journeyFailed = true
				}
				call, err = loadCallAPIUntilCallSucceeds(client, c, conf, to)
				if err != nil {
					journeyFailed = true
				}

				var reportCall *load_report.APICall
				if call != nil {
					reportCall = toLoadReporterCall(call)
				}

				jl.LogStep(step.Desc, duration, reportCall, err)
			} else {
				stepStarted := time.Now()
				call, err := runner.CallAPI(client, c, conf)
				duration := time.Since(stepStarted)

				var reportCall *load_report.APICall
				if call != nil {
					reportCall = toLoadReporterCall(call)
				}

				jl.LogStep(step.Desc, duration, reportCall, err)
			}

		default:
			return nil, fmt.Errorf("unknown step type: %v", step.Type)
		}
	}

	return jl, nil
}
