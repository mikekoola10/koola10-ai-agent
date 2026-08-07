/* Live verification of the deployed Nova static showcase site. */
const { chromium } = require("playwright");

const URL = "http://127.0.0.1:3999/";
const results = [];
const log = (k, v) => { results.push(`[${k}] ${v}`); console.log(`[${k}] ${v}`); };

(async () => {
  const browser = await chromium.launch({
    headless: true,
    executablePath: "/usr/bin/google-chrome",
    args: ["--no-sandbox", "--disable-dev-shm-usage"],
  });
  const page = await browser.newPage({ viewport: { width: 1280, height: 900 } });
  const consoleErrors = [];
  page.on("console", (m) => { if (m.type() === "error") consoleErrors.push(m.text()); });
  page.on("pageerror", (e) => consoleErrors.push("PAGEERROR: " + e.message));

  // 1. Load
  await page.goto(URL, { waitUntil: "networkidle", timeout: 45000 });
  await page.waitForTimeout(1200);

  log("title", await page.title());
  const hero = await page.locator("#hero-view h1").textContent().catch(() => "(missing)");
  log("hero-heading", hero.trim());

  // 2. Static-mode signals (topbar pills)
  const brainPill = await page.locator("#brain-pill").textContent().catch(() => "");
  const mockPill = await page.locator("#mock-pill").textContent().catch(() => "");
  const toolsChip = await page.locator("#tools-chip").textContent().catch(() => "");
  log("topbar", `brain="${brainPill.trim()}" mock="${mockPill.trim()}" tools="${toolsChip.trim()}"`);

  // 3. Sample pills present?
  const pillCount = await page.locator(".pill-btn").count();
  const pillLabels = await page.locator(".pill-btn").allTextContents();
  log("sample-pills", `${pillCount} found: ${pillLabels.map((s) => s.trim()).slice(0, 5).join(" | ")}`);

  // 4. Run a sample task → static demo
  const target = page.locator(".pill-btn", { hasText: "Research & report" }).first();
  if (await target.count()) {
    await target.click();
    log("pill-clicked", "Research & report");
  } else {
    // fall back to typing a task
    await page.fill("#task-input", "Research the top 3 AI agent frameworks and compare them");
    log("pill-clicked", "(no pill) typed task manually");
  }
  const sendBtn = page.locator("#send-btn");
  if (await sendBtn.isEnabled().catch(() => false)) {
    await sendBtn.click();
    log("send-clicked", "true");
  }

  // 5. Watch demo steps appear
  await page.waitForTimeout(1200);
  log("task-title", (await page.locator("#task-title").textContent().catch(() => "(n/a)")).trim());
  const stepsDuring = await page.locator("#task-steps .step").count();
  const stepNames = await page.locator("#task-steps .tool-chip").allTextContents().catch(() => []);
  log("steps-mid-run", `${stepsDuring} step cards; tools: ${[...new Set(stepNames)].join(", ")}`);

  // 6. Wait for completion + final report
  await page.waitForTimeout(4000);
  const stepsFinal = await page.locator("#task-steps .step").count();
  const reportVisible = await page.locator("#task-report").isVisible().catch(() => false);
  const reportText = reportVisible ? (await page.locator("#task-report").textContent()).slice(0, 160) : "";
  const status = await page.locator("#task-status").textContent().catch(() => "");
  const stats = await page.locator("#task-stats").textContent().catch(() => "");
  log("final", `steps=${stepsFinal} report=${reportVisible} status="${status.trim()}" stats="${stats.trim()}"`);
  if (reportText) log("report-preview", reportText.replace(/\s+/g, " ").trim().slice(0, 140));
  await page.screenshot({ path: "/tmp/nova-prod-task.png", fullPage: false });

  // 7. Connectors view
  await page.locator('.nav-item[data-view="plugins"]').click();
  await page.waitForTimeout(800);
  const connCards = await page.locator("#conn-grid .conn-card").count();
  const badges = await page.locator("#conn-grid .conn-badge").allTextContents().catch(() => []);
  const connNames = await page.locator("#conn-grid b").allTextContents().catch(() => []);
  log("connectors", `${connCards} cards; badges: ${[...new Set(badges.map((s) => s.trim()))].join(",")}; names: ${connNames.join(", ")}`);
  await page.screenshot({ path: "/tmp/nova-prod-connectors.png", fullPage: false });

  // 8. Hero back
  await page.locator('.nav-item[data-view="agent"]').click();
  await page.waitForTimeout(400);
  const heroVisible = await page.locator("#hero-view").isVisible().catch(() => false);
  log("back-to-hero", String(heroVisible));

  log("console-errors", consoleErrors.length ? consoleErrors.slice(0, 5).join(" || ") : "NONE");
  await browser.close();
})().catch((e) => { console.log("[FATAL]", e.message); process.exit(1); });
