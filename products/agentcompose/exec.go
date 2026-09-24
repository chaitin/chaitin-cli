package agentcompose

import (
	"fmt"
	"io"
	"strings"
	"time"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/spf13/cobra"
)

type execOptions struct {
	Command, CWD, Timeout string
	Env                   []string
	MaxOutput             uint32
}
type execDTO struct {
	ExecID    string   `json:"exec_id"`
	SandboxID string   `json:"sandbox_id"`
	RunID     string   `json:"run_id,omitempty"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	CWD       string   `json:"cwd,omitempty"`
	ExitCode  int32    `json:"exit_code"`
	Success   bool     `json:"success"`
	Stdout    string   `json:"stdout,omitempty"`
	Stderr    string   `json:"stderr,omitempty"`
	Output    string   `json:"output,omitempty"`
	Error     string   `json:"error,omitempty"`
}

func execFromProto(result *agentcomposev2.ExecResult) execDTO {
	return execDTO{ExecID: result.GetExecId(), SandboxID: result.GetSandboxId(), RunID: result.GetRunId(), Command: result.GetCommand().GetCommand(), Args: append([]string(nil), result.GetCommand().GetArgs()...), CWD: result.GetCwd(), ExitCode: result.GetExitCode(), Success: result.GetSuccess(), Stdout: result.GetStdout(), Stderr: result.GetStderr(), Output: result.GetOutput(), Error: result.GetError()}
}

func newExecCommand(state *commandState) *cobra.Command {
	options := execOptions{}
	cmd := &cobra.Command{Use: "exec <sandbox-ref> -- <command...>", Short: "Execute a non-interactive command in a Sandbox", Args: func(_ *cobra.Command, args []string) error {
		if len(args) < 1 {
			return usageError("a Sandbox reference is required", state.options.JSON)
		}
		return nil
	}, RunE: func(cmd *cobra.Command, args []string) error { return executeExec(cmd, state, args, options) }}
	cmd.Flags().StringVar(&options.Command, "command", "", "Command string (alternative to positional command)")
	cmd.Flags().StringVar(&options.CWD, "cwd", "", "Working directory")
	cmd.Flags().StringSliceVar(&options.Env, "env", nil, "Environment KEY=VALUE (repeatable)")
	cmd.Flags().StringVar(&options.Timeout, "exec-timeout", "", "Remote execution timeout")
	cmd.Flags().Uint32Var(&options.MaxOutput, "max-output-bytes", 0, "Maximum captured output bytes")
	return cmd
}
func executeExec(cmd *cobra.Command, state *commandState, args []string, options execOptions) error {
	if state.options.DryRun {
		return unsupportedDryRun("exec", state.options.JSON)
	}
	commandArgs := args[1:]
	if options.Command != "" && len(commandArgs) > 0 {
		return usageError("--command and positional command are mutually exclusive", state.options.JSON)
	}
	commandName := options.Command
	if commandName == "" {
		if len(commandArgs) == 0 {
			return usageError("a command is required", state.options.JSON)
		}
		commandName = commandArgs[0]
		commandArgs = commandArgs[1:]
	}
	env := make([]*agentcomposev2.EnvVarSpec, 0, len(options.Env))
	for _, item := range options.Env {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			return usageError("--env must use KEY=VALUE", state.options.JSON)
		}
		env = append(env, &agentcomposev2.EnvVarSpec{Name: key, Value: value})
	}
	var timeoutMS uint32
	if options.Timeout != "" {
		duration, err := time.ParseDuration(options.Timeout)
		if err != nil || duration <= 0 || duration.Milliseconds() > int64(^uint32(0)) {
			return usageError("--exec-timeout must be a positive duration within uint32 milliseconds", state.options.JSON)
		}
		timeoutMS = uint32(duration.Milliseconds())
	}
	ctx, cancel := requestContext(cmd, state)
	project, err := resolveProject(ctx, state, state.clients().project)
	if err != nil {
		cancel()
		return err
	}
	sandbox, err := resolveSandbox(ctx, state, state.clients().sandbox, project.GetSummary().GetProjectId(), args[0])
	cancel()
	if err != nil {
		return err
	}
	request := &agentcomposev2.ExecRequest{Target: &agentcomposev2.ExecRequest_SandboxId{SandboxId: sandbox.GetSandboxId()}, Command: &agentcomposev2.ExecCommand{Command: commandName, Args: commandArgs}, Cwd: options.CWD, Env: env, TimeoutMs: timeoutMS, MaxOutputBytes: options.MaxOutput}
	ctx, cancel = streamContext(cmd)
	defer cancel()
	stream, err := state.clients().exec.StreamExec(ctx, connect.NewRequest(request))
	if err != nil {
		return mapConnectError(err, state.options.URL, state.options.JSON)
	}
	var final *agentcomposev2.ExecResult
	for stream.Receive() {
		event := stream.Msg()
		if event.GetResult() != nil {
			final = event.GetResult()
		}
		if state.options.JSON {
			record := map[string]any{"event_type": enumText(event.GetEventType(), "STREAM_EXEC_EVENT_TYPE_"), "exec_id": event.GetExecId(), "sandbox_id": event.GetSandboxId(), "run_id": event.GetRunId(), "chunk": event.GetChunk(), "stream": enumText(event.GetStream(), "STDIO_STREAM_")}
			if final != nil {
				addExecFields(record, execFromProto(final))
			}
			if err := writeJSON(cmd.OutOrStdout(), record); err != nil {
				return err
			}
		} else {
			target := cmd.OutOrStdout()
			if event.GetStream() == agentcomposev2.StdioStream_STDIO_STREAM_STDERR {
				target = cmd.ErrOrStderr()
			}
			if _, err := io.WriteString(target, event.GetChunk()); err != nil {
				return err
			}
		}
	}
	if err := stream.Err(); err != nil {
		return mapConnectError(err, state.options.URL, state.options.JSON)
	}
	if final == nil {
		return newError("protocol_error", "Exec stream ended without a final result", exitGeneral, state.options.JSON)
	}
	code := int(final.GetExitCode())
	if code < 0 || code > 255 {
		remote := code
		result := newError("remote_exit_invalid", fmt.Sprintf("remote process returned invalid exit status %d", code), exitGeneral, state.options.JSON)
		result.RemoteExitCode = &remote
		return result
	}
	if code != 0 {
		return exitStatus{code: code}
	}
	return nil
}

func addExecFields(record map[string]any, result execDTO) {
	record["exec_id"] = result.ExecID
	record["sandbox_id"] = result.SandboxID
	record["command"] = result.Command
	record["args"] = result.Args
	record["exit_code"] = result.ExitCode
	record["success"] = result.Success
	if result.RunID != "" {
		record["run_id"] = result.RunID
	}
	if result.CWD != "" {
		record["cwd"] = result.CWD
	}
	if result.Stdout != "" {
		record["stdout"] = result.Stdout
	}
	if result.Stderr != "" {
		record["stderr"] = result.Stderr
	}
	if result.Output != "" {
		record["output"] = result.Output
	}
	if result.Error != "" {
		record["error"] = result.Error
	}
}
