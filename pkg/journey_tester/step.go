// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package journey_tester

import (
	"time"

	"4d63.com/optional"
	journey_tester_context "github.com/trustap/journey_tester/pkg/journey_tester/context"
	"github.com/trustap/journey_tester/pkg/journey_tester/runner"
)

func NewStep(desc string, conf StepFunc) *Step {
	return &Step{
		Type: StepTypeRequest,
		Desc: desc,
		Conf: conf,
	}
}

type Step struct {
	Type            StepType
	Desc            string
	Conf            func(c *journey_tester_context.Context) runner.Step
	IsCleanUpStep   bool
	PollTimeoutConf optional.Optional[string]
	WaitTime        *time.Duration
}

type StepType string

const (
	StepTypeWait    StepType = "wait_step"
	StepTypeRequest StepType = "request_step"
)

type StepFunc func(c *journey_tester_context.Context) runner.Step
