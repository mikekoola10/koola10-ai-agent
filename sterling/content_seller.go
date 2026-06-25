package sterling

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
)

type ContentGenerator struct {
    deepseekKey string
    wordpressURL string
    wordpressUser string
    wordpressPass string
    ledger Ledger
    vault *VaultClient
}

func NewContentGenerator(ledger Ledger, vault *VaultClient, deepseekKey string) *ContentGenerator {
    return &ContentGenerator{
        deepseekKey:  deepseekKey,
        wordpressURL: os.Getenv("WORDPRESS_URL"),
        wordpressUser: os.Getenv("WORDPRESS_USER"),
        wordpressPass: os.Getenv("WORDPRESS_PASS"),
        ledger:       ledger,
        vault:        vault,
    }
}

// GenerateArticle creates a technical article based on a trending topic
func (cg *ContentGenerator) GenerateArticle(topic string) (string, error) {
    prompt := fmt.Sprintf(`
Write a detailed, 1000‑word technical article about "%s". Target audience: developers and AI enthusiasts.
Include:
- Introduction
- How it works
- Practical example (with code if applicable)
- Pros and cons
- Conclusion with affiliate call‑to‑action (e.g., "Try this tool today")

Output in markdown.
`, topic)
    return callDeepSeek(prompt, cg.deepseekKey)
}

// PublishToWordPress publishes the article to your WordPress site
func (cg *ContentGenerator) PublishToWordPress(title, content string) error {
    if cg.wordpressURL == "" {
        log.Printf("[ContentSeller] WordPress URL not configured, skipping publish.")
        return nil
    }
    post := map[string]interface{}{
        "title":   title,
        "content": content,
        "status":  "publish",
    }
    jsonData, _ := json.Marshal(post)
    req, _ := http.NewRequest("POST", cg.wordpressURL+"/wp-json/wp/v2/posts", bytes.NewReader(jsonData))
    req.SetBasicAuth(cg.wordpressUser, cg.wordpressPass)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("wordpress error %d", resp.StatusCode)
    }
    return nil
}

// PostToMedium publishes to Medium (optional)
func (cg *ContentGenerator) PostToMedium(title, content string, mediumToken string) error {
    // Similar to WordPress, using Medium's API
    return nil
}

// RunDailyContentCreation generates and publishes 2‑3 articles per day
func (cg *ContentGenerator) RunDailyContentCreation() {
    topics := []string{
        "AI agent for bug bounty automation",
        "How to use virtual cards for subscription management",
        "Setting up an autonomous financial ledger with Go",
    }
    for _, topic := range topics {
        content, err := cg.GenerateArticle(topic)
        if err != nil {
            log.Printf("[ContentSeller] Generation failed: %v", err)
            continue
        }
        title := fmt.Sprintf("AI-Driven: %s", topic)
        if err := cg.PublishToWordPress(title, content); err != nil {
            log.Printf("[ContentSeller] WordPress publish failed: %v", err)
            continue
        }
        // Record expected income (e.g., affiliate commissions tracked separately)
        cg.ledger.RecordRevenue(20.0, "Content article: "+topic)
        cg.vault.AddEntry(VaultEntry{
            Description: "Published article",
            Amount:      20.0,
            Type:        "income",
            Notes:       topic,
        })
        log.Printf("[ContentSeller] Generated article: %s", title)

        // Save locally for manual publishing
        filename := fmt.Sprintf("article_%d.md", time.Now().UnixNano())
        os.WriteFile(filename, []byte(content), 0644)
        fmt.Printf("--- ARTICLE GENERATED: %s ---\nSaved to: %s\n\n%s\n---------------------------\n", title, filename, content)

        // Skip sleep for emergency local execution if desired, or use a small delay
        time.Sleep(2 * time.Second)
    }
}
