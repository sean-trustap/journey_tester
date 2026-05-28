// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package journey_tester

import (
	"context"
	"fmt"
	"sync"
	"time"

	"4d63.com/optional"
	"github.com/trustap/journey_tester/pkg/hashset"
	journey_tester_context "github.com/trustap/journey_tester/pkg/journey_tester/context"
	"github.com/trustap/journey_tester/pkg/journey_tester/load_report"
	"github.com/trustap/journey_tester/pkg/journey_tester/report"
)

func Main(journeys *Journeys) (testResult, error) {
	fc, err := parseFlags()
	if err != nil {
		return TestResultFail, fmt.Errorf("couldn't parse flags: %w", err)
	}

	printer := fc.printer
	c, err := journey_tester_context.NewContext(
		fc.configFile,
		"configs/journey_tester.sample.yaml",
		printer,
	)
	if err != nil {
		return TestResultFail, fmt.Errorf("couldn't create new context: %w", err)
	}

	tester := &journeyTester{printer: printer, loadPrinter: fc.loadPrinter}
	r := report.New(printer)
	lr := &load_report.LoadReporter{}

	if fc.loadTesting {
		var numThreads, timeout int
		var ok bool
		if numThreads, ok = fc.numThreads.Get(); !ok {
			return TestResultFail, fmt.Errorf("please specify the 'numThreads' parameter")
		}
		if timeout, ok = fc.timeout.Get(); !ok {
			return TestResultFail, fmt.Errorf("please specify the 'timeout' parameter")
		}
		lr = load_report.New(fc.loadPrinter, numThreads, timeout)
	}

	result, err := runPhases(r, lr, c, fc, journeys, tester)
	if err != nil {
		return TestResultFail, fmt.Errorf("couldn't run phases: %w", err)
	}
	if fc.loadTesting {
		err = lr.OutputReport(fc.loadPrinter, "assets/journey_tester_report", "target/journey_tester")
		if err != nil {
			return TestResultFail, fmt.Errorf("couldn't output report: %w", err)
		}
	} else {
		err = r.OutputReport(printer, "assets/journey_tester_report", "target/journey_tester")
		if err != nil {
			return TestResultFail, fmt.Errorf("couldn't output report: %w", err)
		}
	}

	return result, nil
}

type Journeys struct {
	DataJourneys   *DataJourneys
	TestJourneys   []*Journey
	BeforeEachTest []*Step
}

type DataJourneys struct {
	Create []*Journey
	Verify []*Journey
	// Data loading journeys are run once, in sequence (not parallel),
	// before running tests. This is done in order to populate the test
	// context with non-expiring test data that will be used throughout
	// testing, such as user IDs. Expiring test data, such as access tokens,
	// should be added to the test context using `BeforeEachJourney`
	// instead.
	Load   []*Journey
	Delete []*Journey
}

