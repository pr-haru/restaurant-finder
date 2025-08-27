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
	// クライアントインスタンスなどを保持するためのフィールド
	// 例: openaiClient *openai.Client
}

func (g *OpenAIGenerator) GenerateSearchQuery(prompt string) (*entity.HotPepperRequestParams, error) {
	//openAIAPIキー読み込み
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	//prompt作成
	resp, err:= client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
                    Role:    openai.ChatMessageRoleSystem,
                    Content:"このシステムはユーザーからのリクエスト文をOPENAPAPIで解析し、解析結果をhotpepperAPIを使って検索結果として表示します。",
                },
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt + "。この文章からHotPepper APIのリクエストパラメータをJSON形式で生成してください。",
				},
			},
		},
	)
	if err !=nil{
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
