#!/usr/bin/env node
import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { runAgent } from "./agent.js";
import { keyEnvFor, loadConfig, loadDotEnv, type CliFlags } from "./config.js";
import { buildToolDefinitions } from "./tools/index.js";
import { color, firstLine, formatDuration } from "./util.js";

const VERSION = "0.4.0";

function printHelp(): void {
  console.log(`Nova — Manus-style autonomous AI agent (v${VERSION})

USAGE
  nova "<task description>" [options]

Give Nova an open-ended task. It plans, uses tools (shell, files, web),
iterates, and delivers a final report.

OPTIONS
  --mock                Use a scripted mock brain (no API key needed; for testing)
  --provider <name>     deepseek | anthropic | openai (default: deepseek)
  --model <name>        Override NOVA_MODEL (default: per-provider)
  --max-steps <n>       Override NOVA_MAX_STEPS (default: 25)
  --verbose, -v         Print full tool outputs
  --no-report           Do not save the final report to output/last-report.md
  --list-tools          List the tools Nova can use
  -h, --help            Show this help
  --version             Show version

ENVIRONMENT (or .env in the working directory)
  NOVA_PROVIDER         deepseek (default) | anthropic (Claude) | openai
  DEEPSEEK_API_KEY      Required for deepseek provider
  ANTHROPIC_API_KEY     Required for anthropic provider
  OPENAI_API_KEY        Required for openai provider
  NOVA_MODEL            Model name (per-provider defaults if unset)
  NOVA_MAX_STEPS        Max loop iterations
  NOVA_MAX_TOOL_OUTPUT  Max chars per tool result
  NOVA_TOOL_TIMEOUT_MS  Shell command timeout (ms)
  NOVA_SAVE_REPORT      1/0 — save report file
  GITHUB_TOKEN          Enables the github connector tool
  STRIPE_SECRET_KEY     Enables the stripe connector tool
  CLAWDBOT_CLI          Enables the clawdbot (OpenClaw) tool
  NOVA_BROWSER_HEADLESS 1/0 — headless browser (default 1)

EXAMPLES
  nova "Audit this repo: list files, find TODO markers, and write a summary"
  nova "Research the top 5 AI agent frameworks and compare them in a report"
  nova --mock "test the pipeline"

UI
  nova-ui (or: npm run ui)   Launch the Manus-style web interface
  Listens on $PORT (default 3000, 0.0.0.0). Auto-enables mock mode when no
  API key is set so the UI is demoable immediately. API: GET /api/health,
  GET /api/tasks, POST /api/tasks.
`);
}

interface ParsedArgs {
  task: string;
  flags: CliFlags;
  listTools: boolean;
  help: boolean;
  version: boolean;
}

function parseArgs(argv: string[]): ParsedArgs {
  const flags: CliFlags = {};
  const positional: string[] = [];
  const out = { task: "", flags, listTools: false, help: false, version: false };

  const next = (i: number, name: string): string => {
    const v = argv[i + 1];
    if (!v || v.startsWith("--")) {
      console.error(`nova: missing value for ${name}`);
      process.exit(1);
    }
    return v;
  };

  for (let i = 0; i < argv.length; i++) {
    const a = argv[i]!;
    switch (true) {
      case a === "--mock":
        flags.mock = true;
        break;
      case a === "--verbose" || a === "-v":
        flags.verbose = true;
        break;
      case a === "--no-report":
        flags.saveReport = false;
        break;
      case a === "--list-tools":
        out.listTools = true;
        break;
      case a === "--help" || a === "-h":
        out.help = true;
        break;
      case a === "--version":
        out.version = true;
        break;
      case a.startsWith("--max-steps="):
        flags.maxSteps = a.slice("--max-steps=".length);
        break;
      case a === "--max-steps":
        flags.maxSteps = next(i, "--max-steps");
        i += 1;
        break;
      case a.startsWith("--model="):
        flags.model = a.slice("--model=".length);
        break;
      case a === "--model":
        flags.model = next(i, "--model");
        i += 1;
        break;
      case a.startsWith("--provider="):
        flags.provider = a.slice("--provider=".length);
        break;
      case a === "--provider":
        flags.provider = next(i, "--provider");
        i += 1;
        break;
      default:
        positional.push(a);
    }
  }

  out.task = positional.join(" ").trim();
  return out;
}

