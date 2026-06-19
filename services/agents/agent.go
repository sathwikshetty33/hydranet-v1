package agents

import (
	"context"
	"path"
	"strings"
)

// IncidentDetails is the in-process shape every agent emits and every
// reporter consumes. It is deliberately transport-agnostic — proto, JSON,
// stdout, in-memory: all converge on this struct.
type IncidentDetails struct {
	Namespace   string
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

// MatchMode controls how a Pattern's Needle is compared against an exec event.
//
//   - MatchSubstring  : Needle must appear anywhere in the binary path or
//     arguments. Use for path-shaped needles like "/etc/shadow"
//     or "docker.sock".
//   - MatchCommand    : Needle must equal the binary's basename OR a single
//     whitespace-delimited token in args. Use for command
//     names like "nc" or "curl" so we don't false-match
//     substrings inside larger words ("node-cli").
type MatchMode int

const (
	MatchSubstring MatchMode = iota
	MatchCommand
)

// Pattern is a single suspicious-exec rule, deliberately source-agnostic:
// the Tetragon agent uses it today, but a Falco / eBPF / log-tail agent
// can consult the same table. Severity and Technique flow into the
// IncidentDetails the agent emits; the dashboard groups/colors by them.
type Pattern struct {
	Needle    string
	Severity  string // "low" | "medium" | "high" | "critical"
	Technique string // short, dashboarded label, e.g. "credential-access"
	Match     MatchMode
}

// Matches reports whether the pattern fires against an exec event's binary
// path and (already-joined) arguments string.
func (p Pattern) Matches(binary, args string) bool {
	switch p.Match {
	case MatchSubstring:
		return strings.Contains(binary, p.Needle) || strings.Contains(args, p.Needle)
	case MatchCommand:
		if path.Base(binary) == p.Needle {
			return true
		}
		// Catches busybox-style execs where binary is "/bin/busybox" and
		// the actual command appears as the first arg token.
		for _, tok := range strings.Fields(args) {
			if tok == p.Needle {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// DefaultPatterns is the built-in pattern table. Severities are calibrated
// such that "critical" implies near-zero-false-positive escape preparation
// (the eventual SIGKILL / quarantine candidates); "high" is genuinely
// suspicious but ambiguous; "medium" is recon-tier that frequently has
// legitimate uses (health probes, init scripts).
var DefaultPatterns = []Pattern{
	// ── credential access
	{Needle: "/etc/shadow", Severity: "critical", Technique: "credential-access", Match: MatchSubstring},
	{Needle: "/etc/passwd", Severity: "high", Technique: "credential-access", Match: MatchSubstring},

	// ── capability + container-escape probing (eventual terminate set)
	{Needle: "/proc/self/status", Severity: "critical", Technique: "capability-probe", Match: MatchSubstring},
	{Needle: "capsh", Severity: "critical", Technique: "capability-probe", Match: MatchCommand},
	{Needle: "nsenter", Severity: "critical", Technique: "container-escape", Match: MatchCommand},
	{Needle: "docker.sock", Severity: "critical", Technique: "container-escape", Match: MatchSubstring},
	{Needle: "/proc/1/root", Severity: "critical", Technique: "container-escape", Match: MatchSubstring},

	// ── command-and-control / reverse-shell tooling
	{Needle: "nc", Severity: "high", Technique: "command-and-control", Match: MatchCommand},
	{Needle: "ncat", Severity: "high", Technique: "command-and-control", Match: MatchCommand},

	// ── ingress tools — frequently legitimate, kept low signal
	{Needle: "wget", Severity: "medium", Technique: "ingress-tool", Match: MatchCommand},
	{Needle: "curl", Severity: "medium", Technique: "ingress-tool", Match: MatchCommand},
}
