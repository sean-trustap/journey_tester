// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package load_report

import (
	"log"
	"time"
)

func NewMinimalPrinter() *MinimalPrinter {
	return &MinimalPrinter{}
}

type MinimalPrinter struct{}

func (p *MinimalPrinter) PrintResultSummary(numCompletedJourneys, numStartedJourneys int, hadTestErr, hadFrameworkErr bool) {
	log.Printf("\n\n\tload test run finished\n\n")

	msg := "all completed journeys have passed :)"
	if hadFrameworkErr {
		msg = "result: framework error"
	}
	if hadTestErr {
		msg = "result: some tests have failed"
	}
	if numCompletedJourneys == 0 {
		msg = "all journeys were cancelled"
	}
	log.Printf("\n\n\t%s\n\n -> number of journeys started: %d", msg, numStartedJourneys)
	log.Printf("\n\n\t\n\n -> number of journeys completed: %d", numCompletedJourneys)
}

func (p *MinimalPrinter) PrintMsg(msg string) {
	log.Printf("%s\n", msg)
}

func (p *MinimalPrinter) PrintJourneyStart(activeUserNumber int) {
	log.Printf("\n\n executing journey for parallel active user (PAU): %d", activeUserNumber)
}

func (p *MinimalPrinter) PrintTimeout(cancelledJourneys int) {
	log.Printf("\n\n\t timeout reached: number of journeys cancelled: %d", cancelledJourneys)
}

func (p *MinimalPrinter) PrintJourneysParallelStart(name string) {
	log.Printf("\n\n\t starting to run journey in parallel: %v\n\n", name)
}

func (p *MinimalPrinter) PrintStepResult(step string, duration time.Duration, result StepResult) {
}

func (p *MinimalPrinter) PrintStepMsg(msg string) {
	log.Printf("\t%s\n", msg)
}
