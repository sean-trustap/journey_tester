// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package load_report

import (
	"encoding/json"
	"fmt"
	"os"

	journey_tester_report "github.com/trustap/journey_tester/pkg/journey_tester/report"
)

func WriteReportDataToFile(lr *LoadReporter, reportDataFile string) error {
	file, err := json.MarshalIndent(lr, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal data with indentation: %w", err)
	}
	err = os.WriteFile(reportDataFile, file, journey_tester_report.FileModeFileOwnerOnly)
	if err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}
	return nil
}
