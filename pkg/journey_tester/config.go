// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package journey_tester

import (
	"flag"
	"fmt"
	"regexp"
	"strings"

	"4d63.com/optional"
	"github.com/trustap/journey_tester/pkg/hashset"
	"github.com/trustap/journey_tester/pkg/journey_tester/load_report"
	"github.com/trustap/journey_tester/pkg/journey_tester/report"
)

type JourneyTesterConfig struct {
	createData    bool
	deleteData    bool
	preDeleteData bool
	skipTesting   bool
	skipVerify    bool
	loadTesting   bool
	groups        hashset.Set[string]
	skipGroups    hashset.Set[string]
	manualGroups  hashset.Set[string]
	run           *regexp.Regexp
	numThreads    optional.Optional[int]
	printer       report.Printer
	loadPrinter   load_report.Printer
	timeout       optional.Optional[int]
	configFile    string
}

func parseFlags() (*JourneyTesterConfig, error) {
	createData := flag.Bool("create", false, "if true, creates test data")
	deleteData := flag.Bool("delete", false, "if true, deletes test data after running tests")
	preDeleteData := flag.Bool("pre-delete", false, "if true, deletes test data before running tests; failures don't prevent tests")
	skipTesting := flag.Bool("skip-testing", false, "if true, skips running the tests")
	loadTesting := flag.Bool("load-testing", false, "if true, runs journey as a load test")
	numThreads := flag.Int("num-threads", 0, "the number of threads to run in parallel")
	timeout := flag.Int("timeout", 0, "the time in second to cancel the load test execution")
	groups := flag.String("groups", "", "list of comma separated groups to run")
	skipGroups := flag.String("skip-groups", "", "list of comma separated groups to not run")
	manualGroups := flag.String("manual-groups", "", "allow manual tests to run if they contain any of these groups")
	run := flag.String("run", "", "only run tests whose path matches this regular expression")
	skipVerify := flag.Bool("skip-verify", false, "skip verifying the data required")
	output := flag.String("output", "full", "can be set to `minimal` for minimal output")
	configFile := flag.String("config", "configs/journey_tester.yaml", "the journey tester config")

	flag.Parse()

	fc := &JourneyTesterConfig{}
	if createData != nil {
		fc.createData = *createData
	}
	if deleteData != nil {
		fc.deleteData = *deleteData
	}
	if preDeleteData != nil {
		fc.preDeleteData = *preDeleteData
	}
	if skipTesting != nil {
		fc.skipTesting = *skipTesting
	}
	if skipVerify != nil {
		fc.skipVerify = *skipVerify
	}
	if loadTesting != nil {
		fc.loadTesting = *loadTesting
	}
	if numThreads != nil {
		fc.numThreads = optional.Empty[int]()
		if *numThreads > 0 {
			fc.numThreads = optional.OfPtr[int](numThreads)
		}
	}
	if timeout != nil {
		fc.timeout = optional.Empty[int]()
		if *timeout > 0 {
			fc.timeout = optional.OfPtr[int](timeout)
		}
	}
	if groups != nil && len(*groups) > 0 {
		fc.groups = hashset.SetFromSlice(strings.Split(*groups, ","))
	}
	if manualGroups != nil && len(*manualGroups) > 0 {
		fc.manualGroups = hashset.SetFromSlice(strings.Split(*manualGroups, ","))
	}
	if skipGroups != nil && len(*skipGroups) > 0 {
		fc.skipGroups = hashset.SetFromSlice(strings.Split(*skipGroups, ","))
	}
	fc.printer = report.NewLogPrinter()
	fc.loadPrinter = load_report.NewMinimalPrinter()
	if *output == "minimal" {
		fc.printer = report.NewMinimalPrinter()
	}
	fc.configFile = *configFile

	if run != nil {
		re, err := regexp.Compile(*run)
		if err != nil {
			return nil, fmt.Errorf("couldn't compile regex for `-run`: %w", err)
		}
		fc.run = re
	}
	return fc, nil
}
