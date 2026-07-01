package sterling

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type DeepSeekRequest struct {
    Model    string `json:"model"`
    Messages []struct {
        Role    string `json:"role"`
        Content string `json:"content"`
    } `json:"messages"`
}

type DeepSeekResponse struct {
    Choices []struct {
        Message struct {
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}

func callDeepSeek(prompt, apiKey string) (string, error) {
    reqBody := DeepSeekRequest{
        Model: "deepseek-chat",
        Messages: []struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        }{
            {Role: "user", Content: prompt},
        },
    }
    jsonBody, _ := json.Marshal(reqBody)
    req, _ := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewReader(jsonBody))
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("DeepSeek API error: %s", resp.Status)
    }

    var deepResp DeepSeekResponse
    if err := json.NewDecoder(resp.Body).Decode(&deepResp); err != nil {
        return "", err
    }
    if len(deepResp.Choices) == 0 {
        return "", fmt.Errorf("no response from DeepSeek")
    }
    return deepResp.Choices[0].Message.Content, nil
}
