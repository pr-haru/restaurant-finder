package Infrastructure

import (
	"context"
	"os"
	"fmt"
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"restaurant-finder/Domain/entity"
)
type OpenAIGenerator struct {
	// インターフェイス実装時、メソッドの所有者を設定
}

func (g *OpenAIGenerator) GenerateSearchQuery(prompt string) (*entity.HotPepperRequestParams, error) {
	//openAIAPIキー読み込み
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	//prompt作成
	resp, err:= client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			//formatがjsonにならなかったためここで指定
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type:openai.ChatCompletionResponseFormatTypeJSONObject,
			Messages: []openai.ChatCompletionMessage{
				{
                    Role:    openai.ChatMessageRoleSystem,
                    Content:"このシステムは、ユーザーのリクエスト文を解析し、hotpepperAPIに解析結果を送り、結果を受け取ります。
例：					
- 主要な地名とエリアコードの対応:
  - 福岡: large_area: Z019
  - 天神: middle_area: Y066
  - 博多: middle_area: Y065
- 主要なジャンルとコードの対応:
  - 居酒屋: G001
  - 和食: G003
  - 焼肉: G008
- ルール:
  1. ユーザーの入力から、最も具体的で一致する地名（例: 天神、博多）を抽出し、middle_areaを設定します。エリアに合った地域を検索してください。
  2. 料理名（例: もつ鍋）や店名（例: ラーメン）は、keywordに設定します。
  3. ジャンルコードに一致しない料理名や店名は、ジャンルを設定せず、keywordに含めてください。
  4. パラメータの型はentityに定義した型に合わせてください",
    },
                },
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt + "。この文章からHotPepper APIのリクエストパラメータをJSON形式で生成してください。",
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	responseContent := resp.Choices[0].Message.Content
	fmt.Printf("OpenAI response: %s\n", responseContent)
	
	var params entity.HotPepperRequestParams
	if err := json.Unmarshal([]byte(responseContent), &params); err != nil {
		fmt.Printf("Failed to parse OpenAI response as JSON: %v\n", err)
		return nil,err
	}
	
	return &params,nil
}
