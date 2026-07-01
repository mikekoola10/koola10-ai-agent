package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"
)

type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
	ID      interface{}     `json:"id"`
}

type MCPClient struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	pending   map[interface{}]chan *MCPResponse
	pendingMu sync.Mutex
	id        int
	idMu      sync.Mutex
}

func NewMCPClient(command string, args ...string) (*MCPClient, error) {
	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	client := &MCPClient{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		pending: make(map[interface{}]chan *MCPResponse),
	}

	go client.listen()

	// 1. Initialize
	initParams := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]string{
			"name":    "koola10-go-client",
			"version": "1.0.0",
		},
	}
	_, err = client.Call("initialize", initParams)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("initialization failed: %w", err)
	}

	// 2. Initialized Notification
	err = client.Notify("notifications/initialized", nil)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("initialized notification failed: %w", err)
	}

	return client, nil
}

func (c *MCPClient) listen() {
	reader := bufio.NewReader(c.stdout)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("MCP client read error: %v", err)
			}
			return
		}

		var resp MCPResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			log.Printf("MCP client unmarshal error: %v", err)
			continue
		}

		if resp.ID != nil {
			c.pendingMu.Lock()
			ch, ok := c.pending[resp.ID]
			if ok {
				ch <- &resp
				delete(c.pending, resp.ID)
			}
			c.pendingMu.Unlock()
		}
	}
}

func (c *MCPClient) Call(method string, params interface{}) (json.RawMessage, error) {
	c.idMu.Lock()
	c.id++
	id := c.id
	c.idMu.Unlock()

	p, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req := MCPRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  p,
		ID:      id,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	ch := make(chan *MCPResponse, 1)
	c.pendingMu.Lock()
	c.pending[id] = ch
	c.pendingMu.Unlock()

	if _, err := fmt.Fprintf(c.stdin, "%s\n", data); err != nil {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return nil, err
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("MCP error: %v", resp.Error)
		}
		return resp.Result, nil
	case <-time.After(30 * time.Second):
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
		return nil, fmt.Errorf("MCP request timed out")
	}
}

func (c *MCPClient) Notify(method string, params interface{}) error {
	p, err := json.Marshal(params)
	if err != nil {
		return err
	}

	req := MCPRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  p,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.stdin, "%s\n", data)
	return err
}

func (c *MCPClient) Close() error {
	c.stdin.Close()
	return c.cmd.Wait()
}

// Global registry for MCP clients
var (
	mcpClients = make(map[string]*MCPClient)
	mcpMu      sync.RWMutex
)

func RegisterMCPClient(name string, command string, args ...string) error {
	client, err := NewMCPClient(command, args...)
	if err != nil {
		return err
	}
	mcpMu.Lock()
	mcpClients[name] = client
	mcpMu.Unlock()
	return nil
}

func CallMCP(clientName string, method string, params interface{}) (json.RawMessage, error) {
	mcpMu.RLock()
	client, ok := mcpClients[clientName]
	mcpMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("MCP client %s not found", clientName)
	}
	return client.Call(method, params)
}

func ShutdownMCPClients() {
	mcpMu.Lock()
	defer mcpMu.Unlock()
	for name, client := range mcpClients {
		log.Printf("Shutting down MCP client: %s", name)
		client.Close()
	}
	mcpClients = make(map[string]*MCPClient)
}
