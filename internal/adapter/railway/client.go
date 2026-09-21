package railway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultGraphQLURL = "https://backboard.railway.com/graphql/v2"

// Client is a live Railway GraphQL HTTP client.
type Client struct {
	Token        string
	BaseURL      string
	HTTP         *http.Client
	ProjectToken bool // use Project-Access-Token header instead of Bearer
}

// NewClient builds a client from token and optional base URL.
func NewClient(token, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultGraphQLURL
	}
	c := &Client{
		Token:   token,
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}
	// Project tokens are typically shorter opaque strings used with Project-Access-Token.
	// Callers can force ProjectToken via RAILWAY_PROJECT_TOKEN env (see NewFromEnv).
	return c
}

type gqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message    string `json:"message"`
		Extensions struct {
			Code    string `json:"code"`
			TraceID string `json:"traceId"`
		} `json:"extensions"`
	} `json:"errors"`
}

// GraphQL executes a query/mutation against Railway's public API.
func (c *Client) GraphQL(ctx context.Context, query string, variables map[string]interface{}) ([]byte, error) {
	if c == nil || c.Token == "" {
		return nil, fmt.Errorf("railway client: missing token")
	}
	body, err := json.Marshal(gqlRequest{Query: query, Variables: variables})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.ProjectToken {
		req.Header.Set("Project-Access-Token", c.Token)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("railway client: rate limited (429)")
	}
	if res.StatusCode >= 400 && res.StatusCode != 200 {
		return nil, fmt.Errorf("railway client: HTTP %d: %s", res.StatusCode, truncate(string(raw), 200))
	}
	var parsed gqlResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("railway client: decode: %w", err)
	}
	if len(parsed.Errors) > 0 {
		msgs := make([]string, 0, len(parsed.Errors))
		for _, e := range parsed.Errors {
			msgs = append(msgs, e.Message)
		}
		return parsed.Data, fmt.Errorf("railway graphql: %s", strings.Join(msgs, "; "))
	}
	return parsed.Data, nil
}

// ValidateToken performs a lightweight auth check (me or projectToken).
func (c *Client) ValidateToken(ctx context.Context) error {
	if c == nil || c.Token == "" {
		return fmt.Errorf("railway client: missing token")
	}
	if c.ProjectToken {
		_, err := c.GraphQL(ctx, `query { projectToken { projectId environmentId } }`, nil)
		return err
	}
	_, err := c.GraphQL(ctx, `query { me { id name email } }`, nil)
	if err != nil {
		// Fallback: project token style
		_, err2 := c.GraphQL(ctx, `query { projectToken { projectId environmentId } }`, nil)
		if err2 == nil {
			c.ProjectToken = true
			return nil
		}
		return err
	}
	return nil
}

// GetService fetches service metadata.
func (c *Client) GetService(ctx context.Context, serviceID string) (id, name string, err error) {
	data, err := c.GraphQL(ctx, `
		query service($id: String!) {
			service(id: $id) { id name projectId }
		}`, map[string]interface{}{"id": serviceID})
	if err != nil {
		return "", "", err
	}
	var out struct {
		Service struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"service"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", "", err
	}
	if out.Service.ID == "" {
		return "", "", fmt.Errorf("railway: service %q not found", serviceID)
	}
	return out.Service.ID, out.Service.Name, nil
}

