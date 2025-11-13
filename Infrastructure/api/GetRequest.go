package Infrastructure

import (
	"context"
	"os"
	"fmt"
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"restaurant-finder/Domain/entity"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)
type OpenAIGenerator struct {
	// クライアントインスタンスなどを保持するためのフィールド
	// 例: openaiClient *openai.Client
	// クライアントインスタンスなどを保持するためのフィールド
	// 例: openaiClient *openai.Client
	//gitテスト
}


func (g *OpenAIGenerator) GenerateSearchQuery(prompt string) (*entity.HotPepperRequestParams, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	bucketName := "hotpepper-api-definitions" 
	objectKey := "format.json" 

	// --- 4. S3からファイル（オブジェクト）を取得 ---
	// GetObjectInputに必要な情報を詰めてAPIを呼び出します。
	resp, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		log.Fatalf("failed to get object, %v", err)
	}

	// 関数を抜ける際に必ずBodyをクローズします。
	// これを忘れるとリソースリークの原因になります。

	defer resp.Body.Close() // これで Body は常にクローズされる

formatFile,err := io.ReadAll(resp.Body)

if err != nil {
    // S3オブジェクトの読み取りエラー（ファイルがないなど）
    log.Fatalf("ファイルの内容を読み取れませんでした: %v", err) 
}
var mappingData map[string]interface{}
err = json.Unmarshal(formatFile, &mappingData)
    fmt.Printf("フォーマットファイル: %v\n", mappingData) // fmt.Printf の書式も修正しました
if err != nil {
    // JSON形式が不正な場合のエラー処理
    log.Fatalf("マッピングファイルの内容をJSONとしてパースできませんでした: %v", err)
}
log.Printf("JSONデータがGoのマップへ正常にパースされました。ルートキーの数: %d", len(mappingData))
// ここから mappingData を使って処理を続行
// ...	//openAIAPIキー読み込み
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	//prompt作成
	resp, err:= client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
                    Role:    openai.ChatMessageRoleSystem,
                    Content:"あなたはHotPepper APIのリクエストパラメータを生成するアシスタントです。ユーザーの入力から適切な検索パラメータをJSON形式で返してください。**応答は、先頭から末尾まで純粋なJSONオブジェクトのみで構成してください。いかなるコメント、説明文、前書き、後書き、またはコードブロック記号（```）も厳禁です。**",
                },
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(`以下のHotPepper APIのリクエストパラメータを生成してください。
                        
                        --- マッピング情報ファイルの内容 ---
                        %s
                        --- マッピング情報ファイルの内容 終 ---
                        
                        生成対象のユーザーリクエスト: %s
                        `, mappingData, prompt),
				}
			},
		},
	)
	
	if err !=nil{
		return nil, err
	}
	
	responseContent := resp.Choices[0].Message.Content
	fmt.Printf("OpenAI response: %s\n", responseContent)
	
	// レスポンスからJSON部分のみを抽出する試み
	jsonStart := 0
	jsonEnd := len(responseContent)
	
	// コードブロック内のJSONを探す
	if start := findJSONStart(responseContent); start != -1 {
		jsonStart = start
	}
	if end := findJSONEnd(responseContent); end != -1 {
		jsonEnd = end
	}
	
	jsonContent := responseContent[jsonStart:jsonEnd]
	fmt.Printf("Extracted JSON content: %s\n", jsonContent)
	
	var params entity.HotPepperRequestParams
	if err := json.Unmarshal([]byte(jsonContent), &params); err != nil {
		fmt.Printf("Failed to parse OpenAI response as JSON: %v\n", err)
		fmt.Printf("Raw response: %s\n", responseContent)
		return nil, err
	}
	
	return &params,nil
}

// findJSONStart レスポンスからJSONの開始位置を見つける
func findJSONStart(content string) int {
	// 最初の { を見つける
	for i, char := range content {
		if char == '{' {
			return i
		}
	}
	return -1
}

// findJSONEnd レスポンスからJSONの終了位置を見つける
func findJSONEnd(content string) int {
	// 最後の } を見つける
	for i := len(content) - 1; i >= 0; i-- {
		if content[i] == '}' {
			return i + 1
		}
	}
	return -1
}
