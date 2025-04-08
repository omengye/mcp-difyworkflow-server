package main

import (
	"fmt"
	"github.com/hb1707/dify-go-sdk/dify"
)

type StreamMsgHandler struct {
	ResultChan chan interface{}
}

func (h *StreamMsgHandler) OnMessageWorkflow(response *dify.WorkflowStreamResponse) error {
	if response.StreamResponse.Event == "node_finished" && response.Data.NodeType == "end" {
		fmt.Printf("StreamMessageHandler.OnMessageWorkflow: %v\n", response.Data.Inputs)
		h.ResultChan <- response.Data.Inputs
	}
	return nil
}

func (h *StreamMsgHandler) OnMessage(resp *dify.MessageStreamResponse) error {
	return nil
}

func (h *StreamMsgHandler) OnMessageEnd(resp *dify.MessageEndStreamResponse) error {
	return nil
}

func (h *StreamMsgHandler) OnTTS(resp *dify.TTSStreamResponse) error {
	return nil
}

func (h *StreamMsgHandler) OnError(err error) error {
	return nil
}