// GetServiceInstanceHealth returns latest deployment status for health checks.
func (c *Client) GetServiceInstanceHealth(ctx context.Context, serviceID, envID string) (status string, healthy bool, err error) {
	data, err := c.GraphQL(ctx, `
		query serviceInstance($serviceId: String!, $environmentId: String!) {
			serviceInstance(serviceId: $serviceId, environmentId: $environmentId) {
				latestDeployment { id status }
				healthcheckPath
			}
		}`, map[string]interface{}{"serviceId": serviceID, "environmentId": envID})
	if err != nil {
		return "", false, err
	}
	var out struct {
		ServiceInstance struct {
			LatestDeployment struct {
				Status string `json:"status"`
			} `json:"latestDeployment"`
		} `json:"serviceInstance"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", false, err
	}
	status = out.ServiceInstance.LatestDeployment.Status
	switch strings.ToUpper(status) {
	case "SUCCESS", "ACTIVE":
		healthy = true
	}
	return status, healthy, nil
}

// CreateService creates a service (optionally from a Docker image).
func (c *Client) CreateService(ctx context.Context, projectID, name, image string) (string, error) {
	input := map[string]interface{}{
		"projectId": projectID,
		"name":      name,
	}
	if image != "" {
		input["source"] = map[string]interface{}{"image": image}
	}
	data, err := c.GraphQL(ctx, `
		mutation serviceCreate($input: ServiceCreateInput!) {
			serviceCreate(input: $input) { id name }
		}`, map[string]interface{}{"input": input})
	if err != nil {
		return "", err
	}
	var out struct {
		ServiceCreate struct {
			ID string `json:"id"`
		} `json:"serviceCreate"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	if out.ServiceCreate.ID == "" {
		return "", fmt.Errorf("railway: serviceCreate returned empty id")
	}
	return out.ServiceCreate.ID, nil
}

// DeployService triggers serviceInstanceDeployV2.
func (c *Client) DeployService(ctx context.Context, serviceID, envID string) error {
	_, err := c.GraphQL(ctx, `
		mutation serviceInstanceDeployV2($serviceId: String!, $environmentId: String!) {
			serviceInstanceDeployV2(serviceId: $serviceId, environmentId: $environmentId)
		}`, map[string]interface{}{"serviceId": serviceID, "environmentId": envID})
	return err
}

// UpsertVariables sets service variables (used for write-barrier signaling).
func (c *Client) UpsertVariables(ctx context.Context, projectID, envID, serviceID string, vars map[string]string) error {
	_, err := c.GraphQL(ctx, `
		mutation variableCollectionUpsert($input: VariableCollectionUpsertInput!) {
			variableCollectionUpsert(input: $input)
		}`, map[string]interface{}{
		"input": map[string]interface{}{
			"projectId":     projectID,
			"environmentId": envID,
			"serviceId":     serviceID,
			"variables":     vars,
			"skipDeploys":   true,
		},
	})
	return err
}

// CreateVolume attaches a volume to a service.
func (c *Client) CreateVolume(ctx context.Context, projectID, serviceID, mountPath, envID string) (string, error) {
	input := map[string]interface{}{
		"projectId": projectID,
		"serviceId": serviceID,
		"mountPath": mountPath,
	}
	if envID != "" {
		input["environmentId"] = envID
	}
	data, err := c.GraphQL(ctx, `
		mutation volumeCreate($input: VolumeCreateInput!) {
			volumeCreate(input: $input) { id }
		}`, map[string]interface{}{"input": input})
	if err != nil {
		return "", err
	}
	var out struct {
		VolumeCreate struct {
			ID string `json:"id"`
		} `json:"volumeCreate"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	return out.VolumeCreate.ID, nil
}

// CreateServiceDomain attaches a *.railway.app domain to a service (traffic target).
func (c *Client) CreateServiceDomain(ctx context.Context, serviceID, envID string) (string, error) {
	data, err := c.GraphQL(ctx, `
		mutation serviceDomainCreate($input: ServiceDomainCreateInput!) {
			serviceDomainCreate(input: $input) { domain }
		}`, map[string]interface{}{
		"input": map[string]interface{}{
			"serviceId":     serviceID,
			"environmentId": envID,
		},
	})
	if err != nil {
		return "", err
	}
	var out struct {
		ServiceDomainCreate struct {
			Domain string `json:"domain"`
		} `json:"serviceDomainCreate"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	return out.ServiceDomainCreate.Domain, nil
}

// DeleteService removes a service.
func (c *Client) DeleteService(ctx context.Context, serviceID string) error {
	_, err := c.GraphQL(ctx, `
		mutation serviceDelete($id: String!) {
			serviceDelete(id: $id)
		}`, map[string]interface{}{"id": serviceID})
	return err
}

// RollbackDeployment rolls back to a prior deployment id.
func (c *Client) RollbackDeployment(ctx context.Context, deploymentID string) error {
	_, err := c.GraphQL(ctx, `
		mutation deploymentRollback($id: String!) {
			deploymentRollback(id: $id) { id }
		}`, map[string]interface{}{"id": deploymentID})
	return err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
