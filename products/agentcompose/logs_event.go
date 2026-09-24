package agentcompose

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/spf13/cobra"
)

const eventLogPrefix = "evt_"

// eventLogScopeCap mirrors the daemon-resolved event scope limit (see
// ListRunsRequest.event_id); the CLI only uses it in the truncation warning.
const eventLogScopeCap = 1000

// validateEventLogTarget rejects values that are not full event-bus event
// ids. Prefix matching is intentionally unsupported for --event.
func validateEventLogTarget(eventID string, jsonOutput bool) error {
	if !strings.HasPrefix(eventID, eventLogPrefix) || strings.TrimSpace(strings.TrimPrefix(eventID, eventLogPrefix)) == "" {
		return usageError(fmt.Sprintf("--event requires a full event id starting with %s; prefix matching is not supported", eventLogPrefix), jsonOutput)
	}
	return nil
}

// executeLogsForEvent lists the agent runs recorded against an event via the
// daemon-side ListRuns event_id filter and replays their logs. The daemon
// resolves the event scope, so the CLI only pages through the result with the
// same offset+total loop as other list views.
func executeLogsForEvent(cmd *cobra.Command, state *commandState, projectID string, options logsOptions) error {
	client := state.clients().run
	runs := make([]*agentcomposev2.RunSummary, 0, 200)
	eventScopeTruncated := false
	offset := uint32(0)
	for {
		resp, err := client.ListRuns(cmd.Context(), connect.NewRequest(&agentcomposev2.ListRunsRequest{
			ProjectId: strings.TrimSpace(projectID),
			AgentName: strings.TrimSpace(options.Agent),
			EventId:   strings.TrimSpace(options.Event),
			Offset:    offset,
			Limit:     200,
		}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		if resp.Msg.GetEventScopeTruncated() {
			eventScopeTruncated = true
		}
		runs = append(runs, resp.Msg.GetRuns()...)
		offset += uint32(len(resp.Msg.GetRuns()))
		if len(resp.Msg.GetRuns()) == 0 || offset >= resp.Msg.GetTotal() {
			break
		}
	}
	if eventScopeTruncated {
		_, err := fmt.Fprintf(cmd.ErrOrStderr(), "Warning: event %s has more associated events than the daemon resolves (cap %d); runs recorded only against events beyond the cap are missing from this output\n", strings.TrimSpace(options.Event), eventLogScopeCap)
		if err != nil {
			return err
		}
	}
	if len(runs) == 0 {
		// JSON mode writes nothing, matching the record-stream contract of
		// followOneRun and the empty output of logs --json with no targets;
		// scripts detect emptiness by the absence of records, not a wrapper.
		if !state.options.JSON {
			_, err := fmt.Fprintf(cmd.ErrOrStderr(), "No runs are associated with event %s\n", strings.TrimSpace(options.Event))
			return err
		}
		return nil
	}
	// ListRuns returns newest first; replay oldest first by start time.
	// Ties keep the daemon's order (SliceStable) — the sibling CLI additionally
	// breaks ties by agent/run id.
	sort.SliceStable(runs, func(i, j int) bool {
		return runStartedBefore(runs[i], runs[j])
	})
	for _, run := range runs {
		if err := followOneRun(cmd, state, projectID, run, options); err != nil {
			return err
		}
	}
	return nil
}

// runStartedBefore orders by start time; runs without a usable start time sort
// last.
func runStartedBefore(left, right *agentcomposev2.RunSummary) bool {
	leftTime, leftOK := runStartedTime(left)
	rightTime, rightOK := runStartedTime(right)
	switch {
	case leftOK && rightOK:
		return leftTime.Before(rightTime)
	case leftOK != rightOK:
		return leftOK
	}
	return false
}

func runStartedTime(run *agentcomposev2.RunSummary) (time.Time, bool) {
	started := run.GetStartedAt()
	if started == nil || !started.IsValid() {
		return time.Time{}, false
	}
	value := started.AsTime()
	if value.IsZero() {
		return time.Time{}, false
	}
	return value, true
}
