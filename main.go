package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("###################################")
	fmt.Println("# AIと対話しながら進めるブラックジャック #")
	fmt.Println("###################################")
	fmt.Println("Commjackへようこそ！")

	// AI機能を使うか確認
	var aiDisabled bool
	for {
		fmt.Print("AIディーラー機能を使用しますか？ (yes/no): ")
		input, _ := reader.ReadString('\n')
		answer := strings.TrimSpace(strings.ToLower(input))
		if answer == "yes" || answer == "y" {
			aiDisabled = false
			fmt.Println("AIディーラーがあなたに話しかけます。接続を確認します...")
			// Ollamaサーバー接続確認（簡易）
			_, err := askAIDealer("接続テスト") // 接続テスト用の簡単なプロンプト
			if err != nil {
				fmt.Printf("警告: AIディーラーとの接続に失敗しました: %v\n", err)
				fmt.Println("AI機能なしでゲームを開始します。")
				aiDisabled = true // エラーならAI無効にする
			} else {
				fmt.Println("AIディーラーとの接続を確認しました。")
			}
			break
		} else if answer == "no" || answer == "n" {
			aiDisabled = true
			fmt.Println("AI機能なしでゲームを開始します。")
			break
		} else {
			fmt.Println("無効な入力です。'yes' または 'no' で答えてください。")
		}
	}

	for { // ゲームループ (繰り返しプレイ)
		game := NewGame(aiDisabled)
		fmt.Println("\n--- 新しいゲームを開始します ---")
		fmt.Println("カードを配ります...")

		err := game.DealInitialHands()
		if err != nil {
			fmt.Printf("ゲームの初期化に失敗しました: %v\n", err)
			return
		}

		// 最初に手札を表示 (ディーラーの2枚目は隠す)
		game.PrintHands(true)

		// ブラックジャック判定
		playerHasBJ := game.PlayerHand.IsBlackjack()
		dealerHasBJ := game.DealerHand.IsBlackjack()
		initialResultDetermined := false // 初期BJで結果が決まったか

		if playerHasBJ {
			if dealerHasBJ {
				fmt.Println("ディーラーもブラックジャック！ プッシュ！")
				game.PrintHands(false) // 両方の手札を公開
				// --- AIコメント追加 ---
				prompt := generateAIResultPrompt(PushBlackjack, game.PlayerHand, game.DealerHand)
				printAIDealerComment(prompt, game.AIDisabled)
				// --- AIコメント追加完了 ---
				game.GameState = GameOver
				initialResultDetermined = true
			} else {
				fmt.Println("ブラックジャック！ あなたの勝ちです！")
				game.PrintHands(false) // 両方の手札を公開
				// --- AIコメント追加 ---
				prompt := generateAIResultPrompt(PlayerBlackjack, game.PlayerHand, game.DealerHand)
				printAIDealerComment(prompt, game.AIDisabled)
				// --- AIコメント追加完了 ---
				game.GameState = GameOver
				initialResultDetermined = true
			}
		} else if dealerHasBJ {
			// プレイヤーがBJでない場合のみディーラーのBJを公開
			fmt.Println("ディーラーがブラックジャック！ あなたの負けです。")
			game.PrintHands(false) // ディーラーの手札を公開
			// --- AIコメント追加 ---
			prompt := generateAIResultPrompt(DealerBlackjack, game.PlayerHand, game.DealerHand)
			printAIDealerComment(prompt, game.AIDisabled)
			// --- AIコメント追加完了 ---
			game.GameState = GameOver
			initialResultDetermined = true
		}

		// プレイヤーのターン (初期BJでゲームが終わっていない場合)
		if game.GameState == PlayerTurn {
			for game.GameState == PlayerTurn {
				err := game.PlayerAction(reader)
				if err != nil {
					fmt.Printf("エラーが発生しました: %v\n", err)
					game.GameState = GameOver // エラー時はゲーム終了
				}
			}
		}

		// ディーラーのターン (プレイヤーがバストしておらず、初期BJでもない場合)
		if game.GameState == DealerTurn {
			err := game.DealerAction()
			if err != nil {
				fmt.Printf("エラーが発生しました: %v\n", err)
				game.GameState = GameOver // エラー時はゲーム終了
			}
		}

		// 勝敗判定 (ゲームが終了状態 かつ 初期BJで結果が決まっていない場合)
		if game.GameState == GameOver && !initialResultDetermined {
			// バストの場合はすでにメッセージとAIコメントが表示されている
			if !game.PlayerHand.IsBust() && !game.DealerHand.IsBust() {
				game.DetermineWinner() // DetermineWinner内でメッセージとAIコメント表示
			}
		}

		// もう一度プレイするか確認
		fmt.Print("\nもう一度プレイしますか？ (yes/no): ")
		input, _ := reader.ReadString('\n')
		answer := strings.TrimSpace(strings.ToLower(input))
		if answer != "yes" && answer != "y" {
			break // ループを抜けて終了
		}
	}

	fmt.Println("ゲームを終了します。ありがとうございました！")
}
