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

func GenerateReports(printer Printer, r *Reporter, reportAssetsDir, reportOutputDir string) error {
	info := newTestRunInfo(r)

	err := GenerateFullReport(printer, info, reportAssetsDir, reportOutputDir)
	if err != nil {
		return fmt.Errorf("failed to generate full report: %w", err)
	}

	err = GenerateSummaryReport(printer, info, reportAssetsDir, reportOutputDir)
	if err != nil {
		return fmt.Errorf("failed to generate summary report: %w", err)
	}

	journeys := append(info.FailedJourneys, info.SucceededJourneys...)
	successID, failID := 0, 0
	for _, j := range journeys {
		if j.HasError {
			err := GenerateJourneyReport(printer, j, failID, "f", reportAssetsDir, reportOutputDir)
			failID++
			if err != nil {
				return fmt.Errorf("failed to generate journey report for failure %d: %w", failID, err)
			}
		} else {
			err := GenerateJourneyReport(printer, j, successID, "s", reportAssetsDir, reportOutputDir)
			successID++
			if err != nil {
				return fmt.Errorf("failed to generate journey report for success %d: %w", successID, err)
			}
		}
	}

	err = copyFile(reportAssetsDir+"/styles.css", reportOutputDir+"/styles.css")
	if err != nil {
		return fmt.Errorf("failed to copy css file: %w", err)
	}

	err = copyFile(reportAssetsDir+"/summary_styles.css", reportOutputDir+"/summary_styles.css")
	if err != nil {
		return fmt.Errorf("failed to copy css file: %w", err)
	}

	return nil
}

func GenerateFullReport(printer Printer, i *TestRunInfo, reportAssetsDir, reportOutputDir string) error {
	f, err := os.Create(reportOutputDir + "/report.html")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	tmpl := template.Must(template.ParseFiles(reportAssetsDir + "/template.html"))
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

func copyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read file: %v err: %w", src, err)
	}
	err = os.WriteFile(dst, data, FileModeFileOwnerOnly)
	if err != nil {
		return fmt.Errorf("failed to write file: %v err: %w", dst, err)
	}
	return nil
}

func GenerateSummaryReport(printer Printer, i *TestRunInfo, reportAssetsDir, reportOutputDir string) error {
	f, err := os.Create(reportOutputDir + "/summary.html")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	tmpl := template.Must(template.ParseFiles(reportAssetsDir + "/summary_report_template.html"))
	err = tmpl.Execute(f, *i)
	if err != nil {
		return fmt.Errorf("failed to generate html template: %w", err)
	}
	return nil
}

func GenerateJourneyReport(printer Printer, j *JourneyLog, index int, status, reportAssetsDir, reportOutputDir string) error {
	f, err := os.Create(reportOutputDir + "/report-" + status + "-" + strconv.Itoa(index) + ".html")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	tmpl := template.Must(template.ParseFiles(reportAssetsDir + "/journey_template.html"))
	err = tmpl.Execute(f, *j)
	if err != nil {
		return fmt.Errorf("failed to generate html template: %w", err)
	}
	return nil
}
