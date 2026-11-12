package agentcompose

import (
	"context"

	"connectrpc.com/connect"
	agentcomposev2 "github.com/chaitin/agent-compose/proto/agentcompose/v2"
)

func resolveResourceTargets(ctx context.Context, state *commandState, ref string, kinds ...agentcomposev2.ResourceKind) ([]*agentcomposev2.ResourceTarget, bool, error) {
	resp, err := state.clients().resource.ResolveID(ctx, connect.NewRequest(&agentcomposev2.ResolveResourceIDRequest{Id: ref, Kinds: kinds}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeUnimplemented {
			return nil, false, nil
		}
		return nil, true, mapConnectError(err, state.options.URL, state.options.JSON)
	}
	return resp.Msg.GetTargets(), true, nil
}
