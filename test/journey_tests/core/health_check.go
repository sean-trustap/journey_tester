// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package tests

import (
	"github.com/trustap/journey_tester/pkg/journey_tester"
	journey_tester_context "github.com/trustap/journey_tester/pkg/journey_tester/context"
	"github.com/trustap/journey_tester/pkg/journey_tester/runner"
)

func HealthCheck() *journey_tester.Journey {
	j := journey_tester.New("Health check")

	j.AddStep("Server responds to heartbeat", CheckHealth())

	return j
}

func CheckHealth() journey_tester.StepFunc {
	return func(c *journey_tester_context.Context) runner.Step {
		return runner.StrictStepConfig{
			Req: runner.Req{
				URL:    c.Conf.Get("api.host") + "/heartbeat",
				Method: "GET",
			},
			Resp: runner.StrictResp{
				Status: 204,
			},
		}
	}
}
