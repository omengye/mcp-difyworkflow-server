package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
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
	baseURL := flag.String("base-url", "http://localhostr/v1", "Base URL for Dify API")
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

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
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
			return mcp.NewToolResultError(fmt.Sprintf("Workflow %s not found", workflowName)), nil
		}

		input := ""
		if in, ok := request.Params.Arguments["input"].(string); ok {
			input = in
		}

		response, err := callDifyWorkflowAPI(baseURL, apiKey, input)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to execute workflow: %v", err)), nil
		}

		return mcp.NewToolResultText(response), nil
	}
}

func callDifyWorkflowAPI(baseURL, apiKey, input string) (string, error) {
	client := &http.Client{}

	body := fmt.Sprintf(`{
		"inputs": {"message":"%s"},
		"response_mode": "blocking",
		"user": "mcp-server"
	}`, input)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/workflows/run", baseURL), strings.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var completionResponse DifyResponse
	if err := json.Unmarshal(respBody, &completionResponse); err != nil {
		return "", err
	}

	if completionResponse.Data.Status == "succeeded" {
		return fmt.Sprintf("%s", completionResponse.Data.Outputs["text"]), nil
	}

	return "", fmt.Errorf("Workflow execution failed: %s", respBody)
}
