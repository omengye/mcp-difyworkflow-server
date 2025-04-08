package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/hb1707/dify-go-sdk/dify"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	apiKeys := os.Getenv("DIFY_API_KEYS")
	if apiKeys == "" {
		fmt.Println("DIFY_API_KEYS environment variable is required")
		return
	}

	workflowNames := os.Getenv("DIFY_WORKFLOW_NAME")
	if workflowNames == "" {
		fmt.Println("DIFY_WORKFLOW_NAME environment variable is required")
		return
	}
	workflows := parseAPIKeys(apiKeys, workflowNames)
	baseURL := flag.String("base-url", "http://localhost/v1", "Base URL for Dify API")
	flag.Parse()

	s := server.NewMCPServer(
		"Dify Workflow Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	listWorkflowsTool := mcp.NewTool("list_workflows",
		mcp.WithDescription("List authorized workflows"),
	)
	s.AddTool(listWorkflowsTool, listWorkflowsHandler(workflows))

	executeWorkflowTool := mcp.NewTool("execute_workflow",
		mcp.WithDescription("Execute a specified workflow"),
		mcp.WithString("workflow_name",
			mcp.Required(),
			mcp.Description("Name of the workflow to execute"),
		),
		mcp.WithString("input",
			mcp.Description("Input data for the workflow"),
		),
	)
	s.AddTool(executeWorkflowTool, executeWorkflowHandler(workflows, *baseURL))

	//if err := server.ServeStdio(s); err != nil {
	//	fmt.Printf("Server error: %v\n", err)
	//}

	if err := server.NewSSEServer(s, server.WithBaseURL("http://127.0.0.1:8096")).Start("127.0.0.1:8096"); err != nil {
		fmt.Printf("Server error: %v\n", err)
		return
	}

}

func parseAPIKeys(apiKeys, workflowNames string) map[string]string {
	workflows := make(map[string]string)

	apiKeyList := strings.Split(apiKeys, ",")
	workflowNameList := strings.Split(workflowNames, ",")

	if len(apiKeyList) != len(workflowNameList) {
		fmt.Printf("The number of API Keys does not match the number of workflow names\n")
		os.Exit(1)
	}

	for i, name := range workflowNameList {
		workflows[name] = apiKeyList[i]
	}

	return workflows
}

func listWorkflowsHandler(workflows map[string]string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workflowNames := make([]string, 0, len(workflows))
		for name := range workflows {
			workflowNames = append(workflowNames, name)
		}
		return mcp.NewToolResultText(fmt.Sprintf("Authorized workflows: %v", workflowNames)), nil
	}
}

func executeWorkflowHandler(workflows map[string]string, baseURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workflowName := request.Params.Arguments["workflow_name"].(string)
		apiKey, ok := workflows[workflowName]
		if !ok {
			return mcp.NewToolResultText(fmt.Sprintf("Workflow %s not found", workflowName)), nil
		}

		input := ""
		if in, ok := request.Params.Arguments["input"].(string); ok {
			input = in
		}

		response, err := callDifyWorkflowAPI(baseURL, apiKey, input)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Failed to execute workflow: %v", err)), nil
		}

		return response, nil
	}
}

func callDifyWorkflowAPI(baseURL, apiKey, input string) (*mcp.CallToolResult, error) {

	//ctx := context.Background()
	//serv := server.ServerFromContext(ctx)
	//serv.SendNotificationToClient(ctx, "", map[string]any{})

	c := dify.NewClient(apiKey,
		dify.WithBaseURL(baseURL),
	)

	req := dify.WorkflowRequest{
		ResponseMode: "streaming",
		Inputs: map[string]interface{}{
			"query": input,
		},
		User: "mcp-user",
	}

	ch := make(chan interface{})
	resp := &strings.Builder{}
	go func() {
		select {
		case data := <-ch:
			m := data.(map[string]interface{})
			resp.WriteString(m["result"].(string))
		}
	}()
	handler := &StreamMsgHandler{
		ResultChan: ch,
	}
	if err := c.WorkflowRunStreaming(req, handler); err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resp.String(),
			},
		},
	}, nil
}
