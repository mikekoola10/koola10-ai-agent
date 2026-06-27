# Jarvis MCP Complete Guide

Jarvis is the orchestration layer for Koola10. This guide covers the full MCP setup for Jarvis.

## Architecture

Jarvis uses MCP as its primary interface for external tools.

1. **Koola10 Core (Go)**: Host of the main MCP hub.
2. **Spiral (TS)**: Consumer of MCP services for asset management.
3. **MCP Servers**: Independent processes (Node/Python) providing tools and resources.

## Deployment

Deploying the MCP hub involves:

1. Updating the Dockerfile with Node/Python.
2. Building the container.
3. Deploying to Fly.io or Render.

See `DEPLOYMENT_SUMMARY.md` for specific steps.
