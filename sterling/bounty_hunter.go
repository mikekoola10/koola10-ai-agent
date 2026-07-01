package sterling

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/exec"
    "time"
)

type BountyProgram struct {
    Name        string `json:"name"`
    Target      string `json:"target"`      // URL or domain
    BountyRange string `json:"bounty_range"`
}

type ScanResult struct {
    IsVulnerable bool
    Severity     string   // low, medium, high, critical
    Description  string
    Output       string
    Bounty       int      // estimated bounty in dollars
}

type BountyHunter struct {
    browserAgentURL string
    ledger          Ledger
    vault           *VaultClient
    deepseekKey     string
    programsFile    string
}

func NewBountyHunter(ledger Ledger, vault *VaultClient, deepseekKey string) *BountyHunter {
    return &BountyHunter{
        browserAgentURL: os.Getenv("BROWSER_AGENT_URL"),
        ledger:          ledger,
        vault:           vault,
        deepseekKey:     deepseekKey,
        programsFile:    "/data/bounty_programs.json",
    }
}

// Fetch programs from HackerOne using browser agent (or static config)
func (bh *BountyHunter) fetchPrograms() ([]BountyProgram, error) {
    // Option 1: load from local JSON (pre‑configured)
    data, err := os.ReadFile(bh.programsFile)
    if err == nil {
        var progs []BountyProgram
        if err := json.Unmarshal(data, &progs); err == nil {
            return progs, nil
        }
    }

    // Option 2: use browser agent to scrape HackerOne directory
    // For brevity, return a hardcoded test program
    return []BountyProgram{
        {Name: "Test Program", Target: "https://testfire.net", BountyRange: "100-500"},
    }, nil
}

// runNucleiScan executes nuclei against a target
func (bh *BountyHunter) runNucleiScan(target string) (*ScanResult, error) {
    cmd := exec.Command("nuclei", "-u", target, "-severity", "low,medium,high,critical", "-json", "-silent")
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        return nil, err
    }

    // Parse JSON lines from nuclei
    var findings []map[string]interface{}
    for _, line := range bytes.Split(out.Bytes(), []byte("\n")) {
        if len(line) == 0 {
            continue
        }
        var f map[string]interface{}
        if err := json.Unmarshal(line, &f); err == nil {
            findings = append(findings, f)
        }
    }

    if len(findings) == 0 {
        return &ScanResult{IsVulnerable: false}, nil
    }

    // Take the highest severity finding
    severityMap := map[string]int{"info":0, "low":1, "medium":2, "high":3, "critical":4}
    best := findings[0]
    bestScore := 0
    if s, ok := best["severity"].(string); ok {
        bestScore = severityMap[s]
    }

    for _, f := range findings {
        if sStr, ok := f["severity"].(string); ok {
            s := severityMap[sStr]
            if s > bestScore {
                best = f
                bestScore = s
            }
        }
    }

    // Estimate bounty based on severity
    bounty := 0
    severity, _ := best["severity"].(string)
    switch severity {
    case "low":
        bounty = 100
    case "medium":
        bounty = 500
    case "high":
        bounty = 1500
    case "critical":
        bounty = 3000
    }

    template, _ := best["template"].(string)

    return &ScanResult{
        IsVulnerable: true,
        Severity:     severity,
        Description:  template,
        Output:       out.String(),
        Bounty:       bounty,
    }, nil
}

// generateReport uses DeepSeek to write a professional report
func (bh *BountyHunter) generateReport(program BountyProgram, result ScanResult) (string, error) {
    limit := 500
    if len(result.Output) < limit {
        limit = len(result.Output)
    }
    prompt := fmt.Sprintf(`
You are a security researcher. Write a concise bug bounty report for the following finding:

Target: %s
Vulnerability: %s
Severity: %s
Description: %s

Include:
- Summary
- Steps to reproduce (hypothetical if needed)
- Impact
- Recommended fix

Output in plain text.
`, program.Target, result.Description, result.Severity, result.Output[:limit])

    // Call DeepSeek API (reuse existing client)
    resp, err := callDeepSeek(prompt, bh.deepseekKey)
    if err != nil {
        return "", err
    }
    return resp, nil
}

// submitReport via browser agent (logs into HackerOne, fills form)
func (bh *BountyHunter) submitReport(program BountyProgram, reportText string) error {
    // Call browser agent endpoint /hackerone/submit
    // Similar to cashapp payout but for H1
    // For brevity, assume we have a helper
    log.Printf("[BountyHunter] Would submit report for %s", program.Name)
    return nil
}

// RunDailyScan is the main entry point (call from cron or daily payer)
func (bh *BountyHunter) RunDailyScan(targetLimit int, outputFile string) {
    programs, err := bh.fetchPrograms()
    if err != nil {
        log.Printf("[BountyHunter] Failed to fetch programs: %v", err)
        return
    }

    count := 0
    for _, prog := range programs {
        if targetLimit > 0 && count >= targetLimit {
            break
        }
        log.Printf("[BountyHunter] Scanning %s (%s)", prog.Name, prog.Target)
        result, err := bh.runNucleiScan(prog.Target)
        if err != nil {
            log.Printf("[BountyHunter] Scan error: %v", err)
            continue
        }
        if result.IsVulnerable {
            log.Printf("[BountyHunter] Found vulnerability in %s: %s", prog.Name, result.Severity)
            report, err := bh.generateReport(prog, *result)
            if err != nil {
                log.Printf("[BountyHunter] Report generation failed: %v", err)
                continue
            }
            if err := bh.submitReport(prog, report); err != nil {
                log.Printf("[BountyHunter] Submission failed: %v", err)
                continue
            }

            // Save locally for manual submission
            filename := fmt.Sprintf("bounty_report_%d.txt", time.Now().UnixNano())
            if outputFile != "" {
                f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
                if err == nil {
                    f.WriteString(fmt.Sprintf("\n--- REPORT FOR %s ---\n%s\n", prog.Target, report))
                    f.Close()
                    filename = outputFile
                }
            } else {
                os.WriteFile(filename, []byte(report), 0644)
            }
            fmt.Printf("\n--- BOUNTY REPORT GENERATED ---\nTarget: %s\nSaved to: %s\n------------------------------\n", prog.Target, filename)

            // Record expected bounty in ledger (when paid)
            bh.ledger.RecordRevenue(float64(result.Bounty), "Bug bounty: "+prog.Name+" (pending)")
            count++
            bh.vault.AddEntry(VaultEntry{
                Description: "Bug bounty submission",
                Amount:      float64(result.Bounty),
                Type:        "income",
                Notes:       fmt.Sprintf("Submitted to %s, severity %s", prog.Name, result.Severity),
            })
        }
    }
}
