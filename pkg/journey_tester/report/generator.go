// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package report

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type TestRunInfo struct {
	StartTime           string
	EndTime             string
	Duration            string
	HasFailures         bool
	NumberOfFailedTests int
	NumberOfTests       int
	FailedJourneys      []*JourneyLog
	SucceededJourneys   []*JourneyLog
	FrameworkError      string
}

func newTestRunInfo(r *Reporter) *TestRunInfo {
	var numJourneys int
	var numFailedJourneys int
	var failed []*JourneyLog
	var succeeded []*JourneyLog
	for _, v := range r.Journeys {
		numJourneys++
		if v.HasError {
			numFailedJourneys++
			failed = append(failed, v)
		} else {
			succeeded = append(succeeded, v)
		}
	}
	var frameErr string
	if r.FrameworkError != nil {
		frameErr = (*r.FrameworkError).Error()
	}

	return &TestRunInfo{
		HasFailures:         numFailedJourneys > 0 || r.FrameworkError != nil,
		NumberOfFailedTests: numFailedJourneys,
		NumberOfTests:       numJourneys,
		FailedJourneys:      failed,
		SucceededJourneys:   succeeded,
		StartTime:           r.StartTime.Format(time.RFC3339),
		EndTime:             r.EndTime.Format(time.RFC3339),
		Duration:            r.EndTime.Sub(r.StartTime).String(),
		FrameworkError:      frameErr,
	}
}

func GenerateReports(printer Printer, r *Reporter, reportOutputDir string) error {
	info := newTestRunInfo(r)

	err := GenerateFullReport(printer, info, reportOutputDir)
	if err != nil {
		return fmt.Errorf("failed to generate full report: %w", err)
	}

	err = GenerateSummaryReport(printer, info, reportOutputDir)
	if err != nil {
		return fmt.Errorf("failed to generate summary report: %w", err)
	}

	journeys := append(info.FailedJourneys, info.SucceededJourneys...)
	successID, failID := 0, 0
	for _, j := range journeys {
		if j.HasError {
			err := GenerateJourneyReport(printer, j, failID, "f", reportOutputDir)
			failID++
			if err != nil {
				return fmt.Errorf("failed to generate journey report for failure %d: %w", failID, err)
			}
		} else {
			err := GenerateJourneyReport(printer, j, successID, "s", reportOutputDir)
			successID++
			if err != nil {
				return fmt.Errorf("failed to generate journey report for success %d: %w", successID, err)
			}
		}
	}

	err = copyFromEmbed("assets/styles.css", reportOutputDir+"/styles.css")
	if err != nil {
		return fmt.Errorf("failed to copy styles.css: %w", err)
	}

	err = copyFromEmbed("assets/summary_styles.css", reportOutputDir+"/summary_styles.css")
	if err != nil {
		return fmt.Errorf("failed to copy summary_styles.css: %w", err)
	}

	return nil
}

func GenerateFullReport(printer Printer, i *TestRunInfo, reportOutputDir string) error {
	f, err := os.Create(reportOutputDir + "/report.html")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	tmpl := template.Must(template.ParseFS(reportAssets, "assets/template.html"))
	err = tmpl.Execute(f, *i)
	if err != nil {
		return fmt.Errorf("failed to generate html template: %w", err)
	}

	abs, err := filepath.Abs(f.Name())
	if err != nil {
		return fmt.Errorf("failed to get absolute file path: %w", err)
	}
	printer.PrintReportPath(abs)

	return nil
}

func copyFromEmbed(embeddedPath, dst string) error {
	data, err := reportAssets.ReadFile(embeddedPath)
	if err != nil {
		return fmt.Errorf("couldn't read embedded file '%s': %w", embeddedPath, err)
	}
	err = os.WriteFile(dst, data, FileModeFileOwnerOnly)
	if err != nil {
		return fmt.Errorf("couldn't write '%s': %w", dst, err)
	}
	return nil
}

func GenerateSummaryReport(printer Printer, i *TestRunInfo, reportOutputDir string) error {
	f, err := os.Create(reportOutputDir + "/summary.html")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	tmpl := template.Must(template.ParseFS(reportAssets, "assets/summary_report_template.html"))
	err = tmpl.Execute(f, *i)
	if err != nil {
		return fmt.Errorf("failed to generate html template: %w", err)
	}
	return nil
}

func GenerateJourneyReport(printer Printer, j *JourneyLog, index int, status, reportOutputDir string) error {
	f, err := os.Create(reportOutputDir + "/report-" + status + "-" + strconv.Itoa(index) + ".html")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	tmpl := template.Must(template.ParseFS(reportAssets, "assets/journey_template.html"))
	err = tmpl.Execute(f, *j)
	if err != nil {
		return fmt.Errorf("failed to generate html template: %w", err)
	}
	return nil
}
