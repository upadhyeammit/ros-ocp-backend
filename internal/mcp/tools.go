package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// listRecommendationsHandler returns a handler that uses the given API client (closed over by server).
func listRecommendationsHandler(client *APIClient) mcp.ToolHandlerFor[ListRecommendationsParams, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListRecommendationsParams) (*mcp.CallToolResult, any, error) {
		body, err := client.ListRecommendations(input)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: body}},
		}, nil, nil
	}
}

// getRecommendationHandler returns a handler that uses the given API client (closed over by server).
func getRecommendationHandler(client *APIClient) mcp.ToolHandlerFor[GetRecommendationInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetRecommendationInput) (*mcp.CallToolResult, any, error) {
		if input.RecommendationID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "recommendation_id is required"}},
				IsError: true,
			}, nil, nil
		}
		params := GetRecommendationParams{
			CPUUnit:    input.CPUUnit,
			MemoryUnit: input.MemoryUnit,
			TrueUnits:  input.TrueUnits,
		}
		body, err := client.GetRecommendation(input.RecommendationID, params)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
				IsError: true,
			}, nil, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: body}},
		}, nil, nil
	}
}

// GetRecommendationInput is the MCP tool input for get_recommendation.
type GetRecommendationInput struct {
	RecommendationID string `json:"recommendation_id" jsonschema:"required" jsonschema_description:"UUID of the recommendation"`
	CPUUnit          string `json:"cpu_unit,omitempty" jsonschema_description:"CPU unit: millicores or cores"`
	MemoryUnit       string `json:"memory_unit,omitempty" jsonschema_description:"Memory unit: bytes, MiB, GiB"`
	TrueUnits        string `json:"true_units,omitempty" jsonschema_description:"Show real-world units: true or false"`
}
