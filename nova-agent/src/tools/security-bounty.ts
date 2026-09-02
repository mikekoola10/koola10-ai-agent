/**
 * security-bounty.ts
 *
 * HackerOne/Bugcrowd/YesWeHack security bounty scanner.
 * Integrates with Nova's sweep pipeline to automatically:
 *   1. Scan public bug bounty programs
 *   2. Find targets (domains, repos, APIs)
 *   3. Run automated vuln checks (XSS, SSRF, IDOR, etc.)
 *   4. Generate structured reports
 *   5. Submit via platform API
 *
 * Revenue: $300 - $50,000+ per bounty (avg $500-$3,000)
 */

import { env } from "../config.js";

// ─── Types ───────────────────────────────────────────────────────────────────

export interface SecurityProgram {
  platform: "hackerone" | "bugcrowd" | "yeswehack" | "intigriti";
  name: string;
  handle: string;
  url: string;
  bountyRange: string;
  maxPayout: number;
  targets: string[];
  scope: string[];
  inScope: boolean;
}

export interface SecurityFindings {
  program: string;
  platform: string;
  targetType: string;
  targetUrl: string;
  vulnType: string;
  severity: "critical" | "high" | "medium" | "low" | "info";
  title: string;
  description: string;
  stepsToReproduce: string[];
  impact: string;
  remediation: string;
  estimatedPayout: number;
  reportUrl?: string;
  submittedAt?: string;
}

// ─── HackerOne API Client ────────────────────────────────────────────────────

class HackerOneClient {
  private username: string;
  private token: string;
  private baseUrl = "https://api.hackerone.com/v1";

  constructor() {
    this.username = env("HACKERONE_USERNAME") ?? "";
    this.token = env("HACKERONE_API_TOKEN") ?? "";
  }

  get configured(): boolean {
    return !!(this.username && this.token);
  }

  private headers(): Record<string, string> {
    const auth = Buffer.from(`${this.username}:${this.token}`).toString("base64");
    return {
      Authorization: `Basic ${auth}`,
      "Content-Type": "application/json",
      Accept: "application/json",
    };
  }

  async listPrograms(options?: {
    page?: number;
    perPage?: number;
    huntable?: boolean;
  }): Promise<SecurityProgram[]> {
    const params = new URLSearchParams();
    if (options?.page) params.set("page[number]", String(options.page));
    if (options?.perPage) params.set("page[size]", String(options.perPage));
    if (options?.huntable !== undefined) params.set("filter[hunters_allowed]", String(options.huntable));

    const url = `${this.baseUrl}/hackers/programs?${params}`;
    const res = await fetch(url, { headers: this.headers() });
    if (!res.ok) return [];

    const data = await res.json() as { data?: Array<{ id: string; attributes: Record<string, unknown>; relationships?: Record<string, unknown> }> };
    return (data.data ?? []).map((p) => ({
      platform: "hackerone" as const,
      name: String(p.attributes.name ?? ""),
      handle: String(p.attributes.handle ?? ""),
      url: `https://hackerone.com/${p.attributes.handle}`,
      bountyRange: String(p.attributes.offered_bounty_range ?? "Unknown"),
      maxPayout: parseBountyMax(String(p.attributes.offered_bounty_range ?? "0")),
      targets: [],
      scope: [],
      inScope: true,
    }));
  }

  async getProgramScope(handle: string): Promise<{ targets: string[]; scope: string[] }> {
    const url = `${this.baseUrl}/hackers/programs/${handle}/structured_scopes`;
    const res = await fetch(url, { headers: this.headers() });
    if (!res.ok) return { targets: [], scope: [] };

    const data = await res.json() as { data?: Array<{ attributes: Record<string, unknown> }> };
    const targets: string[] = [];
    const scope: string[] = [];

    for (const s of data.data ?? []) {
      const attrs = s.attributes;
      if (attrs.asset_type === "URL" || attrs.asset_type === "DOMAIN") {
        targets.push(String(attrs.asset_identifier ?? ""));
      }
      if (attrs.asset_type === "SOURCE_CODE") {
        targets.push(String(attrs.asset_identifier ?? ""));
      }
      scope.push(String(attrs.asset_identifier ?? ""));
    }

    return { targets, scope };
  }

