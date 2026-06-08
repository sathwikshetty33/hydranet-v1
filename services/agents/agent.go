package agents

import "context"

type IncidentDetails struct {
	PodName     string
	ServiceName string
	Message     string
	Severity    string
}

type Agent interface {
    ConsumeEvents(ctx context.Context) error
    ReportIncident(ctx context.Context, incident IncidentDetails) error
}


