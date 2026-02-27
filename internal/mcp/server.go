package mcp

import (
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/redhatinsights/ros-ocp-backend/internal/config"
	"github.com/redhatinsights/ros-ocp-backend/internal/logging"
	"github.com/sirupsen/logrus"
)

var mcpLog *logrus.Entry = logging.GetLogger()

const (
	implementationName    = "ros-ocp-mcp"
	implementationVersion = "1.0.0"
)

// NewStreamableHTTPHandler returns an http.Handler that serves the MCP Streamable HTTP transport.
// For each new session, getServer is called with the incoming request. The handler reads
// X-Rh-Identity from the request and creates an MCP server whose tools call the ROS REST API
// with that identity, so user data is scoped (same as the REST API). If X-Rh-Identity is
// missing, getServer returns nil and the client receives 400 Bad Request.
func NewStreamableHTTPHandler(cfg *config.Config) http.Handler {
	apiBaseURL := strings.TrimSuffix(cfg.ROSAPIBaseURL, "/")
	if apiBaseURL == "" {
		apiBaseURL = "http://localhost:8000"
	}
	return mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		identity := req.Header.Get("X-Rh-Identity")
		if identity == "" {
			// Optional: support Authorization Bearer with the same value (e.g. agent sends token)
			auth := req.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				identity = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		if identity == "" {
			mcpLog.Warn("MCP request missing X-Rh-Identity and Authorization; rejecting session")
			return nil
		}
		return newServerWithIdentity(identity, apiBaseURL)
	}, nil)
}

// newServerWithIdentity creates an MCP server that exposes list_recommendations and
// get_recommendation tools. All API calls use the given identity (X-Rh-Identity).
func newServerWithIdentity(identity, apiBaseURL string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    implementationName,
		Version: implementationVersion,
	}, nil)
	client := NewAPIClient(apiBaseURL, identity)

	mcp.AddTool(server, &mcp.Tool{
		Name: "list_recommendations",
		Description: "List resource optimization recommendations for OpenShift. " +
			"Returns a summary for each container: current config, recommended config (cost and performance engines), " +
			"variation percentages, and notification codes. " +
			"Supports filters: cluster, workload_type, workload, container, project, start_date, end_date. " +
			"Use limit/offset for pagination. " +
			"Tip: pass start_date=1970-01-01 to get all recommendations regardless of age.",
	}, listRecommendationsHandler(client))

	mcp.AddTool(server, &mcp.Tool{
		Name: "get_recommendation",
		Description: "Get detailed recommendation for a container by UUID. " +
			"Returns everything from list_recommendations PLUS box plot data (CPU and memory usage distributions " +
			"with min, q1, median, q3, max over time) for each recommendation term (short/medium/long). " +
			"Use this to get full details including usage patterns and box plots for a specific recommendation.",
	}, getRecommendationHandler(client))

	return server
}
