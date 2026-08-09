#!/usr/bin/env node
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { runAgent } from "./agent.js";
import { keyEnvFor, loadConfig, loadDotEnv, type CliFlags } from "./config.js";
import { buildToolDefinitions, verifyConnectors } from "./tools/index.js";
import { automationTool, buildDailyReport, reportDeliveryProvider } from "./tools/automations.js";
import { applyVaultOverrides, vaultDelete, vaultGet, vaultInfo, vaultList, vaultPushToRemote, vaultSet, vaultSyncFromRemote } from "./tools/vault.js";
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
  --verify-connectors   Ping every configured connector and report OK/FAIL
  --report              Build and send the daily Nova report email now
  --vault <action>      list | set <name> <value> | get <name> | rm <name>
  --job <file>          Run an agent job from a JSON file ({ "task": "..." })
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
  COMPOSIO_API_KEY      Enables shared Composio tools across connected apps
  COMPOSIO_USER_ID      Stable Composio user identity (default: nova-user)
  ZAPIER_WEBHOOK_URL    Enables the zapier tool (Zapier Catch Hook URL)
  N8N_WEBHOOK_URL       Enables the n8n tool (n8n Cloud production webhook)
  N8N_API_KEY           Optional header value for n8n Header Auth webhooks
  MAKE_WEBHOOK_URL      Enables the make tool (Make custom webhook URL)
  HUGGINGFACE_TOKEN     Enables the huggingface tool (Inference API + Hub search)
  NOVA_BROWSER_HEADLESS 1/0 — headless browser (default 1)