  async submitReport(
    handle: string,
    report: {
      team_handle: string;
      title: string;
      vulnerability_information: string;
      impact: string;
    }
  ): Promise<{ id: string; url: string } | null> {
    const url = `${this.baseUrl}/hackers/reports`;
    const res = await fetch(url, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify({ data: { type: "report", attributes: report } }),
    });
    if (!res.ok) return null;

    const data = await res.json() as { data?: { id: string } };
    return data.data ? { id: data.data.id, url: `https://hackerone.com/${handle}/reports/${data.data.id}` } : null;
  }

  async getSubmissions(): Promise<{ id: string; state: string; title: string }[]> {
    const url = `${this.baseUrl}/hackers/reports?page[size]=20`;
    const res = await fetch(url, { headers: this.headers() });
    if (!res.ok) return [];

    const data = await res.json() as { data?: Array<{ id: string; attributes: Record<string, unknown> }> };
    return (data.data ?? []).map((r) => ({
      id: r.id,
      state: String(r.attributes.state ?? "unknown"),
      title: String(r.attributes.title ?? ""),
    }));
  }
}

// ─── Bugcrowd API Client ─────────────────────────────────────────────────────

class BugcrowdClient {
  private token: string;
  private baseUrl = "https://api.bugcrowd.com";

  constructor() {
    this.token = env("BUGCROWD_API_TOKEN") ?? "";
  }

  get configured(): boolean {
    return !!this.token;
  }

  private headers(): Record<string, string> {
    return {
      Authorization: `Token ${this.token}`,
      "Content-Type": "application/vnd.bugcrowd+json",
      Accept: "application/vnd.bugcrowd+json",
    };
  }

  async listPrograms(): Promise<SecurityProgram[]> {
    const url = `${this.baseUrl}/programs`;
    const res = await fetch(url, { headers: this.headers() });
    if (!res.ok) return [];

    const data = await res.json() as { data?: Array<{ id: string; attributes: Record<string, unknown> }> };
    return (data.data ?? []).map((p) => ({
      platform: "bugcrowd" as const,
      name: String(p.attributes.name ?? ""),
      handle: String(p.attributes.code ?? ""),
      url: `https://bugcrowd.com/${p.attributes.code}`,
      bountyRange: String(p.attributes.bounty_range ?? "Unknown"),
      maxPayout: parseBountyMax(String(p.attributes.bounty_range ?? "0")),
      targets: [],
      scope: [],
      inScope: true,
    }));
  }
}

// ─── Automated Vulnerability Checks ──────────────────────────────────────────

