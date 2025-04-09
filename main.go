package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hb1707/dify-go-sdk/dify"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type DifyApp struct {
	ApiKey     string             `json:"apiKey"`
	AppInfo    dify.AppInfo       `json:"appInfo"`
	Parameters dify.AppParameters `json:"parameters"`
}

var (
	DifyBaseUrl  string
	DifyWorkflow map[string]*DifyApp = make(map[string]*DifyApp)
)

func main() {
	apiKeys := os.Getenv("DIFY_API_KEYS")
	if apiKeys == "" {
		fmt.Println("DIFY_API_KEYS environment variable is required")
		return
	}

	baseURL := flag.String("base-url", "http://localhost/v1", "Base URL for Dify API")
	flag.Parse()

	DifyBaseUrl = *baseURL
	if err := parseAPIKeys(apiKeys); err != nil {
		fmt.Printf("parse dify workflow error: %v\n", err)
		return
	}

	s := server.NewMCPServer(
		"Dify Workflow Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	for _, difyApp := range DifyWorkflow {
		inputSchema := mcp.ToolInputSchema{}
		for _, input := range difyApp.Parameters.UserInputForm {
			for k, v := range input {
				inputSchema.Type = k
				paramInfo := v.(map[string]interface{})
				propertyName := paramInfo["variable"].(string)
				if inputSchema.Properties == nil {
					inputSchema.Properties = make(map[string]interface{})
				}
				inputSchema.Properties[propertyName] = map[string]interface{}{
					"type":        k,
					"description": paramInfo["label"].(string),
				}
				if paramInfo["required"] != nil {
					inputSchema.Required = append(inputSchema.Required, propertyName)
				}
			}
		}

		workflowsTool := mcp.Tool{
			Name:        difyApp.AppInfo.Name,
			Description: difyApp.AppInfo.Description,
			InputSchema: inputSchema,
		}
		s.AddTool(workflowsTool, executeWorkflowHandler())
	}

	//if err := server.ServeStdio(s); err != nil {
	//	fmt.Printf("Server error: %v\n", err)
	//}

	if err := server.NewSSEServer(s, server.WithBaseURL("http://127.0.0.1:8096")).Start("127.0.0.1:8096"); err != nil {
		fmt.Printf("Server error: %v\n", err)
		return
	}

}

func parseAPIKeys(apiKeys string) error {

	apiKeyList := strings.Split(apiKeys, ",")

	for _, apiKey := range apiKeyList {
		c := dify.NewClient(apiKey,
			dify.WithBaseURL(DifyBaseUrl),
		)
		info, err := c.GetAppInfo()
		if err != nil {
			fmt.Printf("GetAppInfo error: %v\n", err)
			return err
		}
		parameters, err := c.GetAppParameters()
		if err != nil {
			fmt.Printf("GetAppParameters error: %v\n", err)
			return err
		}
		DifyWorkflow[info.Name] = &DifyApp{
			ApiKey:     apiKey,
			AppInfo:    *info,
			Parameters: *parameters,
		}
	}

	return nil
}

func executeWorkflowHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workflowName := request.Params.Name
		difyApp, ok := DifyWorkflow[workflowName]
		if !ok {
			return mcp.NewToolResultText(fmt.Sprintf("Workflow %s not found", workflowName)), nil
		}
		inputs := make(map[string]interface{})
		for _, forms := range difyApp.Parameters.UserInputForm {
			for _, v := range forms {
				paramInfo := v.(map[string]interface{})
				propertyName := paramInfo["variable"].(string)
				inputs[propertyName] = request.Params.Arguments[propertyName]
			}
		}

		response, err := callDifyWorkflowAPI(difyApp, inputs)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Failed to execute workflow: %v", err)), nil
		}

		return response, nil
	}
}

func callDifyWorkflowAPI(difyApp *DifyApp, input map[string]interface{}) (*mcp.CallToolResult, error) {

	c := dify.NewClient(difyApp.ApiKey,
		dify.WithBaseURL(DifyBaseUrl),
	)

	req := dify.WorkflowRequest{
		ResponseMode: "streaming",
		Inputs:       input,
		User:         "mcp-user",
	}

	ch := make(chan interface{})
	resp := &strings.Builder{}
	go func() {
		select {
		case data := <-ch:
			m := data.(map[string]interface{})
			if str, err := json.Marshal(m); err != nil {
				return
			} else {
				resp.WriteString(string(str))
			}
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