func runPhases(
	r *report.Reporter,
	lr *load_report.LoadReporter,
	c *journey_tester_context.Context,
	fc *JourneyTesterConfig,
	journeys *Journeys,
	jt *journeyTester,
) (testResult, error) {
	filteredJourneys := filterJourneys(journeys.TestJourneys, fc)
	if len(filteredJourneys) == 0 {
		return TestResultFail, fmt.Errorf("file(s) does not exist. Provided 'run' parameter is invalid")
	}

	if len(filteredJourneys) != 1 && fc.loadTesting {
		return TestResultFail, fmt.Errorf("please specify exactly one test to run for load testing")
	}

	phases := []struct {
		name       string
		skip       bool
		js         []*Journey
		numThreads optional.Optional[int]
		main       bool
		ignoreFail bool
	}{
		{
			name:       "external data pre-deletion",
			skip:       !fc.preDeleteData,
			js:         journeys.DataJourneys.Delete,
			ignoreFail: true,
		},
		{
			name: "external data creation",
			skip: !fc.createData,
			// We use a single thread for data creation to force the
			// order of execution.
			//
			// TODO Consider other ways of allowing ordering for
			// data journeys, while still allowing for parallelism.
			numThreads: optional.Of(1),
			js:         journeys.DataJourneys.Create,
		},
		{
			name: "external data verification",
			skip: fc.skipVerify,
			js:   journeys.DataJourneys.Verify,
		},
		{
			name: "journey tests",
			skip: fc.skipTesting,
			js:   filteredJourneys,
			main: true,
		},
		{
			name: "external data deletion",
			skip: !fc.deleteData,
			js:   journeys.DataJourneys.Delete,
		},
	}

	for _, phase := range phases {
		if phase.skip {
			continue
		}
		jt.printer.PrintPhase(phase.name)

		if phase.main {
			jls, err := jt.runJourneysInSequence(
				c,
				true,
				[]*Step{},
				journeys.DataJourneys.Load,
			)
			if err != nil {
				return TestResultFail, fmt.Errorf("couldn't load external data: %w", err)
			}
			r.AddJourneyLogs(jls)

			if anyJourneyHasError(jls) {
				return TestResultFail, nil
			}

			if fc.loadTesting {
				if fc.numThreads == nil {
					return TestResultFail, fmt.Errorf("please specify the 'numThreads' parameter")
				}
				if fc.timeout == nil {
					return TestResultFail, fmt.Errorf("please specify the 'timeout' parameter")
				}

				timeout, numParallelActiveUsers := fc.timeout, fc.numThreads
				jls, jq, errs := jt.runJourneyInParallel(
					c,
					journeys.BeforeEachTest,
					phase.js[0],
					numParallelActiveUsers.ElseZero(),
					timeout.ElseZero(),
				)
				if len(errs) > 0 {
					return TestResultFail, fmt.Errorf("couldn't run journeys: %v", errs)
				}
				lr.AddJourneyLogs(jls)
				lr.AddJourneysQuantity(jq)
			} else {
				jls, errs := jt.runJourneys(c, journeys.BeforeEachTest, fc.numThreads, phase.js)
				if len(errs) > 0 {
					return TestResultFail, fmt.Errorf("couldn't run journeys: %v", errs)
				}
				r.AddJourneyLogs(jls)

				if anyJourneyHasError(jls) {
					return TestResultFail, nil
				}
			}
		} else {
			numThreads := fc.numThreads
			if n, ok := phase.numThreads.Get(); ok {
				numThreads = optional.Of(n)
			}

			jls, errs := jt.runJourneys(c, []*Step{}, numThreads, phase.js)
			if len(errs) > 0 {
				return TestResultFail, fmt.Errorf("couldn't run journeys: %v", errs)
			}

			if anyJourneyHasError(jls) {
				if !phase.ignoreFail {
					r.AddJourneyLogs(jls)
					return TestResultFail, nil
				}

				fc.printer.PrintMsg("failures in this phase were ignored")
			} else {
				r.AddJourneyLogs(jls)
			}
		}
	}

	return TestResultPass, nil
}

type testResult int

const (
	TestResultPass testResult = iota
	TestResultFail
)

func anyJourneyHasError(jls []*report.JourneyLog) bool {
	for _, jl := range jls {
		if jl.HasError {
			return true
		}
	}
	return false
}

type journeyTester struct {
	printer     report.Printer
	loadPrinter load_report.Printer
}

func (jt *journeyTester) runJourneys(
	c *journey_tester_context.Context,
	setupSteps []*Step,
	numThreads optional.Optional[int],
	js []*Journey,
) ([]*report.JourneyLog, []error) {
	parallel := []*Journey{}
	nonParallel := []*Journey{}
	for _, j := range js {
		if j.Parallel {
			parallel = append(parallel, j)
		} else {
			nonParallel = append(nonParallel, j)
		}
	}

	jls, errs := jt.runJourneysInParallel(c, numThreads, setupSteps, parallel)
	if len(errs) != 0 {
		return nil, errs
	}

	seqJls, err := jt.runJourneysInSequence(c, false, setupSteps, nonParallel)
	if err != nil {
		return nil, []error{err}
	}

	return append(jls, seqJls...), nil
}

