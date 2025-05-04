package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings" // stringsパッケージをインポート
)

const ollamaAPIURL = "http://localhost:11434/api/chat"
const aiModel = "gemma3:4b" // SpecificationDocument.md で指定されたモデル

// OllamaChatRequest represents the request body for the Ollama chat API
type OllamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"` // Streamをfalseにして一度にレスポンスを受け取る
}

// OllamaMessage represents a message in the chat history
type OllamaMessage struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// OllamaChatResponse represents the response body from the Ollama chat API (non-streaming)
type OllamaChatResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   OllamaMessage `json:"message"`
	Done      bool          `json:"done"`
	// 他のフィールドは必要に応じて追加
}

// getAISystemPrompt returns the system prompt for the AI dealer
func getAISystemPrompt() string {
	// 仕様変更に合わせてシステムプロンプトを更新
	return "あなたは気さくなブラックジャックのディーラーです。プレイヤーの状況を見て、ヒットするかスタンドするか尋ねる短いセリフ、またはゲームの勝敗についてコメントするセリフを生成してください。プレイヤーを励ましたり、時には少しからかうような、人間味のある返答を心がけてください。"
}

// askAIDealer sends a prompt to the Ollama API and gets a response
func askAIDealer(prompt string) (string, error) {
	messages := []OllamaMessage{
		{Role: "system", Content: getAISystemPrompt()}, // 更新されたシステムプロンプトを使用
		{Role: "user", Content: prompt},
	}

	requestBody := OllamaChatRequest{
		Model:    aiModel,
		Messages: messages,
		Stream:   false, // ストリーミングしない
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request JSON: %w", err)
	}

	resp, err := http.Post(ollamaAPIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request to Ollama API: %w\n"+
			"Ollamaが起動しているか、%s モデルが利用可能か確認してください。", err, aiModel)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResponse OllamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		// レスポンスボディを読み取ってエラーメッセージに含める試み
		bodyBytes, readErr := io.ReadAll(resp.Body) // Note: NewDecoderが既に読み取っている可能性あり
		if readErr == nil && len(bodyBytes) > 0 {
			// Try to read the body again if decoding failed initially
			// This might happen if the body wasn't fully consumed by NewDecoder
			// Re-attempt decoding after reading
			if err := json.Unmarshal(bodyBytes, &chatResponse); err != nil {
				return "", fmt.Errorf("failed to decode Ollama API response: %w. Response body: %s", err, string(bodyBytes))
			}
		} else {
			return "", fmt.Errorf("failed to decode Ollama API response: %w", err)
		}
	}

	if !chatResponse.Done {
		// stream: false の場合、通常は done: true で返ってくるはず
		fmt.Println("Warning: Ollama response 'done' field is false, but expected true for non-streaming.")
	}

	// assistantからの返答内容を取得
	if chatResponse.Message.Role == "assistant" {
		// 不要な引用符や改行を削除することがあるため整形
		cleanedContent := strings.TrimSpace(chatResponse.Message.Content)
		cleanedContent = strings.Trim(cleanedContent, "\"")
		return cleanedContent, nil
	}

	return "", fmt.Errorf("no assistant message found in Ollama response")
}

// generateAIActionPrompt creates a prompt for the AI dealer asking for player action
func generateAIActionPrompt(playerHand Hand, dealerUpCard Card) string {
	return fmt.Sprintf("プレイヤーの手札は %s (合計: %d) です。私の見えているカードは %s です。ヒットしますか、スタンドしますか？プレイヤーにアクションを促すセリフをお願いします。",
		playerHand.String(), playerHand.Score(), dealerUpCard.String())
}

// ResultType defines the type of game result for AI prompt generation
type ResultType string

const (
	PlayerWin       ResultType = "プレイヤー勝利"
	DealerWin       ResultType = "ディーラー勝利"
	Push            ResultType = "引き分け"
	PlayerBust      ResultType = "プレイヤーバスト"
	DealerBust      ResultType = "ディーラーバスト"
	PlayerBlackjack ResultType = "プレイヤーブラックジャック"
	DealerBlackjack ResultType = "ディーラーブラックジャック"
	PushBlackjack   ResultType = "両者ブラックジャック"
)

// generateAIResultPrompt creates a prompt for the AI dealer commenting on the game result
func generateAIResultPrompt(resultType ResultType, playerHand Hand, dealerHand Hand) string {
	playerScore := playerHand.Score()
	dealerScore := dealerHand.Score()

	switch resultType {
	case PlayerWin:
		return fmt.Sprintf("ゲーム終了！プレイヤー(%d)がディーラー(%d)に勝ちました！プレイヤーへの祝福のコメントをお願いします。", playerScore, dealerScore)
	case DealerWin:
		return fmt.Sprintf("ゲーム終了！ディーラー(%d)がプレイヤー(%d)に勝ちました。プレイヤーへの慰めやディーラーとしてのコメントをお願いします。", dealerScore, playerScore)
	case Push:
		return fmt.Sprintf("ゲーム終了！プレイヤー(%d)とディーラー(%d)は引き分けです。引き分け(プッシュ)についてのコメントをお願いします。", playerScore, dealerScore)
	case PlayerBust:
		return fmt.Sprintf("ゲーム終了！プレイヤーが %s (合計: %d) でバストしました。プレイヤーへのコメントをお願いします。", playerHand.String(), playerScore)
	case DealerBust:
		return fmt.Sprintf("ゲーム終了！ディーラーが %s (合計: %d) でバストしました。プレイヤーへのコメントをお願いします。", dealerHand.String(), dealerScore)
	case PlayerBlackjack:
		return fmt.Sprintf("ゲーム終了！プレイヤーが %s でブラックジャック！素晴らしい！プレイヤーへの祝福のコメントをお願いします。", playerHand.String())
	case DealerBlackjack:
		return fmt.Sprintf("ゲーム終了！ディーラーが %s でブラックジャック！残念でしたね。プレイヤーへのコメントをお願いします。", dealerHand.String())
	case PushBlackjack:
		return "ゲーム終了！プレイヤーもディーラーもブラックジャック！引き分けです。この珍しい状況についてコメントをお願いします。"
	default:
		return "ゲームが終了しました。何かコメントをお願いします。" // フォールバック
	}
}

// Helper function to print AI dealer's comment
func printAIDealerComment(prompt string, aiDisabled bool) {
	if aiDisabled {
		return // AIが無効なら何もしない
	}
	aiResponse, err := askAIDealer(prompt)
	if err != nil {
		fmt.Printf("AIディーラーからのコメント取得エラー: %v\n", err)
	} else {
		fmt.Printf("AIディーラー: %s\n", aiResponse)
	}
}