VERIFICATION
  Run \`nova --verify-connectors\` after adding any key. It pings each
  connector with a harmless test call and prints an OK/FAIL report.

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
  verifyConnectors: boolean;
  report: boolean;
  vault: string[] | null;
  job: string;
  help: boolean;
  version: boolean;
}

function parseArgs(argv: string[]): ParsedArgs {
  const flags: CliFlags = {};
  const positional: string[] = [];
  const out: ParsedArgs = {
    task: "",
    flags,
    listTools: false,
    verifyConnectors: false,
    report: false,
    vault: null,
    job: "",
    help: false,
    version: false,
  };

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
      case a === "--verify-connectors" || a === "--check-connectors":
        out.verifyConnectors = true;
        break;
      case a === "--report":
        out.report = true;
        break;
      case a === "--vault":
        out.vault = [];
        break;
      case a === "--job":
        out.job = next(i, "--job");
        i += 1;
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
        if (Array.isArray(out.vault)) out.vault.push(a);
        else positional.push(a);
    }
  }

  out.task = positional.join(" ").trim();
  return out;
}

async function main(): Promise<void> {
  const { task, flags, listTools, verifyConnectors: runVerify, report: sendReport, vault: vaultArgs, job, help, version } = parseArgs(process.argv.slice(2));

  if (help) {
    printHelp();
    return;
  }
  if (version) {
    console.log(VERSION);
    return;
  }

  loadDotEnv(process.cwd());
  const restored = await vaultSyncFromRemote();
  if (restored.error) console.log(color.yellow(`  vault remote restore skipped — ${restored.error}`));
  const config = applyVaultOverrides(loadConfig({ ...flags, cwd: process.cwd() }));

  if (vaultArgs) {
    const [action, name, value] = vaultArgs;
    if (!action) {
      console.error(color.red("nova: --vault needs an action: list | set <name> <value> | get <name> | rm <name>"));
      process.exit(1);
    }
    const info = vaultInfo();
    switch (action) {
      case "list": {
        const names = vaultList();
        console.log(color.bold(`\nVault — ${info.dir} (${info.count} entries${info.usingEnvKey ? ", NOVA_VAULT_KEY set" : ""})`));
        if (!names.length) console.log(color.dim("  empty — store credentials with: nova --vault set NAME value"));
        else names.forEach((n) => console.log(`  • ${n}`));
        return;
      }
      case "set": {
        if (!name || !value) {
          console.error(color.red("nova: --vault set needs <name> <value>"));
          process.exit(1);
        }
        const res = vaultSet(name, value);
        if (!res.ok) {
          console.error(color.red(`nova: ${res.error}`));
          process.exit(1);
        }
        const sync = await vaultPushToRemote();
        const syncText = sync.ok ? color.dim("(remote backup synced)") : color.yellow(`(remote backup FAILED: ${sync.error})`);
        console.log(color.green(`  Stored ${name} (encrypted) — ${info.dir} ${syncText}`));
        return;
      }
      case "get": {
        if (!name) {
          console.error(color.red("nova: --vault get needs <name>"));
          process.exit(1);
        }
        const v = vaultGet(name);
        if (v === null) {
          console.error(color.red(`nova: no vault entry named ${name}`));
          process.exit(1);
        }
        console.log(v);
        return;
      }
      case "rm":
      case "delete": {
        if (!name) {
          console.error(color.red(`nova: --vault ${action} needs <name>`));
          process.exit(1);
        }
        const removed = vaultDelete(name);
        if (removed) await vaultPushToRemote();
        console.log(removed ? color.green(`  Removed ${name} from the vault`) : color.yellow(`  No vault entry named ${name}`));
        return;
      }
      default:
        console.error(color.red(`nova: unknown --vault action "${action}" (list | set | get | rm)`));
        process.exit(1);
    }
  }

  if (listTools) {
    console.log("Nova tools:\n");
    for (const t of buildToolDefinitions()) {
      console.log(`  ${color.bold(t.function.name)}\n      ${t.function.description}`);
    }
    return;
  }

  if (runVerify) {
    console.log(color.bold("\nNova connector verification"));
    console.log(color.dim("  pings each configured connector with a harmless test call\n"));
    const checks = await verifyConnectors(config);
    let passed = 0;
    let notConfigured = 0;
    let failed = 0;
    for (const c of checks) {
      if (!c.configured) {
        notConfigured += 1;
        console.log(`  ${color.dim("·")} ${color.bold(c.provider.padEnd(10))} ${color.dim("not configured —")} ${c.detail}`);
        continue;
      }
      if (c.ok) {
        passed += 1;
        console.log(`  ${color.green("✔")} ${color.bold(c.provider.padEnd(10))} ${color.green("OK")} — ${c.detail}`);
      } else {
        failed += 1;
        console.log(`  ${color.red("✖")} ${color.bold(c.provider.padEnd(10))} ${color.red("FAIL")} — ${c.detail}`);
      }
    }
    console.log(color.dim("─".repeat(48)));
    console.log(`  ${passed} verified · ${failed} failed · ${notConfigured} not configured`);
    if (failed === 0 && notConfigured === 0) {
      console.log(color.green("  All connectors are operating ✔\n"));
    } else if (failed === 0) {
      console.log(color.yellow("  Configured connectors all pass. Add the missing keys and re-run to complete the setup.\n"));
    } else {
      console.log(color.red("  Fix the failures above, then re-run --verify-connectors.\n"));
    }
    return;
  }

  if (sendReport) {
    const provider = reportDeliveryProvider(config);
    if (!provider) {
      console.error(
        color.red(
          "nova: no report delivery target configured.\n" +
            "  Set NOVA_REPORT_PROVIDER (zapier|make|n8n) or one of ZAPIER_WEBHOOK_URL / MAKE_WEBHOOK_URL / N8N_WEBHOOK_URL.",
        ),
      );
      process.exit(1);
    }
    console.log(color.bold("\nNova daily report"));
    const checks = await verifyConnectors(config);
    const report = buildDailyReport({ checks, version: VERSION });
    console.log(`  ${color.dim("deliver via:  ")} ${provider}`);
    console.log(`  ${color.dim("subject:      ")} ${report.subject}`);
    console.log(color.dim(report.body.split("\n").map((l) => `  ${l}`).join("\n")));
    const outText = await automationTool(config, provider, "trigger", { payload: report.payload });
    if (String(outText).startsWith("ERROR")) {
      console.log(`  ${color.red("result: FAIL")} — ${outText.slice(0, 200)}`);
      console.log(color.red("  Report was NOT delivered — fix the target and re-run.\n"));
      process.exit(1);
    }
    console.log(`  ${color.green("result: OK")} — ${outText.slice(0, 200)}`);
    console.log(color.green("  Report delivered ✔\n"));
    return;
  }

  let jobTask = task;
  if (job) {
    try {
      const raw = JSON.parse(readFileSync(job, "utf8")) as { task?: unknown };
      if (typeof raw.task !== "string" || !raw.task.trim())
        throw new Error(`"${job}" must contain a non-empty "task" string field`);
      jobTask = raw.task.trim();
    } catch (err) {
      console.error(color.red(`nova: cannot load job — ${err instanceof Error ? err.message : String(err)}`));
      process.exit(1);
    }
  }

  if (!jobTask) {
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
  if (config.composioApiKey) connectors.push("composio");
  if (config.zapierWebhookUrl) connectors.push("zapier");
  if (config.n8nWebhookUrl) connectors.push("n8n");
  if (config.makeWebhookUrl) connectors.push("make");
  if (config.huggingfaceApiKey) connectors.push("huggingface");
  console.log(color.dim("─".repeat(64)));
  console.log(
    `${color.magenta("⚡ Nova")} ${color.dim(`v${VERSION}`)} · ${color.cyan(config.provider)}/${color.cyan(config.model)}` +
      (config.mock ? ` · ${color.yellow("mock brain")}` : "") +
      ` · max ${config.maxSteps} steps` +
      (connectors.length ? ` · ${color.green(`connectors: ${connectors.join(", ")}`)}` : ""),
  );
  console.log(color.dim(`task: ${jobTask.slice(0, 200)}${jobTask.length > 200 ? "…" : ""}`));
  console.log(color.dim("─".repeat(64)));

  const result = await runAgent(jobTask, config, {
    onStep: ({ step, toolNames, elapsedMs }) => {
      console.log(`${color.dim(`[${step}]`)} ${color.bold(toolNames.join(", "))} ${color.dim(`· ${formatDuration(elapsedMs)}`)}`);
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
  console.log(color.dim(`⚡ ${formatDuration(result.durationMs)} · ${result.steps} steps · ${result.toolCalls} tool calls · model ${config.model}`));

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
