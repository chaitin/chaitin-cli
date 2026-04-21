package ddr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/chaitin/workspace-cli/config"
	"github.com/spf13/cobra"
)

type jwtClaims struct {
	UserID string `json:"UserID"`
	UserNS string `json:"UserNS"`
}

type accessKeyBatchRequest struct {
	Enable        bool                       `json:"enable"`
	EnableSubRule bool                       `json:"enable_sub_rule"`
	Permissions   []accessKeyPermissionEntry `json:"permissions"`
}

type accessKeyPermissionEntry struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type accessKeyBatchResponse struct {
	Data struct {
		Data []struct {
			AccessKey struct {
				AccessKey string `json:"access_key"`
				SecretKey string `json:"secret_key"`
			} `json:"access_key"`
		} `json:"data"`
	} `json:"data"`
}

type nsAttributesResponse struct {
	Data struct {
		Attributes []struct {
			Key   string `json:"k"`
			Value string `json:"v"`
		} `json:"attributes"`
	} `json:"data"`
}

func newGetAPITokenCommand() *cobra.Command {
	var jwtToken string
	var rawURL string

	cmd := &cobra.Command{
		Use:   "get-api-token",
		Short: "Create a Serval API token from a JWT token and persist it to config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runtimeConfigPath == "" {
				return fmt.Errorf("config path is empty")
			}

			client := getClient(cmd)
			if rawURL != "" {
				client.config.URL = rawURL
				client.baseURL = strings.TrimSuffix(rawURL, "/")
			}
			if client.config.URL == "" {
				return fmt.Errorf("--url is required")
			}

			result, err := createAndPersistAPIToken(cmd.Context(), client, runtimeConfigPath, jwtToken)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "api_key: %s\ncompany_id: %s\nconfig: %s\n", result.Token, result.CompanyID, runtimeConfigPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&jwtToken, "jwt-token", "", `JWT token copied from the Authorization header, for example "Bearer xxx"`)
	cmd.Flags().StringVar(&rawURL, "url", "", "DDR URL used to create and persist API config")
	_ = cmd.MarkFlagRequired("jwt-token")

	return cmd
}

type apiTokenResult struct {
	Token     string
	CompanyID string
}

func createAndPersistAPIToken(ctx context.Context, client *Client, configPath, jwtToken string) (*apiTokenResult, error) {
	claims, err := parseJWTClaims(jwtToken)
	if err != nil {
		return nil, err
	}

	accessKey, err := createAccessKey(ctx, client, jwtToken, claims)
	if err != nil {
		return nil, err
	}

	token := buildServalToken(accessKey)
	companyID, err := fetchCompanyID(ctx, client, token)
	if err != nil {
		return nil, err
	}

	newCfg := runtimeCfg
	if client.config.URL != "" {
		newCfg.URL = normalizeDDRConfigURL(client.config.URL)
	}
	newCfg.APIKey = token
	newCfg.CompanyID = companyID

	if err := config.SetProduct(configPath, "ddr", newCfg); err != nil {
		return nil, err
	}

	runtimeCfg = newCfg
	return &apiTokenResult{Token: token, CompanyID: companyID}, nil
}

func parseJWTClaims(jwtToken string) (*jwtClaims, error) {
	token := strings.TrimSpace(jwtToken)
	if token == "" {
		return nil, fmt.Errorf("jwt token is required")
	}
	if !strings.HasPrefix(token, "Bearer ") {
		return nil, fmt.Errorf(`jwt token must start with "Bearer "`)
	}

	raw := strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode jwt payload: %w", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parse jwt payload: %w", err)
	}
	if claims.UserID == "" || claims.UserNS == "" {
		return nil, fmt.Errorf("jwt token missing UserID or UserNS")
	}

	return &claims, nil
}

func createAccessKey(ctx context.Context, client *Client, jwtToken string, claims *jwtClaims) (string, error) {
	originalToken := client.config.APIKey
	originalCompanyID := client.config.CompanyID
	client.config.APIKey = jwtToken
	client.config.CompanyID = ""
	defer func() {
		client.config.APIKey = originalToken
		client.config.CompanyID = originalCompanyID
	}()

	reqBody := accessKeyBatchRequest{
		Enable:        true,
		EnableSubRule: false,
		Permissions: []accessKeyPermissionEntry{
			{Resource: "g:allResource", Action: "g:allAction"},
		},
	}

	userHeader, err := json.Marshal(map[string]interface{}{
		"user_id": claims.UserID,
		"extra": map[string]string{
			"ns": claims.UserNS,
		},
	})
	if err != nil {
		return "", fmt.Errorf("marshal user header: %w", err)
	}

	var resp accessKeyBatchResponse
	if err := client.Do(ctx, http.MethodPost, buildFixedAuthURL(client.config.URL, "/qzh/api/auth/v1/access_key/batch"), nil, map[string]string{
		"User": string(userHeader),
	}, reqBody, &resp); err != nil {
		return "", err
	}

	if len(resp.Data.Data) == 0 || resp.Data.Data[0].AccessKey.AccessKey == "" {
		return "", fmt.Errorf("access key not found in response")
	}

	return resp.Data.Data[0].AccessKey.AccessKey, nil
}

func buildServalToken(accessKey string) string {
	return base64.StdEncoding.EncodeToString([]byte("serval:" + accessKey))
}

func fetchCompanyID(ctx context.Context, client *Client, token string) (string, error) {
	originalToken := client.config.APIKey
	originalCompanyID := client.config.CompanyID
	client.config.APIKey = token
	client.config.CompanyID = ""
	defer func() {
		client.config.APIKey = originalToken
		client.config.CompanyID = originalCompanyID
	}()

	var resp nsAttributesResponse
	if err := client.Do(ctx, http.MethodGet, buildFixedAuthURL(client.config.URL, "/qzh/api/auth/v1/ns/attributes"), nil, nil, nil, &resp); err != nil {
		return "", err
	}

	for _, attr := range resp.Data.Attributes {
		if attr.Key == "corp_name" && attr.Value != "" {
			return attr.Value, nil
		}
	}

	return "", fmt.Errorf("corp_name not found in namespace attributes")
}

func normalizeDDRConfigURL(raw string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return ""
	}
	if strings.HasSuffix(trimmed, "/qzh/api/v1") {
		return trimmed
	}
	return trimmed + "/qzh/api/v1"
}

func buildFixedAuthURL(rawBaseURL, authPath string) string {
	base := strings.TrimRight(strings.TrimSpace(rawBaseURL), "/")
	if base == "" {
		return authPath
	}

	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return authPath
	}

	return parsed.Scheme + "://" + parsed.Host + authPath
}