func (jt *journeyTester) runJourneysInSequence(
	ctx *journey_tester_context.Context,
	reuseContext bool,
	setupSteps []*Step,
	js []*Journey,
) ([]*report.JourneyLog, error) {
	jls := []*report.JourneyLog{}

	for _, j := range js {
		c := ctx
		if !reuseContext {
			c = ctx.Copy()
		}

		jl, err := runJourney(jt.printer, setupSteps, c, j)
		if err != nil {
			return nil, fmt.Errorf("failed to run journey: %w", err)
		}
		jls = append(jls, jl)
	}

	return jls, nil
}

func (jt *journeyTester) runJourneysInParallel(
	ctx *journey_tester_context.Context,
	optionalNumThreads optional.Optional[int],
	setupSteps []*Step,
	js []*Journey,
) ([]*report.JourneyLog, []error) {
	var jls []*report.JourneyLog
	var syncErrors sync.Map

	numThreads := len(js)
	if optionalNumThreads.IsPresent() && optionalNumThreads.ElseZero() < numThreads {
		numThreads = optionalNumThreads.ElseZero()
	}
	threads := make(chan struct{}, numThreads)

	if numThreads < len(js) {
		for i := 0; i < numThreads; i++ {
			threads <- struct{}{}
		}
	}

	for _, j := range js {
		if numThreads < len(js) {
			<-threads
		}

		go func(j *Journey) {
			defer func() {
				threads <- struct{}{}
			}()

			c := ctx.Copy()

			jl, err := runJourney(jt.printer, setupSteps, c, j)
			if err != nil {
				syncErrors.Store(j.Desc, err)
				return
			}
			jls = append(jls, jl)
		}(j)
	}

	for i := 0; i < numThreads; i++ {
		<-threads
	}

	var errors []error
	syncErrors.Range(func(key, value any) bool {
		errors = append(errors, value.(error))
		return true
	})
	return jls, errors
}

func filterJourneys(js []*Journey, config *JourneyTesterConfig) []*Journey {
	filteredJourneys := make([]*Journey, 0, len(js))

	for _, testJourney := range js {
		testJourneyGroups := hashset.SetFromSlice(testJourney.Groups)

		isOutsideGroups := len(config.groups) > 0 && !config.groups.HasAny(testJourneyGroups)
		if isOutsideGroups || config.skipGroups.HasAny(testJourneyGroups) {
			continue
		}

		if config.run != nil && !config.run.MatchString(testJourney.FilePath) {
			continue
		}

		if testJourneyGroups.Has("manual") && !config.manualGroups.HasAny(testJourneyGroups) {
			continue
		}

		filteredJourneys = append(filteredJourneys, testJourney)
	}
	return filteredJourneys
}

// `runJourneyInParallel` run `x` amount of `goroutines`
// where `x` is the number of 'parallel active users'.
// Each `goroutine` will execute the journey specified
// as a parameter until a `timeout` is reached.
// If the `timeout` is reached in the middle of
// journey/journeys execution, it will be interrupted
// and no logs/errors will be generated for cancelled
// journeys.
func (jt *journeyTester) runJourneyInParallel(
	ctx *journey_tester_context.Context,
	setupSteps []*Step,
	journey *Journey,
	numParallelActiveUsers int,
	timeout int,
) ([]*load_report.LoadJourneyLog, int, []error) {
	var jls []*load_report.LoadJourneyLog
	steps := append(setupSteps, journey.Steps...)
	var syncErrors sync.Map
	timeoutContext, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()

	jt.loadPrinter.PrintJourneysParallelStart(journey.Desc)

	journeysCount := 0
	for i := 1; i <= numParallelActiveUsers; i++ {
		go func(activeUserNum int) {
			for {
				select {
				case <-timeoutContext.Done():
					return
				default:
				}

				jt.loadPrinter.PrintJourneyStart(activeUserNum)
				journeysCount++
				log, err := runSteps(jt.loadPrinter, ctx.Copy(), steps, journey)
				jls = append(jls, log)
				if err != nil {
					syncErrors.Store(journey.Desc, err)
					return
				}
			}
		}(i)
	}

	<-timeoutContext.Done()
	cancelledJourneys := journeysCount - len(jls)
	jt.loadPrinter.PrintTimeout(cancelledJourneys)

	var errors []error
	syncErrors.Range(func(key, value any) bool {
		errors = append(errors, value.(error))
		return true
	})

	return jls, journeysCount, errors
}
