# MCP Integration Guide

This guide explains how the Model Context Protocol (MCP) is integrated into the Koola10 ecosystem.

## Overview

MCP allows Koola10 to connect with external tools and data sources via a standardized protocol. Our implementation uses a Go-based client for core services and a TypeScript client for the Spiral ecosystem.

## Go Client (Koola10)

The Go client is located in `tools/mcp_client.go`. It connects to MCP servers over standard input/output (stdio).

### Configuration

MCP clients can be registered in `main.go`:

```go
tools.RegisterMCPClient("everything", "node", "./node_modules/@modelcontextprotocol/server-everything/dist/index.js")
```

### Usage

Call an MCP tool from anywhere in the application:

```go
res, err := tools.CallMCP("everything", "tools/list", nil)
```

## TypeScript Client (Spiral)

The TypeScript client is located in `spiral-mcp-client.ts`. It provides similar functionality for Node.js-based services.

## Testing

A test endpoint is available at `/admin/mcp-test`. It attempts to list tools from the 'everything' MCP server.
