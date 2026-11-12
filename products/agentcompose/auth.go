package agentcompose

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/chaitin/chaitin-cli/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newAuthCommand(state *commandState) *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Manage the local API Token", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }}
	tokenStdin := false
	login := &cobra.Command{Use: "login", Short: "Validate and save an API Token", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		if state.options.DryRun {
			return unsupportedDryRun("auth login", state.options.JSON)
		}
		token, err := readLoginToken(cmd, tokenStdin)
		if err != nil {
			return usageError(err.Error(), state.options.JSON)
		}
		oldToken := state.options.Token
		state.options.Token = token
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		err = verifyToken(ctx, state)
		if err != nil {
			state.options.Token = oldToken
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		stored, err := loadStoredConfig(state.options.ConfigPath)
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		stored.URL = state.options.URL
		stored.APIToken = token
		stored.APIKey = ""
		if stored.Timeout == "" {
			stored.Timeout = state.options.timeoutText
		}
		if state.options.Project != "" {
			stored.DefaultProject = state.options.Project
		}
		stored.Insecure = state.options.Insecure
		if err := config.SetProduct(state.options.ConfigPath, productName, stored); err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), map[string]any{"url": state.options.URL, "token_valid": true, "config_path": state.options.ConfigPath})
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Token validated and saved to %s\n", state.options.ConfigPath)
		return nil
	}}
	login.Flags().BoolVar(&tokenStdin, "token-stdin", false, "Read Token from stdin")
	status := &cobra.Command{Use: "status", Short: "Validate the configured API Token", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error { return runAuthStatus(cmd, state) }}
	logout := &cobra.Command{Use: "logout", Short: "Clear the locally saved API Token", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		if state.options.DryRun {
			return unsupportedDryRun("auth logout", state.options.JSON)
		}
		stored, err := loadStoredConfig(state.options.ConfigPath)
		if err != nil {
			return err
		}
		stored.APIToken = ""
		stored.APIKey = ""
		if err := config.SetProduct(state.options.ConfigPath, productName, stored); err != nil {
			return err
		}
		warning := envTokenActive()
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), map[string]any{"logged_out": true, "environment_token_active": warning})
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Local Agent Compose Token cleared")
		if warning {
			fmt.Fprintln(cmd.ErrOrStderr(), "Warning: AGENT_COMPOSE_API_TOKEN or AGENT_COMPOSE_API_KEY is still active")
		}
		return nil
	}}
	cmd.AddCommand(login, status, logout)
	return cmd
}

func readLoginToken(cmd *cobra.Command, stdin bool) (string, error) {
	if stdin {
		data, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), 1<<20))
		if err != nil {
			return "", err
		}
		token := strings.TrimSpace(string(data))
		if token == "" {
			return "", fmt.Errorf("Token is empty")
		}
		return token, nil
	}
	file, ok := cmd.InOrStdin().(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return "", fmt.Errorf("terminal input is required; use --token-stdin")
	}
	fmt.Fprint(cmd.ErrOrStderr(), "API Token: ")
	data, err := term.ReadPassword(int(file.Fd()))
	fmt.Fprintln(cmd.ErrOrStderr())
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("Token is empty")
	}
	return token, nil
}

func loadStoredConfig(path string) (productConfig, error) {
	raw, err := config.Load(path)
	if err != nil {
		return productConfig{}, err
	}
	node, ok := raw[productName]
	if !ok {
		return productConfig{}, nil
	}
	var stored productConfig
	if err := node.Decode(&stored); err != nil {
		return stored, fmt.Errorf("decode stored Agent Compose config: %w", err)
	}
	return stored, nil
}

func runAuthStatus(cmd *cobra.Command, state *commandState) error {
	valid := false
	if strings.TrimSpace(state.options.Token) == "" {
		return newError("authentication_failed", "Token is not configured; run agent-compose auth login", exitAuth, state.options.JSON)
	}
	ctx, cancel := requestContext(cmd, state)
	defer cancel()
	if err := verifyToken(ctx, state); err != nil {
		return mapConnectError(err, state.options.URL, state.options.JSON)
	}
	valid = true
	if state.options.JSON {
		return writeJSON(cmd.OutOrStdout(), map[string]any{"url": state.options.URL, "token_source": state.options.TokenSource, "token_configured": true, "token_valid": valid})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "URL: %s\nToken source: %s\nToken configured: yes\nToken valid: yes\n", state.options.URL, state.options.TokenSource)
	return nil
}

// verifyToken calls a protected RPC. HealthService is exempt from daemon
// authentication, so it accepts any Token and cannot validate one.
func verifyToken(ctx context.Context, state *commandState) error {
	_, err := state.clients().project.ListProjects(ctx, connect.NewRequest(&agentcomposev2.ListProjectsRequest{Limit: 1}))
	return err
}
