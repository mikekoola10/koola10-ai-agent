import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StdioClientTransport } from "@modelcontextprotocol/sdk/client/stdio.js";

async function main() {
  const transport = new StdioClientTransport({
    command: "node",
    args: ["./node_modules/@modelcontextprotocol/server-everything/dist/index.js"],
  });

  const client = new Client(
    {
      name: "spiral-mcp-client",
      version: "1.0.0",
    },
    {
      capabilities: {
        resources: {},
        tools: {},
        prompts: {},
      },
    }
  );

  await client.connect(transport);

  console.log("Spiral MCP Client connected");

  const tools = await client.listTools();
  console.log("Available tools:", JSON.stringify(tools, null, 2));

  // Example: Call a tool if needed
  // const result = await client.callTool({
  //   name: "some-tool",
  //   arguments: { arg1: "value" },
  // });
  // console.log("Result:", result);
}

main().catch((error) => {
  console.error("Spiral MCP Client error:", error);
  process.exit(1);
});
