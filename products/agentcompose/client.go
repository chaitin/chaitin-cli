package agentcompose

import (
	"github.com/chaitin/agent-compose/proto/agentcompose/v2/agentcomposev2connect"
	"github.com/chaitin/agent-compose/proto/health/v1/healthv1connect"
)

type clients struct {
	project  agentcomposev2connect.ProjectServiceClient
	run      agentcomposev2connect.RunServiceClient
	exec     agentcomposev2connect.ExecServiceClient
	sandbox  agentcomposev2connect.SandboxServiceClient
	resource agentcomposev2connect.ResourceServiceClient
	health   healthv1connect.HealthServiceClient
}

func (s *commandState) clients() clients {
	httpClient := s.httpClient()
	baseURL := s.options.URL
	return clients{
		project:  agentcomposev2connect.NewProjectServiceClient(httpClient, baseURL),
		run:      agentcomposev2connect.NewRunServiceClient(httpClient, baseURL),
		exec:     agentcomposev2connect.NewExecServiceClient(httpClient, baseURL),
		sandbox:  agentcomposev2connect.NewSandboxServiceClient(httpClient, baseURL),
		resource: agentcomposev2connect.NewResourceServiceClient(httpClient, baseURL),
		health:   healthv1connect.NewHealthServiceClient(httpClient, baseURL),
	}
}
