// Package mcp provides an MCP server that exposes the ROS OCP REST API as tools.
// All calls are forwarded to the REST API with the user's X-Rh-Identity so that
// data access is scoped per user/org (aligned with the REST API).
package mcp

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redhatinsights/ros-ocp-backend/internal/logging"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry = logging.GetLogger()

const (
	recommendationsPath = "/api/cost-management/v1/recommendations/openshift"
)

// APIClient calls the ROS OCP REST API with a fixed X-Rh-Identity (user context).
// Each MCP session should use a client created with that session's identity.
type APIClient struct {
	baseURL    string
	identity   string
	httpClient *http.Client
}

// NewAPIClient returns a client that sends X-Rh-Identity on every request.
// identity must be the base64-encoded X-Rh-Identity header value (same as REST API).
func NewAPIClient(baseURL, identity string) *APIClient {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &APIClient{
		baseURL:  baseURL,
		identity: identity,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListRecommendations calls GET /api/cost-management/v1/recommendations/openshift
// with optional query parameters. Returns the response body as string or an error.
func (c *APIClient) ListRecommendations(params ListRecommendationsParams) (string, error) {
	u, err := url.Parse(c.baseURL + recommendationsPath)
	if err != nil {
		return "", fmt.Errorf("parse list url: %w", err)
	}
	q := u.Query()
	if params.Format != "" {
		q.Set("format", params.Format)
	}
	if params.Cluster != "" {
		q.Set("cluster", params.Cluster)
	}
	if params.WorkloadType != "" {
		q.Set("workload_type", params.WorkloadType)
	}
	if params.Workload != "" {
		q.Set("workload", params.Workload)
	}
	if params.Container != "" {
		q.Set("container", params.Container)
	}
	if params.Project != "" {
		q.Set("project", params.Project)
	}
	if params.StartDate != "" {
		q.Set("start_date", params.StartDate)
	}
	if params.EndDate != "" {
		q.Set("end_date", params.EndDate)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Offset > 0 {
		q.Set("offset", fmt.Sprintf("%d", params.Offset))
	}
	if params.OrderBy != "" {
		q.Set("order_by", params.OrderBy)
	}
	if params.OrderHow != "" {
		q.Set("order_how", params.OrderHow)
	}
	q.Set("cpu-unit", withDefault(params.CPUUnit, "millicores"))
	q.Set("memory-unit", withDefault(params.MemoryUnit, "MiB"))
	q.Set("true-units", withDefault(params.TrueUnits, "true"))
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	c.setIdentity(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

// GetRecommendation calls GET /api/cost-management/v1/recommendations/openshift/:id
// with optional query parameters. Returns the response body as string or an error.
func (c *APIClient) GetRecommendation(recommendationID string, params GetRecommendationParams) (string, error) {
	u, err := url.Parse(c.baseURL + recommendationsPath + "/" + url.PathEscape(recommendationID))
	if err != nil {
		return "", fmt.Errorf("parse get url: %w", err)
	}
	q := u.Query()
	q.Set("cpu-unit", withDefault(params.CPUUnit, "millicores"))
	q.Set("memory-unit", withDefault(params.MemoryUnit, "MiB"))
	q.Set("true-units", withDefault(params.TrueUnits, "true"))
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	c.setIdentity(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

func (c *APIClient) setIdentity(req *http.Request) {
	req.Header.Set("X-Rh-Identity", c.identity)
	req.Header.Set("Accept", "application/json")
}

func withDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// ListRecommendationsParams mirrors the REST API query parameters for list.
// All fields are optional; the REST API applies defaults when omitted.
type ListRecommendationsParams struct {
	Format       string `json:"format,omitempty" jsonschema_description:"Response format: json or csv (default json)"`
	Cluster      string `json:"cluster,omitempty" jsonschema_description:"Filter by cluster alias or UUID"`
	WorkloadType string `json:"workload_type,omitempty" jsonschema_description:"Filter by workload type: daemonset, deployment, deploymentconfig, replicaset, replicationcontroller, statefulset"`
	Workload     string `json:"workload,omitempty" jsonschema_description:"Filter by workload name"`
	Container    string `json:"container,omitempty" jsonschema_description:"Filter by container name"`
	Project      string `json:"project,omitempty" jsonschema_description:"Filter by project (namespace) name"`
	StartDate    string `json:"start_date,omitempty" jsonschema_description:"Start date YYYY-MM-DD"`
	EndDate      string `json:"end_date,omitempty" jsonschema_description:"End date YYYY-MM-DD"`
	Limit        int    `json:"limit,omitempty" jsonschema_description:"Max results (1-100, default 10)"`
	Offset       int    `json:"offset,omitempty" jsonschema_description:"Pagination offset"`
	OrderBy      string `json:"order_by,omitempty" jsonschema_description:"Order by: cluster, workload_type, workload, project, container, last_reported"`
	OrderHow     string `json:"order_how,omitempty" jsonschema_description:"Order direction: ASC or DESC"`
	CPUUnit      string `json:"cpu_unit,omitempty" jsonschema_description:"CPU unit: millicores or cores"`
	MemoryUnit   string `json:"memory_unit,omitempty" jsonschema_description:"Memory unit: bytes, MiB, GiB"`
	TrueUnits    string `json:"true_units,omitempty" jsonschema_description:"Show real-world units: true or false"`
}

// GetRecommendationParams mirrors the REST API query parameters for get by ID.
type GetRecommendationParams struct {
	CPUUnit    string `json:"cpu_unit,omitempty" jsonschema_description:"CPU unit: millicores or cores"`
	MemoryUnit string `json:"memory_unit,omitempty" jsonschema_description:"Memory unit: bytes, MiB, GiB"`
	TrueUnits  string `json:"true_units,omitempty" jsonschema_description:"Show real-world units: true or false"`
}
