package http

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yourorg/mcp-marketplace/backend/internal/models"
)

func mcpInitializeResultWithSDK() map[string]interface{} {
	impl := mcp.Implementation{
		Name:    "mcp-marketplace-hub",
		Version: "1.0.0",
	}
	return map[string]interface{}{
		"protocolVersion": "2025-11-05",
		"serverInfo": map[string]string{
			"name":    impl.Name,
			"version": impl.Version,
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]bool{
				"listChanged": false,
			},
		},
	}
}

func mcpToolsFromRoutesWithSDK(routes []models.HubRoute) []map[string]interface{} {
	sdkTools := make([]mcp.Tool, 0, len(routes))
	for _, route := range routes {
		if !route.Enabled {
			continue
		}
		sdkTools = append(sdkTools, mcp.Tool{
			Name:        route.ToolName,
			Description: "Tool proxied via personal marketplace hub.",
		})
	}

	out := make([]map[string]interface{}, 0, len(sdkTools))
	for _, tool := range sdkTools {
		out = append(out, map[string]interface{}{
			"name": tool.Name,
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			"description": tool.Description,
		})
	}
	return out
}

