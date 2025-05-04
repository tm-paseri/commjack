package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

// GameState represents the state of the game
type GameState int

const (
	PlayerTurn GameState = iota
	DealerTurn
	GameOver
)

// Game holds the game state
type Game struct {
	Deck       Deck
	PlayerHand Hand
	DealerHand Hand
	GameState  GameState
	AIDisabled bool // AI機能を無効にするフラグ
}

// NewGame creates a new game instance
func NewGame(aiDisabled bool) *Game {
	deck := NewDeck()
	deck.Shuffle()
	return &Game{
		Deck:       deck,
		PlayerHand: Hand{},
		DealerHand: Hand{},
		GameState:  PlayerTurn,
		AIDisabled: aiDisabled,
	}
}

// DealInitialHands deals the initial two cards to player and dealer
func (g *Game) DealInitialHands() error {
	for i := 0; i < 2; i++ {
		card, err := g.Deck.Draw()
		if err != nil {
			return fmt.Errorf("failed to draw card for player: %w", err)
		}
		g.PlayerHand.AddCard(card)

		card, err = g.Deck.Draw()
		if err != nil {
			return fmt.Errorf("failed to draw card for dealer: %w", err)
		}
		g.DealerHand.AddCard(card)
	}
	return nil
}

// PrintHands prints the current hands (hiding dealer's second card initially)
func (g *Game) PrintHands(hideDealerCard bool) {
	fmt.Println("--------------------")
	fmt.Printf("ディーラーの手札: ")
	if hideDealerCard && len(g.DealerHand) > 1 {
		fmt.Printf("%s [伏せられたカード]\n", g.DealerHand[0].String())
	} else {
		fmt.Printf("%s (合計: %d)\n", g.DealerHand.String(), g.DealerHand.Score())
	}

	fmt.Printf("あなたの手札: %s (合計: %d)\n", g.PlayerHand.String(), g.PlayerHand.Score())
	fmt.Println("--------------------")
}

// PlayerAction handles the player's turn (hit or stand)
func (g *Game) PlayerAction(reader *bufio.Reader) error {
	// AIディーラーの声かけ (アクションを促す)
	if !g.AIDisabled && len(g.DealerHand) > 0 {
		// generateAIActionPrompt を使用
		prompt := generateAIActionPrompt(g.PlayerHand, g.DealerHand[0])
		aiResponse, err := askAIDealer(prompt)
		if err != nil {
			fmt.Printf("AIディーラーからの応答取得エラー: %v\n", err)
			fmt.Println("ヒット(h) or スタンド(s)?") // エラー時は通常のプロンプト
		} else {
			fmt.Printf("AIディーラー: %s\n", aiResponse)
			fmt.Print("あなたの選択 (ヒット[h] / スタンド[s]): ") // AIのセリフの後に選択肢を提示
		}
	} else {
		fmt.Print("ヒット(h) or スタンド(s)? ") // AI無効時または初回Deal直後など
	}

	input, _ := reader.ReadString('\n')
	action := strings.TrimSpace(strings.ToLower(input))

	switch action {
	case "h", "hit":
		card, err := g.Deck.Draw()
		if err != nil {
			return fmt.Errorf("ヒット中にエラー発生: %w", err)
		}
		g.PlayerHand.AddCard(card)
		fmt.Printf("ヒット！ 引いたカード: %s\n", card.String())
		g.PrintHands(true) // ディーラーのカードはまだ隠す

		if g.PlayerHand.IsBust() {
			fmt.Println("バスト！ あなたの負けです。")
			// --- AIコメント追加 ---
			prompt := generateAIResultPrompt(PlayerBust, g.PlayerHand, g.DealerHand)
			printAIDealerComment(prompt, g.AIDisabled)
			// --- AIコメント追加完了 ---
			g.GameState = GameOver
		}
		// バストしていなければプレイヤーのターン継続
	case "s", "stand":
		fmt.Println("スタンドしました。ディーラーのターンに移ります。")
		g.GameState = DealerTurn
	default:
		fmt.Println("無効な入力です。[h]か[s]を入力してください。")
		// 状態は変えずに再度入力を促す (ループはmain側で制御)
	}
	return nil
}

// DealerAction handles the dealer's turn
func (g *Game) DealerAction() error {
	fmt.Println("\nディーラーのターン")
	g.PrintHands(false) // ディーラーのカードを公開

	// ディーラーの手札が17未満の場合、ヒットし続ける
	for g.DealerHand.Score() < 17 {
		fmt.Println("ディーラーはヒットします。")
		time.Sleep(1 * time.Second) // 演出のための待機
		card, err := g.Deck.Draw()
		if err != nil {
			return fmt.Errorf("ディーラーのヒット中にエラー発生: %w", err)
		}
		g.DealerHand.AddCard(card)
		fmt.Printf("ディーラーが引いたカード: %s\n", card.String())
		g.PrintHands(false)

		if g.DealerHand.IsBust() {
			fmt.Println("ディーラーバスト！ あなたの勝ちです！")
			// --- AIコメント追加 ---
			prompt := generateAIResultPrompt(DealerBust, g.PlayerHand, g.DealerHand)
			printAIDealerComment(prompt, g.AIDisabled)
			// --- AIコメント追加完了 ---
			g.GameState = GameOver
			return nil // ディーラーがバストしたら即終了
		}
		time.Sleep(1 * time.Second) // 演出のための待機
	}

	if g.GameState != GameOver { // ディーラーがバストしていなければ結果判定へ
		fmt.Printf("ディーラーは %d でスタンドします。\n", g.DealerHand.Score())
		g.GameState = GameOver // ディーラーのアクション終了後、ゲーム終了状態へ
	}
	return nil
}

// DetermineWinner determines the winner after both turns are complete
func (g *Game) DetermineWinner() {
	playerScore := g.PlayerHand.Score()
	dealerScore := g.DealerHand.Score()
	var resultType ResultType
	resultMessage := ""

	fmt.Println("\n--- 結果 ---")
	g.PrintHands(false) // 最終結果表示

	// 勝敗判定ロジック (バストは既に処理されている前提)
	if playerScore > 21 {
		resultType = PlayerBust // ここには通常到達しないはずだが念のため
		resultMessage = "あなたはバストしています。ディーラーの勝ちです。"
	} else if dealerScore > 21 {
		resultType = DealerBust // ここにも通常到達しないはず
		resultMessage = "ディーラーがバストしました。あなたの勝ちです！"
	} else if playerScore == dealerScore {
		resultType = Push
		resultMessage = "プッシュ！ 引き分けです。"
	} else if playerScore > dealerScore {
		resultType = PlayerWin
		resultMessage = "あなたの勝ちです！"
	} else {
		resultType = DealerWin
		resultMessage = "ディーラーの勝ちです。"
	}

	fmt.Println(resultMessage)
	// --- AIコメント追加 ---
	prompt := generateAIResultPrompt(resultType, g.PlayerHand, g.DealerHand)
	printAIDealerComment(prompt, g.AIDisabled)
	// --- AIコメント追加完了 ---

	fmt.Println("----------")
}
