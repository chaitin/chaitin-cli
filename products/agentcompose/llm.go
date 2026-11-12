package agentcompose

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
	"github.com/chaitin/agent-compose/proto/agentcompose/v2/agentcomposev2connect"
	"github.com/spf13/cobra"
)

type llmProbeDTO struct {
	Protocol string `json:"protocol"`
	Outcome  string `json:"outcome"`
	Detail   string `json:"detail,omitempty"`
}

// llmProviderDTO deliberately omits per-model headers: they are free-form and
// may carry credentials.
type llmProviderDTO struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	BaseURL   string        `json:"base_url"`
	Protocol  string        `json:"protocol"`
	Auth      string        `json:"auth"`
	Enabled   bool          `json:"enabled"`
	APIKeySet bool          `json:"api_key_set"`
	Models    []string      `json:"models"`
	Probes    []llmProbeDTO `json:"probes,omitempty"`
	ProbedAt  string        `json:"probed_at,omitempty"`
	UpdatedAt string        `json:"updated_at,omitempty"`
}

type llmModelRefDTO struct {
	ProviderID string `json:"provider_id"`
	ModelID    string `json:"model_id"`
}

func llmProviderFromProto(provider *agentcomposev2.LLMProvider) llmProviderDTO {
	models := make([]string, 0, len(provider.GetModels()))
	for _, model := range provider.GetModels() {
		models = append(models, model.GetId())
	}
	probes := make([]llmProbeDTO, 0, len(provider.GetCapabilities().GetProbes()))
	for _, probe := range provider.GetCapabilities().GetProbes() {
		probes = append(probes, llmProbeDTO{Protocol: probe.GetProtocol(), Outcome: probe.GetOutcome(), Detail: probe.GetDetail()})
	}
	return llmProviderDTO{ID: provider.GetId(), Name: provider.GetName(), BaseURL: provider.GetBaseUrl(), Protocol: provider.GetProtocol(), Auth: enumText(provider.GetAuth(), "LLM_PROVIDER_AUTH_"), Enabled: provider.GetEnabled(), APIKeySet: provider.GetApiKeySet(), Models: models, Probes: probes, ProbedAt: timestampText(provider.GetCapabilities().GetProbedAt()), UpdatedAt: timestampText(provider.GetUpdatedAt())}
}

func newLLMCommand(state *commandState) *cobra.Command {
	cmd := &cobra.Command{Use: "llm", Short: "Query LLM provider configuration", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }}
	list := &cobra.Command{Use: "ls", Short: "List LLM providers, their models, and the default model", Args: noArgs(state), RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := requestContext(cmd, state)
		defer cancel()
		client := state.clients().llm
		providers, err := listLLMProviders(ctx, client)
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		resp, err := client.GetDefaultModel(ctx, connect.NewRequest(&agentcomposev2.GetDefaultModelRequest{}))
		if err != nil {
			return mapConnectError(err, state.options.URL, state.options.JSON)
		}
		var defaultModel *llmModelRefDTO
		if model := resp.Msg.GetModel(); model != nil {
			defaultModel = &llmModelRefDTO{ProviderID: model.GetProviderId(), ModelID: model.GetModelId()}
		}
		output := make([]llmProviderDTO, 0, len(providers))
		for _, provider := range providers {
			output = append(output, llmProviderFromProto(provider))
		}
		if state.options.JSON {
			return writeJSON(cmd.OutOrStdout(), struct {
				Providers    []llmProviderDTO `json:"providers"`
				DefaultModel *llmModelRefDTO  `json:"default_model,omitempty"`
			}{output, defaultModel})
		}
		table := newTable(cmd.OutOrStdout(), "NAME\tPROTOCOL\tENABLED\tAPI KEY\tMODELS\tPROBES")
		for _, provider := range output {
			models := make([]string, 0, len(provider.Models))
			for _, model := range provider.Models {
				if defaultModel != nil && defaultModel.ProviderID == provider.ID && defaultModel.ModelID == model {
					model += " (default)"
				}
				models = append(models, model)
			}
			probes := make([]string, 0, len(provider.Probes))
			for _, probe := range provider.Probes {
				probes = append(probes, probe.Protocol+"="+probe.Outcome)
			}
			fmt.Fprintf(table, "%s\t%s\t%t\t%t\t%s\t%s\n", provider.Name, provider.Protocol, provider.Enabled, provider.APIKeySet, strings.Join(models, ","), strings.Join(probes, ","))
		}
		if err := table.Flush(); err != nil {
			return err
		}
		if defaultModel == nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "No default model is configured")
		}
		return nil
	}}
	cmd.AddCommand(list)
	return cmd
}

func listLLMProviders(ctx context.Context, client agentcomposev2connect.LLMServiceClient) ([]*agentcomposev2.LLMProvider, error) {
	providers := make([]*agentcomposev2.LLMProvider, 0)
	for {
		resp, err := client.ListProviders(ctx, connect.NewRequest(&agentcomposev2.ListProvidersRequest{Offset: uint32(len(providers)), Limit: 100}))
		if err != nil {
			return nil, err
		}
		page := resp.Msg.GetProviders()
		providers = append(providers, page...)
		if len(page) == 0 || uint32(len(providers)) >= resp.Msg.GetTotal() {
			return providers, nil
		}
	}
}