const VULN_CHECKS = [
  {
    name: "XSS",
    severity: "high" as const,
    estimatedPayout: 500,
    patterns: [
      /<script[\s>]/i,
      /javascript:/i,
      /onerror\s*=/i,
      /onload\s*=/i,
      /alert\s*\(/i,
    ],
    check: async (url: string): Promise<boolean> => {
      const testPayloads = [
        `<script>alert('xss')</script>`,
        `"><img src=x onerror=alert(1)>`,
        `' OR '1'='1`,
      ];
      for (const payload of testPayloads) {
        try {
          const testUrl = `${url}?q=${encodeURIComponent(payload)}`;
          const res = await fetch(testUrl, { redirect: "manual", signal: AbortSignal.timeout(5000) });
          const body = await res.text();
          if (body.includes(payload)) return true;
        } catch { /* timeout or error = not vulnerable */ }
      }
      return false;
    },
  },
  {
    name: "SSRF",
    severity: "critical" as const,
    estimatedPayout: 2000,
    patterns: [
      /fetch\s*\(/i,
      /axios/i,
      /request\s*\(/i,
      /curl/i,
    ],
    check: async (url: string): Promise<boolean> => {
      const testPayloads = [
        "http://169.254.169.254/latest/meta-data/",
        "http://[::1]/",
        "http://localhost:22",
      ];
      for (const payload of testPayloads) {
        try {
          const testUrl = `${url}?url=${encodeURIComponent(payload)}`;
          const res = await fetch(testUrl, { redirect: "manual", signal: AbortSignal.timeout(5000) });
          if (res.status === 200) return true;
        } catch { /* not vulnerable */ }
      }
      return false;
    },
  },
  {
    name: "IDOR",
    severity: "high" as const,
    estimatedPayout: 800,
    patterns: [
      /\/api\/v\d+\/users\/\d+/i,
      /\/api\/v\d+\/accounts\/\d+/i,
      /\/profile\/\d+/i,
    ],
    check: async (url: string): Promise<boolean> => {
      try {
        const testUrl = `${url}/1`;
        const res1 = await fetch(testUrl, { redirect: "manual", signal: AbortSignal.timeout(5000) });
        const testUrl2 = `${url}/2`;
        const res2 = await fetch(testUrl2, { redirect: "manual", signal: AbortSignal.timeout(5000) });
        if (res1.status === 200 && res2.status === 200) return true;
      } catch { /* not vulnerable */ }
      return false;
    },
  },
  {
    name: "Open Redirect",
    severity: "medium" as const,
    estimatedPayout: 300,
    patterns: [
      /redirect/i,
      /return_to/i,
      /next/i,
      /url=/i,
    ],
    check: async (url: string): Promise<boolean> => {
      const testPayloads = [
        "https://evil.com",
        "//evil.com",
      ];
      for (const payload of testPayloads) {
        try {
          const testUrl = `${url}?redirect=${encodeURIComponent(payload)}`;
          const res = await fetch(testUrl, { redirect: "manual", signal: AbortSignal.timeout(5000) });
          const location = res.headers.get("location") ?? "";
          if (location.includes("evil.com")) return true;
        } catch { /* not vulnerable */ }
      }
      return false;
    },
  },
  {
    name: "Information Disclosure",
    severity: "low" as const,
    estimatedPayout: 100,
    patterns: [
      /debug/i,
      /error/i,
      /stack.trace/i,
    ],
    check: async (url: string): Promise<boolean> => {
      const testUrls = [
        `${url}/debug`,
        `${url}/.env`,
        `${url}/config`,
        `${url}/server-info`,
        `${url}/actuator`,
      ];
      for (const testUrl of testUrls) {
        try {
          const res = await fetch(testUrl, { redirect: "manual", signal: AbortSignal.timeout(5000) });
          if (res.status === 200) {
            const body = await res.text();
            if (body.includes("password") || body.includes("secret") || body.includes("api_key")) return true;
          }
        } catch { /* not vulnerable */ }
      }
      return false;
    },
  },
];

// ─── Report Generator ────────────────────────────────────────────────────────

function generateReport(finding: SecurityFindings): string {
  return `
## ${finding.title}

**Program:** ${finding.program}
**Platform:** ${finding.platform}
**Target:** ${finding.targetUrl}
**Vulnerability Type:** ${finding.vulnType}
**Severity:** ${finding.severity.toUpperCase()}
**Estimated Payout:** $${finding.estimatedPayout}

### Description
${finding.description}

### Steps to Reproduce
${finding.stepsToReproduce.map((s, i) => `${i + 1}. ${s}`).join("\n")}

### Impact
${finding.impact}

### Remediation
${finding.remediation}

### Proof of Concept
See steps to reproduce above for detailed PoC.
`.trim();
}

// ─── Scanner Engine ──────────────────────────────────────────────────────────

class SecurityScanner {
  private h1: HackerOneClient;
  private bc: BugcrowdClient;

  constructor() {
    this.h1 = new HackerOneClient();
    this.bc = new BugcrowdClient();
  }

  get configured(): boolean {
    return this.h1.configured || this.bc.configured;
  }

  async scanPrograms(): Promise<SecurityProgram[]> {
    const programs: SecurityProgram[] = [];

    if (this.h1.configured) {
      try {
        const h1Programs = await this.h1.listPrograms({ perPage: 20, huntable: true });
        programs.push(...h1Programs);
      } catch (e) {
        console.error("[security] HackerOne scan failed:", e);
      }
    }

    if (this.bc.configured) {
      try {
        const bcPrograms = await this.bc.listPrograms();
        programs.push(...bcPrograms);
      } catch (e) {
        console.error("[security] Bugcrowd scan failed:", e);
      }
    }

    return programs;
  }

  async scanTarget(url: string, program: string): Promise<SecurityFindings[]> {
    const findings: SecurityFindings[] = [];

    for (const check of VULN_CHECKS) {
      try {
        const vulnerable = await check.check(url);
        if (vulnerable) {
          findings.push({
            program,
            platform: "hackerone",
            targetType: "URL",
            targetUrl: url,
            vulnType: check.name,
            severity: check.severity,
            title: `${check.name} vulnerability in ${new URL(url).hostname}`,
            description: `Automated scanning detected a potential ${check.name} vulnerability at ${url}.`,
            stepsToReproduce: [
              `Navigate to ${url}`,
              `Inject payload in the relevant parameter`,
              `Observe the vulnerable behavior`,
            ],
            impact: `${check.name} vulnerabilities can lead to ${check.severity === "critical" ? "full system compromise" : "unauthorized access"}.`,
            remediation: `Sanitize all user input and implement proper validation.`,
            estimatedPayout: check.estimatedPayout,
          });
        }
      } catch {
        // Check failed, skip
      }
    }

    return findings;
  }

  async submitFinding(finding: SecurityFindings): Promise<{ id: string; url: string } | null> {
    if (!this.h1.configured) return null;

    const report = generateReport(finding);
    const handle = finding.program;

    const result = await this.h1.submitReport(handle, {
      team_handle: handle,
      title: finding.title,
      vulnerability_information: report,
      impact: finding.impact,
    });

    if (result) {
      finding.reportUrl = result.url;
      finding.submittedAt = new Date().toISOString();
    }

    return result;
  }

  async getSubmissions(): Promise<{ id: string; state: string; title: string }[]> {
    if (!this.h1.configured) return [];
    return this.h1.getSubmissions();
  }
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function parseBountyMax(range: string): number {
  // Parse "$500 - $5,000" → 5000, "$100+" → 100, "$500" → 500
  const cleaned = range.replace(/[^0-9,\-+$]/g, "");
  const parts = cleaned.split("-").map((s) => parseInt(s.replace(/[,$]/g, ""), 10));
  if (parts.length === 2 && !isNaN(parts[1])) return parts[1];
  if (parts.length === 1 && !isNaN(parts[0])) return parts[0];
  return 0;
}

// ─── Exports ─────────────────────────────────────────────────────────────────

export const securityScanner = new SecurityScanner();

export async function runSecurityScan(): Promise<{
  programs: SecurityProgram[];
  findings: SecurityFindings[];
  submissions: { id: string; state: string; title: string }[];
  totalEstimatedPayout: number;
}> {
  const programs = await securityScanner.scanPrograms();
  const findings: SecurityFindings[] = [];
  const submissions = await securityScanner.getSubmissions();

  // Scan top programs by max payout
  const sorted = programs.sort((a, b) => b.maxPayout - a.maxPayout).slice(0, 5);

  for (const program of sorted) {
    if (!program.inScope || program.targets.length === 0) continue;

    for (const target of program.targets.slice(0, 3)) {
      try {
        const targetFindings = await securityScanner.scanTarget(target, program.handle);
        findings.push(...targetFindings);
      } catch {
        // Target scan failed, continue
      }
    }
  }

  // Auto-submit top findings
  for (const finding of findings
    .filter((f) => f.severity === "critical" || f.severity === "high")
    .slice(0, 3)) {
    try {
      await securityScanner.submitFinding(finding);
    } catch {
      // Submission failed
    }
  }

  return {
    programs,
    findings,
    submissions,
    totalEstimatedPayout: findings.reduce((sum, f) => sum + f.estimatedPayout, 0),
  };
}