async function main(): Promise<void> {
  const { task, flags, listTools, help, version } = parseArgs(process.argv.slice(2));

  if (help) {
    printHelp();
    return;
  }
  if (version) {
    console.log(VERSION);
    return;
  }

  loadDotEnv(process.cwd());
  const config = loadConfig({ ...flags, cwd: process.cwd() });

  if (listTools) {
    console.log("Nova tools:\n");
    for (const t of buildToolDefinitions()) {
      console.log(`  ${color.bold(t.function.name)}\n      ${t.function.description}`);
    }
    return;
  }

  if (!task) {
    console.error(color.red("nova: missing task. Run `nova --help` for usage."));
    process.exit(1);
  }
  if (!config.apiKey && !config.mock) {
    console.error(
      color.red(
        `nova: ${keyEnvFor(config.provider)} is not set for provider "${config.provider}".\n` +
          "  Add it to a .env file (see env.example), export it, or pass --mock to test without a key.",
      ),
    );
    process.exit(1);
  }

  // Banner
  const connectors: string[] = [];
  if (config.githubToken) connectors.push("github");
  if (config.stripeKey) connectors.push("stripe");
  if (config.clawdbotCli) connectors.push("clawdbot");
  console.log(color.dim("─".repeat(64)));
  console.log(
    `${color.magenta("⚡ Nova")} ${color.dim(`v${VERSION}`)} · ${color.cyan(config.provider)}/${color.cyan(config.model)}` +
      (config.mock ? ` · ${color.yellow("mock brain")}` : "") +
      ` · max ${config.maxSteps} steps` +
      (connectors.length ? ` · ${color.green(`connectors: ${connectors.join(", ")}`)}` : ""),
  );
  console.log(color.dim(`task: ${task.slice(0, 200)}${task.length > 200 ? "…" : ""}`));
  console.log(color.dim("─".repeat(64)));

  const result = await runAgent(task, config, {
    onStep: ({ step, toolNames, elapsedMs }) => {
      console.log(
        `${color.dim(`[${step}]`)} ${color.bold(toolNames.join(", "))} ${color.dim(`· ${formatDuration(elapsedMs)}`)}`,
      );
    },
    onTool: ({ name, output, elapsedMs }) => {
      if (config.verbose) {
        console.log(`${color.dim("  ↳")} ${color.cyan(name)} ${color.dim(`(${formatDuration(elapsedMs)})`)}`);
        for (const line of output.split(/\r?\n/).slice(0, 20)) {
          console.log(color.dim(`    │ ${line}`));
        }
      } else {
        console.log(`  ${color.dim("↳")} ${color.cyan(name)}: ${firstLine(output)}`);
      }
    },
    onError: (err) => {
      console.error(color.red(`  ✖ ${err.message}`));
    },
  });

  // Final report
  console.log();
  console.log(color.green("━".repeat(64)));
  console.log(color.bold("Final report"));
  console.log(result.report);
  console.log(color.dim("━".repeat(64)));
  console.log(
    color.dim(
      `⚡ ${formatDuration(result.durationMs)} · ${result.steps} steps · ${result.toolCalls} tool calls · model ${config.model}`,
    ),
  );

  // Save the report artifact
  if (config.saveReport) {
    try {
      const dir = join(config.cwd, "output");
      mkdirSync(dir, { recursive: true });
      const file = join(dir, "last-report.md");
      writeFileSync(
        file,
        `# Nova report\n\n- **Task:** ${task}\n- **Model:** ${config.model}\n- **Duration:** ${formatDuration(result.durationMs)}\n- **Steps:** ${result.steps}\n- **Tool calls:** ${result.toolCalls}\n\n${result.report}\n`,
        "utf8",
      );
      console.log(color.dim(`📄 report saved → output/last-report.md`));
    } catch (err) {
      console.error(color.red(`  ✖ could not save report: ${(err as Error).message}`));
    }
  }
}

main().catch((err) => {
  console.error(color.red(`nova: ${(err as Error).message}`));
  process.exit(1);
});
