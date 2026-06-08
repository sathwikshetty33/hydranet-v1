package agents

import "context"

// IncidentDetails is the in-process shape every agent emits and every
// reporter consumes. It is deliberately transport-agnostic — proto, JSON,
// stdout, in-memory: all converge on this struct.
type IncidentDetails struct {
	PodName     string
	ServiceName string
	Message     string
	Severity    string
}

// Agent is the source-side contract: subscribe to whatever produces events
// (Tetragon, Falco, eBPF, log tail) and call a Reporter for each one that
// matters. Reporting itself is intentionally NOT on this interface — that
// concern lives on Reporter and is shared across all agent kinds.
type Agent interface {
	ConsumeEvents(ctx context.Context) error
}
