package main

import (
	"encoding/json"
)

type File struct {
	Type           string `json:"type"`
	TransferMethod string `json:"transfer_method"`
	URL            string `json:"url"`
	UploadFileID   string `json:"upload_file_id"`
}

// v0.7.x
type DifyLegacyRequest struct {
	Query            string                 `json:"query"`
	Inputs           map[string]interface{} `json:"inputs"`
	ResponseMode     string                 `json:"response_mode"`
	User             string                 `json:"user"`
	ConversationID   string                 `json:"conversation_id"`
	Files            []File                 `json:"files"`
	AutoGenerateName bool                   `json:"auto_generate_name"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// v0.7.x
type DifyLegacyResponse struct {
	MessageID          string                 `json:"message_id"`
	ConversationID     string                 `json:"conversation_id"`
	Mode               string                 `json:"mode"`
	Answer             string                 `json:"answer"`
	Metadata           map[string]interface{} `json:"metadata"`
	Usage              Usage                  `json:"usage"`
	RetrieverResources []interface{}          `json:"retriever_resources"`
	CreatedAt          int64                  `json:"created_at"`
}

// v0.8.x
type DifyWorkFlowRunRequest struct {
	Inputs       map[string]interface{} `json:"inputs"`
	ResponseMode string                 `json:"response_mode"`
	User         string                 `json:"user"`
	Files        []File                 `json:"files,omitempty"`
}

// v0.8.x
type DifyResponse struct {
	WorkflowRunID string `json:"workflow_run_id"`
	TaskID        string `json:"task_id"`
	Data          struct {
		ID          string                 `json:"id"`
		WorkflowID  string                 `json:"workflow_id"`
		Status      string                 `json:"status"`
		Outputs     map[string]interface{} `json:"outputs"`
		Error       string                 `json:"error"`
		ElapsedTime float64                `json:"elapsed_time"`
		TotalTokens int                    `json:"total_tokens"`
		TotalSteps  int                    `json:"total_steps"`
		CreatedAt   int64                  `json:"created_at"`
		FinishedAt  int64                  `json:"finished_at"`
	} `json:"data"`
}

// Event represents the base structure for all events
type Event struct {
	TaskID        string `json:"task_id"`
	WorkflowRunID string `json:"workflow_run_id,omitempty"`
	MessageID     string `json:"message_id,omitempty"`
	EventType     string `json:"event"`
	CreatedAt     int64  `json:"created_at"`
	Audio         string `json:"audio,omitempty"`
}

// WorkflowStartedEvent represents the workflow_started event
type WorkflowStartedEvent struct {
	Event
	Data struct {
		ID             string `json:"id"`
		WorkflowID     string `json:"workflow_id"`
		SequenceNumber int    `json:"sequence_number"`
		CreatedAt      int64  `json:"created_at"`
	} `json:"data"`
}

// NodeStartedEvent represents the node_started event
type NodeStartedEvent struct {
	Event
	Data struct {
		ID                string                   `json:"id"`
		NodeID            string                   `json:"node_id"`
		NodeType          string                   `json:"node_type"`
		Title             string                   `json:"title"`
		Index             int                      `json:"index"`
		PredecessorNodeID string                   `json:"predecessor_node_id"`
		Inputs            []map[string]interface{} `json:"inputs"`
		CreatedAt         int64                    `json:"created_at"`
	} `json:"data"`
}

// NodeFinishedEvent represents the node_finished event
type NodeFinishedEvent struct {
	Event
	Data struct {
		ID                string          `json:"id"`
		NodeID            string          `json:"node_id"`
		Index             int             `json:"index"`
		PredecessorNodeID string          `json:"predecessor_node_id,omitempty"`
		Inputs            json.RawMessage `json:"inputs,omitempty"`
		ProcessData       json.RawMessage `json:"process_data,omitempty"`
		Outputs           struct {
			Text  string          `json:"text"`
			Usage json.RawMessage `json:"usage,omitempty"`
		}
		Status            string  `json:"status"`
		Error             string  `json:"error,omitempty"`
		ElapsedTime       float64 `json:"elapsed_time,omitempty"`
		ExecutionMetadata struct {
			TotalTokens int    `json:"total_tokens,omitempty"`
			TotalPrice  string `json:"total_price,omitempty"`
			Currency    string `json:"currency,omitempty"`
		} `json:"execution_metadata"`
		CreatedAt int64 `json:"created_at"`
	} `json:"data"`
}

// WorkflowFinishedEvent represents the workflow_finished event
type WorkflowFinishedEvent struct {
	Event
	Data struct {
		ID         string `json:"id"`
		WorkflowID string `json:"workflow_id"`
		Status     string `json:"status"`
		Outputs    struct {
			Text string `json:"text"`
		}
		Error       string  `json:"error,omitempty"`
		ElapsedTime float64 `json:"elapsed_time,omitempty"`
		TotalTokens int     `json:"total_tokens,omitempty"`
		TotalSteps  int     `json:"total_steps"`
		CreatedAt   int64   `json:"created_at"`
		FinishedAt  int64   `json:"finished_at"`
	} `json:"data"`
}

type TextChunkEvent struct {
	TaskID        string `json:"task_id"`
	WorkflowRunID string `json:"workflow_run_id,omitempty"`
	EventType     string `json:"event"`
	Data          struct {
		Text                 string   `json:"text"`
		FromVariableSelector []string `json:"from_variable_selector"`
	}
}

// TTSMessageEvent represents the tts_message event
type TTSMessageEvent struct {
	Event
	ConversationID string `json:"conversation_id"`
}

// TTSMessageEndEvent represents the tts_message_end event
type TTSMessageEndEvent struct {
	Event
	ConversationID string `json:"conversation_id"`
}

// PingEvent represents the ping event
type PingEvent struct {
	Event
}
