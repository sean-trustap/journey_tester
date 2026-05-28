// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package report

import (
	"encoding/json"
	"fmt"
	"os"
)

func WriteReportDataToFile(r *Reporter, reportDataFile string) error {
	file, err := json.MarshalIndent(r, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal data with indentation: %w", err)
	}
	err = os.WriteFile(reportDataFile, file, FileModeFileOwnerOnly)
	if err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}
	return nil
}

const FileModeFileOwnerOnly = 0o655

func WriteReportToStdOut(printer Printer, r *Reporter) {
	printer.PrintResultSummary(r.HasError, r.FrameworkError != nil)

	failureNumber := 1
	if r.HasError || r.FrameworkError != nil {
		for _, journey := range r.Journeys {
			if journey.HasError {
				for _, step := range journey.Steps {
					if step.HasError {
						printer.PrintFailure(failureNumber, journey.Desc, step)
						failureNumber++
					}
				}
			}
		}
	}
}
