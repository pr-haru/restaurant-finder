package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"restaurant-finder/Domain/entity"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIGenerator は OpenAI API を使用して検索パラメータを抽出します
type OpenAIGenerator struct {
	client *openai.Client
}

// NaturalLanguageResponse は検索結果を自然言語で説明する構造体です
type NaturalLanguageResponse struct {
	Summary      string   `json:"summary"`
	Recommendations []string `json:"recommendations"`
}

// NewOpenAIGenerator は新しい OpenAIGenerator を作成します
func NewOpenAIGenerator(apiKey string) *OpenAIGenerator {
	var client *openai.Client
	if apiKey != "" {
		client = openai.NewClient(apiKey)
	}
	return &OpenAIGenerator{client: client}
}

// GenerateSearchQuery はユーザーのプロンプトを HotPepper 検索パラメータに変換します
func (g *OpenAIGenerator) GenerateSearchQuery(prompt string) (*entity.HotPepperRequestParams, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("プロンプトが空です")
	}

	// API キーがない場合はフォールバック
	if g.client == nil {
		return &entity.HotPepperRequestParams{
			Keyword: prompt,
			Count:   10,
		}, nil
	}

	// OpenAI で構造化パラメータを抽出
	aiOut, err := g.extractEntitiesWithOpenAI(prompt)
	if err != nil {
		fmt.Printf("OpenAI 抽出エラー: %v; フォールバック使用\n", err)
		return &entity.HotPepperRequestParams{
			Keyword: prompt,
			Count:   10,
		}, nil
	}

	// format.json を読み込み
	fmtData, err := loadFormatJSON()
	if err != nil {
		fmt.Printf("警告: format.json を読み込めません: %v\n", err)
	}

	// AI 出力を HotPepperRequestParams に変換（コード解決）
	params, err := mergeAIParamsWithCodes(aiOut, fmtData)
	if err != nil {
		fmt.Printf("警告: パラメータマージエラー: %v\n", err)
		params = &entity.HotPepperRequestParams{Keyword: prompt, Count: 10}
	}

	if params.Count == 0 {
		params.Count = 10
	}

	// デバッグ: マッピング結果を出力
	fmt.Printf("マッピング結果 - LargeArea: %s, MiddleArea: %s, SmallArea: %s, Genre: %s, Budget: %s, Keyword: %s\n",
		params.LargeArea, params.MiddleArea, params.SmallArea, params.Genre, params.Budget, params.Keyword)

	return params, nil
}

// aiOutput は AI モデルが出力する JSON 構造です
type aiOutput struct {
	Location      json.RawMessage `json:"location,omitempty"`
	LargeArea     json.RawMessage `json:"large_area,omitempty"`
	MiddleArea    json.RawMessage `json:"middle_area,omitempty"`
	SmallArea     json.RawMessage `json:"small_area,omitempty"`
	Genre         json.RawMessage `json:"genre,omitempty"`
	Keyword       json.RawMessage `json:"keyword,omitempty"`
	PrivateRoom   json.RawMessage `json:"private_room,omitempty"`
	FreeDrink     json.RawMessage `json:"free_drink,omitempty"`
	FreeFood      json.RawMessage `json:"free_food,omitempty"`
	Budget        json.RawMessage `json:"budget,omitempty"`
	Sake          json.RawMessage `json:"sake,omitempty"`
	Cocktail      json.RawMessage `json:"cocktail,omitempty"`
	Wine          json.RawMessage `json:"wine,omitempty"`
	Midnight      json.RawMessage `json:"midnight,omitempty"`
	PartyCapacity json.RawMessage `json:"party_capacity,omitempty"`
}

// extractEntitiesWithOpenAI は OpenAI API を呼び出して検索パラメータを抽出します
func (g *OpenAIGenerator) extractEntitiesWithOpenAI(prompt string) (*aiOutput, error) {
	systemPrompt := `次の指示に従い、ユーザーの自然文リクエストから検索に使える単語を抽出して、必ず「純粋なJSONオブジェクト」のみを返してください。

出力するフィールドは次の通りです（値は人間が読む語句を返すこと。HotPepperの内部コードは返さないでください）:

【地名関連】
- "location": 地名（例: "渋谷", "新宿", "東京"）
- "large_area": 大区分（都道府県レベル、例: "東京", "大阪"）
- "middle_area": 中区分（市区町村レベル、例: "渋谷区", "新宿区"）
- "small_area": 小区分（駅名や地域名、例: "秋葉原", "表参道"）

【ジャンル・キーワード】
- "genre": ジャンル（例: "居酒屋", "イタリアン", "焼肉", "寿司"）
- "keyword": キーワード（例: "個室", "デート", "女子会", "ランチ"）

【値段・予算】
- "budget": 予算・値段（数値: "5000" または 語句: "安い", "高級", "3000円以下" など）

【個室・設備】
- "private_room": 個室の有無（"あり" / "なし" / "true" / "false" / 1 / 0）
- "free_drink": 飲み放題の有無（"あり" / "なし" / "true" / "false" / 1 / 0）
- "free_food": 食べ放題の有無（"あり" / "なし" / "true" / "false" / 1 / 0）
- "midnight": 深夜営業の有無（"あり" / "なし" / "true" / "false" / 1 / 0）

【酒の種類】
- "sake": 日本酒の有無（"あり" / "なし" / "true" / "false" / 1 / 0）
- "cocktail": カクテルの有無（"あり" / "なし" / "true" / "false" / 1 / 0）
- "wine": ワインの有無（"あり" / "なし" / "true" / "false" / 1 / 0）

【その他】
- "party_capacity": 宴会・パーティー対応（"あり" / "なし" / "true" / "false" / 1 / 0）

ルール:
1. 地名について:
   - 県名・市名などの末尾の接尾辞は比較のために削除してよい（例: "東京都" -> "東京", "札幌市" -> "札幌"）
   - ただし、駅名や地域固有名（例: "秋葉原", "渋谷", "表参道"）はそのまま返してください
   - location, large_area, middle_area, small_area のうち、該当するものを抽出してください

2. ジャンル・キーワードについて:
   - ジャンルは料理の種類や店のタイプを抽出（例: "居酒屋", "イタリアン", "焼肉", "寿司", "ラーメン"）
   - キーワードは検索に使える特徴的な単語を抽出（例: "個室", "デート", "女子会", "ランチ", "ディナー"）

3. 値段・予算について:
   - 数値で表現されている場合はその数値を返す（例: "5000円" -> "5000", "3000円以下" -> "3000"）
   - 語句で表現されている場合はその語句を返す（例: "安い", "高級", "リーズナブル"）
   - HotPepperの内部コード（例: "B008"）は返さないでください

4. 個室・設備・酒の種類について:
   - "個室あり", "個室がある", "個室対応" など → "private_room": "あり"
   - "飲み放題あり", "飲み放題付き" など → "free_drink": "あり"
   - "食べ放題あり", "食べ放題付き" など → "free_food": "あり"
   - "深夜営業", "24時間営業" など → "midnight": "あり"
   - "日本酒あり", "日本酒がある" など → "sake": "あり"
   - "カクテルあり", "カクテルがある" など → "cocktail": "あり"
   - "ワインあり", "ワインがある" など → "wine": "あり"
   - 否定形（"個室なし", "飲み放題なし" など）の場合は "なし" を返す

5. 返却する値について:
   - 短い語句にしてください（例: "秋葉原", "居酒屋", "てんぷら", "あり"）
   - 該当する項目がない場合は、そのフィールドを省略してください

このJSONは後でローカルの "format.json" を参照して name -> code に変換します。
例: "private_room": "あり" -> "format.json" の "private_room" の code に解決します。`
	userPrompt := "抽出対象: " + prompt

	resp, err := g.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   800,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI API エラー: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI からレスポンスがありません")
	}

	responseText := resp.Choices[0].Message.Content
	fmt.Printf("OpenAI レスポンス: %s\n", responseText)

	// レスポンスから JSON を抽出
	jsonStr := extractJSON(responseText)
	if jsonStr == "" {
		return nil, fmt.Errorf("レスポンスに JSON が見つかりません: %s", responseText)
	}

	var out aiOutput
	if err := json.Unmarshal([]byte(jsonStr), &out); err != nil {
		return nil, fmt.Errorf("JSON パース失敗: %w; テキスト=%s", err, jsonStr)
	}

	return &out, nil
}

// extractJSON はテキストから最初の JSON オブジェクト {...} を抽出します
func extractJSON(text string) string {
	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}
	depth := 0
	for i := start; i < len(text); i++ {
		if text[i] == '{' {
			depth++
		} else if text[i] == '}' {
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	return ""
}

// loadFormatJSON は format.json を複数のパスから読み込みます
func loadFormatJSON() (map[string]interface{}, error) {
	paths := []string{
		"format.json",
		"../format.json",
		"../../format.json",
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			var top map[string]interface{}
			if err := json.Unmarshal(data, &top); err == nil {
				if results, ok := top["results"].(map[string]interface{}); ok {
					fmt.Printf("format.json を読み込みました: %s\n", path)
					return results, nil
				}
			}
		}
	}

	exePath, err := os.Executable()
	if err == nil {
		absPath := filepath.Join(filepath.Dir(exePath), "format.json")
		if data, err := os.ReadFile(absPath); err == nil {
			var top map[string]interface{}
			if err := json.Unmarshal(data, &top); err == nil {
				if results, ok := top["results"].(map[string]interface{}); ok {
					fmt.Printf("format.json を読み込みました: %s\n", absPath)
					return results, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("format.json が見つかりません")
}

// extractAIParams は OpenAI で抽出した項目を単一の文字列として取得します
func extractAIParams(ai *aiOutput) map[string]string {
	result := make(map[string]string)

	getString := func(rm json.RawMessage) string {
		if len(rm) == 0 {
			return ""
		}
		var s string
		if json.Unmarshal(rm, &s) == nil {
			return strings.TrimSpace(s)
		}
		var n int
		if json.Unmarshal(rm, &n) == nil {
			return fmt.Sprintf("%d", n)
		}
		return strings.Trim(string(rm), " \"")
	}

	// OpenAIで抽出した項目を単一の文字列として格納（各項目は1つずつ）
	if loc := getString(ai.Location); loc != "" {
		result["location"] = loc
	}
	if la := getString(ai.LargeArea); la != "" {
		result["large_area"] = la
	}
	if ma := getString(ai.MiddleArea); ma != "" {
		result["middle_area"] = ma
	}
	if sa := getString(ai.SmallArea); sa != "" {
		result["small_area"] = sa
	}
	if g := getString(ai.Genre); g != "" {
		result["genre"] = g
	}
	if b := getString(ai.Budget); b != "" {
		result["budget"] = b
	}
	if kw := getString(ai.Keyword); kw != "" {
		result["keyword"] = kw
	}

	return result
}

// extractFormatJSONToSlices は format.json から項目をスライスに変換します
func extractFormatJSONToSlices(fmtData map[string]interface{}) map[string][]map[string]interface{} {
	result := make(map[string][]map[string]interface{})

	if fmtData == nil {
		return result
	}

	// format.jsonから各カテゴリの項目をスライスに格納
	categories := []string{"large_area", "middle_area", "small_area", "genre", "budget"}
	for _, category := range categories {
		if arr, ok := fmtData[category].([]interface{}); ok {
			items := make([]map[string]interface{}, 0)
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok {
					items = append(items, m)
				}
			}
			if len(items) > 0 {
				result[category] = items
			}
		}
	}

	return result
}

// mapAIParamsToCodes は OpenAI で抽出した項目（単一値）と format.json のスライスをfor文でマッピングしてコードに変換します
func mapAIParamsToCodes(aiParams map[string]string, fmtSlices map[string][]map[string]interface{}) map[string]string {
	mappedCodes := make(map[string]string)

	// 地名のマッピング
	if loc, ok := aiParams["location"]; ok && loc != "" {
		large, middle, small := resolveLocationFromSlices(loc, fmtSlices)
		if large != "" {
			mappedCodes["large_area"] = large
		}
		if middle != "" {
			mappedCodes["middle_area"] = middle
		}
		if small != "" {
			mappedCodes["small_area"] = small
		}
	}

	// 大区分のマッピング: format.jsonのスライスをfor文でループしてマッピング
	if largeAreaName, ok := aiParams["large_area"]; ok && largeAreaName != "" {
		if items, ok := fmtSlices["large_area"]; ok {
			for _, item := range items {
				if code := matchNameToCode(item, largeAreaName); code != "" {
					mappedCodes["large_area"] = code
					break
				}
			}
		}
	}

	// 中区分のマッピング: format.jsonのスライスをfor文でループしてマッピング
	if middleAreaName, ok := aiParams["middle_area"]; ok && middleAreaName != "" {
		if items, ok := fmtSlices["middle_area"]; ok {
			for _, item := range items {
				if code := matchNameToCode(item, middleAreaName); code != "" {
					mappedCodes["middle_area"] = code
					break
				}
			}
		}
	}

	// 小区分のマッピング: format.jsonのスライスをfor文でループしてマッピング
	if smallAreaName, ok := aiParams["small_area"]; ok && smallAreaName != "" {
		if items, ok := fmtSlices["small_area"]; ok {
			for _, item := range items {
				if code := matchNameToCode(item, smallAreaName); code != "" {
					mappedCodes["small_area"] = code
					break
				}
			}
		}
	}

	// ジャンルのマッピング: format.jsonのスライスをfor文でループしてマッピング
	if genreName, ok := aiParams["genre"]; ok && genreName != "" {
		if items, ok := fmtSlices["genre"]; ok {
			for _, item := range items {
				if code := matchNameToCode(item, genreName); code != "" {
					mappedCodes["genre"] = code
					break
				}
			}
		}
	}

	// 予算のマッピング: format.jsonのスライスをfor文でループしてマッピング
	if budgetValue, ok := aiParams["budget"]; ok && budgetValue != "" {
		// 既に HotPepper コード形式 (例: B008) の場合はそのまま
		if matched, _ := regexp.MatchString(`^[A-Z]\d{3}`, budgetValue); matched {
			mappedCodes["budget"] = budgetValue
		} else {
			// 数値（例: "5000"）ならレンジにマッチさせてコードを探す
			if nmatched, _ := regexp.MatchString(`^\d+$`, budgetValue); nmatched {
				var val int
				fmt.Sscanf(budgetValue, "%d", &val)
				if items, ok := fmtSlices["budget"]; ok {
					for _, item := range items {
						if code := matchNumericBudget(item, val); code != "" {
							mappedCodes["budget"] = code
							break
						}
					}
				}
			} else {
				// 文字列の場合は名前でマッチ
				if items, ok := fmtSlices["budget"]; ok {
					for _, item := range items {
						if code := matchNameToCode(item, budgetValue); code != "" {
							mappedCodes["budget"] = code
							break
						}
					}
				}
			}
		}
	}

	return mappedCodes
}

// matchNameToCode は format.json のオブジェクトと抽出項目名をマッチングしてコードを返します
func matchNameToCode(item map[string]interface{}, extractedName string) string {
	nameLower := normalizeName(extractedName)

	// 直接のnameフィールドをチェック（完全一致を優先）
	if nm, ok := item["name"].(string); ok {
		nmLower := normalizeName(nm)
		// 完全一致を優先
		if nmLower == nameLower {
			if code, ok := item["code"].(string); ok {
				return code
			}
		}
		// 部分一致もチェック
		if strings.Contains(nmLower, nameLower) || strings.Contains(nameLower, nmLower) {
			if code, ok := item["code"].(string); ok {
				return code
			}
		}
	}

	// ネストされた構造をチェック
	for _, v := range item {
		sub, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if nm, ok := sub["name"].(string); ok {
			nmLower := normalizeName(nm)
			// 完全一致を優先
			if nmLower == nameLower {
				if code, ok := sub["code"].(string); ok {
					return code
				}
			}
			// 部分一致もチェック
			if strings.Contains(nmLower, nameLower) || strings.Contains(nameLower, nmLower) {
				if code, ok := sub["code"].(string); ok {
					return code
				}
			}
		}
	}

	return ""
}

// matchNumericBudget は数値予算を format.json のオブジェクトとマッチングしてコードを返します
func matchNumericBudget(item map[string]interface{}, amount int) string {
	name, _ := item["name"].(string)
	// 数字を抽出する（カンマ区切りも考慮）
	re := regexp.MustCompile(`(\d+)`)
	nums := re.FindAllString(name, -1)
	if len(nums) == 0 {
		return ""
	}
	if len(nums) == 1 {
		// 例: "～500円" または "5000円以上" のような表現
		var v int
		fmt.Sscanf(nums[0], "%d", &v)
		// "～" や "以下" を含む場合は上限
		if strings.Contains(name, "～") || strings.HasPrefix(name, "～") || strings.Contains(name, "以下") {
			if amount <= v {
				if code, ok := item["code"].(string); ok {
					return code
				}
			}
		} else if strings.Contains(name, "以上") {
			// "以上" を含む場合は下限
			if amount >= v {
				if code, ok := item["code"].(string); ok {
					return code
				}
			}
		} else {
			// 単一数字は範囲の上限とみなす（例: "5000円" は "～5000円" と同等）
			if amount <= v {
				if code, ok := item["code"].(string); ok {
					return code
				}
			}
		}
	} else if len(nums) >= 2 {
		// 範囲指定（例: "3000円～5000円"）
		var low, high int
		fmt.Sscanf(nums[0], "%d", &low)
		fmt.Sscanf(nums[1], "%d", &high)
		// 範囲内かチェック（境界値も含む）
		if amount >= low && amount <= high {
			if code, ok := item["code"].(string); ok {
				return code
			}
		}
	}
	return ""
}


// resolveLocationFromSlices はスライスからロケーション名を大区分/中区分/小区分のコードに解決します
func resolveLocationFromSlices(name string, fmtSlices map[string][]map[string]interface{}) (largeCode, middleCode, smallCode string) {
	nameLower := normalizeName(name)

	// まず小区分から検索（駅名や地域名は小区分に含まれることが多い）
	if items, ok := fmtSlices["small_area"]; ok {
		for _, item := range items {
			if nm, ok := item["name"].(string); ok {
				nmLower := normalizeName(nm)
				// 完全一致または部分一致をチェック
				if nmLower == nameLower || strings.Contains(nmLower, nameLower) || strings.Contains(nameLower, nmLower) {
					if code, ok := item["code"].(string); ok {
						smallCode = code
					}
					// 小区分から中区分と大区分を取得
					if ma, ok := item["middle_area"].(map[string]interface{}); ok {
						if mc, ok := ma["code"].(string); ok {
							middleCode = mc
						}
						if la, ok := ma["large_area"].(map[string]interface{}); ok {
							if lc, ok := la["code"].(string); ok {
								largeCode = lc
							}
						}
					}
					// 小区分が見つかったら返す
					if smallCode != "" {
						return
					}
				}
			}
		}
	}

	// 中区分から検索
	if items, ok := fmtSlices["middle_area"]; ok {
		for _, item := range items {
			if nm, ok := item["name"].(string); ok {
				nmLower := normalizeName(nm)
				if nmLower == nameLower || strings.Contains(nmLower, nameLower) || strings.Contains(nameLower, nmLower) {
					if code, ok := item["code"].(string); ok {
						middleCode = code
					}
					if la, ok := item["large_area"].(map[string]interface{}); ok {
						if lc, ok := la["code"].(string); ok {
							largeCode = lc
						}
					}
					return
				}
			}
		}
	}

	// 大区分から検索
	if items, ok := fmtSlices["large_area"]; ok {
		for _, item := range items {
			if nm, ok := item["name"].(string); ok {
				nmLower := normalizeName(nm)
				if nmLower == nameLower || strings.Contains(nmLower, nameLower) || strings.Contains(nameLower, nmLower) {
					if code, ok := item["code"].(string); ok {
						largeCode = code
						return
					}
				}
			}
		}
	}

	return "", "", ""
}


// GenerateNaturalLanguageResponse は検索結果を自然言語で説明します
func (g *OpenAIGenerator) GenerateNaturalLanguageResponse(userQuery string, shops []entity.Shop, params *entity.HotPepperRequestParams) (string, error) {
	if g.client == nil || len(shops) == 0 {
		return "", fmt.Errorf("OpenAIクライアントが初期化されていないか、検索結果がありません")
	}

	// 検索結果の要約を作成
	shopSummaries := make([]string, 0)
	for i, shop := range shops {
		if i >= 5 { // 最初の5件のみ
			break
		}
		summary := fmt.Sprintf("- %s（%s、予算: %s、%s）", 
			shop.Name, 
			shop.Genre.Name, 
			shop.Budget.Name,
			shop.Address)
		shopSummaries = append(shopSummaries, summary)
	}

	// 検索パラメータの説明を作成
	paramDesc := "検索条件: "
	if params.Genre != "" {
		paramDesc += fmt.Sprintf("ジャンル指定あり、")
	}
	if params.Budget != "" {
		paramDesc += fmt.Sprintf("予算指定あり、")
	}
	if params.LargeArea != "" || params.MiddleArea != "" || params.SmallArea != "" {
		paramDesc += "エリア指定あり、"
	}
	if params.Keyword != "" {
		paramDesc += fmt.Sprintf("キーワード: %s、", params.Keyword)
	}
	paramDesc = strings.TrimSuffix(paramDesc, "、")

	systemPrompt := `あなたはレストラン検索のアシスタントです。ユーザーの検索クエリと検索結果を基に、自然で親しみやすい日本語で検索結果を説明してください。
検索結果の説明には以下を含めてください：
1. ユーザーの検索意図を理解した上での簡潔な説明
2. 見つかった店舗の特徴やおすすめポイント
3. 検索条件に合致した理由

返答は2-3文程度の簡潔なものにしてください。`

	userPrompt := fmt.Sprintf(`ユーザーの検索クエリ: "%s"
%s
見つかった店舗（%d件）:
%s

上記の検索結果を、ユーザーに分かりやすく自然な日本語で説明してください。`, 
		userQuery,
		paramDesc,
		len(shops),
		strings.Join(shopSummaries, "\n"))

	resp, err := g.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   500,
		Temperature: 0.7,
	})
	if err != nil {
		return "", fmt.Errorf("OpenAI API エラー: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI からレスポンスがありません")
	}

	return resp.Choices[0].Message.Content, nil
}

// mergeAIParamsWithCodes は AI 出力を HotPepperRequestParams に変換し、format.json でコードを解決します
func mergeAIParamsWithCodes(ai *aiOutput, fmtData map[string]interface{}) (*entity.HotPepperRequestParams, error) {
	params := &entity.HotPepperRequestParams{}

	getFlag := func(rm json.RawMessage) int {
		if len(rm) == 0 {
			return 0
		}
		var n int
		if json.Unmarshal(rm, &n) == nil {
			return n
		}
		var s string
		if json.Unmarshal(rm, &s) == nil {
			s = strings.TrimSpace(s)
			if s == "" {
				return 0
			}
			if matchesYes(s) {
				return 1
			}
			return 0
		}
		return 0
	}

	// OpenAIで抽出した項目を単一の文字列として取得（各項目は1つずつ）
	aiParams := extractAIParams(ai)

	// デバッグ: OpenAIで抽出した項目を出力
	fmt.Printf("OpenAI抽出結果: %+v\n", aiParams)

	// format.jsonから項目をスライスに変換
	fmtSlices := extractFormatJSONToSlices(fmtData)

	// デバッグ: format.jsonから読み込んだスライスの数を出力
	for category, items := range fmtSlices {
		fmt.Printf("format.json[%s]: %d件\n", category, len(items))
	}

	// 抽出項目（単一値）とformat.jsonのスライスをfor文でマッピングしてコードに変換
	mappedCodes := mapAIParamsToCodes(aiParams, fmtSlices)

	// デバッグ: マッピングされたコードを出力
	fmt.Printf("マッピングされたコード: %+v\n", mappedCodes)

	// マッピングされたコードをparamsに設定
	if code, ok := mappedCodes["large_area"]; ok {
		params.LargeArea = code
	}
	if code, ok := mappedCodes["middle_area"]; ok {
		params.MiddleArea = code
	}
	if code, ok := mappedCodes["small_area"]; ok {
		params.SmallArea = code
	}
	if code, ok := mappedCodes["genre"]; ok {
		params.Genre = code
	}
	if code, ok := mappedCodes["budget"]; ok {
		params.Budget = code
	}

	// キーワードはそのまま設定
	// キーワードが設定されていない場合でも、他のパラメータがあれば検索可能
	if keyword, ok := aiParams["keyword"]; ok && keyword != "" {
		params.Keyword = keyword
	}
	
	// マッピングが失敗した場合のフォールバック処理
	// 地名、ジャンル、予算のいずれかがマッピングされていない場合、キーワードとして元のプロンプトを使用
	if params.LargeArea == "" && params.MiddleArea == "" && params.SmallArea == "" && 
	   params.Genre == "" && params.Budget == "" && params.Keyword == "" {
		// すべてのパラメータが空の場合、元のプロンプトをキーワードとして使用
		fmt.Printf("警告: すべてのパラメータが空のため、元のプロンプトをキーワードとして使用します\n")
	}

	// フラグ系のパラメータ
	params.PrivateRoom = getFlag(ai.PrivateRoom)
	params.Free_drink = getFlag(ai.FreeDrink)
	params.Free_food = getFlag(ai.FreeFood)
	params.Midnight = getFlag(ai.Midnight)
	params.Sake = getFlag(ai.Sake)
	params.Cacktail = getFlag(ai.Cocktail)
	params.Wine = getFlag(ai.Wine)

	return params, nil
}

// matchesYes は文字列が「はい」を示しているかチェックします
func matchesYes(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Contains(s, "あり") || strings.Contains(s, "有") ||
		strings.Contains(s, "true") || strings.Contains(s, "yes") ||
		s == "1" || s == "○"
}

// resolveCode は人間が読める名前を HotPepper コードに解決します
func resolveCode(category, name string, fmtData map[string]interface{}) string {
	if fmtData == nil {
		return ""
	}

	arr, ok := fmtData[category].([]interface{})
	if !ok {
		return ""
	}

	nameLower := normalizeName(name)

	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if nm, ok := m["name"].(string); ok {
			if strings.Contains(normalizeName(nm), nameLower) || strings.Contains(nameLower, normalizeName(nm)) {
				if code, ok := m["code"].(string); ok {
					return code
				}
			}
		}

		for _, v := range m {
			sub, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			if nm, ok := sub["name"].(string); ok {
				if strings.Contains(normalizeName(nm), nameLower) || strings.Contains(nameLower, normalizeName(nm)) {
					if code, ok := sub["code"].(string); ok {
						return code
					}
				}
			}
		}
	}

	return ""
}

// resolveLocation はロケーション名を大区分/中区分/小区分のコードに解決します
func resolveLocation(name string, fmtData map[string]interface{}) (largeCode, middleCode, smallCode string) {
	if fmtData == nil {
		return "", "", ""
	}

	nameLower := normalizeName(name)

	if arr, ok := fmtData["middle_area"].([]interface{}); ok {
		for _, item := range arr {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if nm, ok := m["name"].(string); ok {
				if strings.Contains(normalizeName(nm), nameLower) || strings.Contains(nameLower, normalizeName(nm)) {
					if code, ok := m["code"].(string); ok {
						middleCode = code
					}
					if la, ok := m["large_area"].(map[string]interface{}); ok {
						if lc, ok := la["code"].(string); ok {
							largeCode = lc
						}
					}
					return
				}
			}
		}
	}

	if arr, ok := fmtData["large_area"].([]interface{}); ok {
		for _, item := range arr {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if nm, ok := m["name"].(string); ok {
				if strings.Contains(normalizeName(nm), nameLower) || strings.Contains(nameLower, normalizeName(nm)) {
					if code, ok := m["code"].(string); ok {
						largeCode = code
						return
					}
				}
			}
		}
	}

	return "", "", ""
}

// normalizeName は名前を比較用に正規化します
func normalizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, suffix := range []string{"市", "区", "都", "県", "道"} {
		s = strings.ReplaceAll(s, suffix, "")
	}
	return strings.TrimSpace(s)
}

// numericBudgetToCode は数値予算（例: 5000）を format.json の budget 名称レンジに基づきコードへ変換します
func numericBudgetToCode(amount int, fmtData map[string]interface{}) string {
	if fmtData == nil {
		return ""
	}
	arr, ok := fmtData["budget"].([]interface{})
	if !ok {
		return ""
	}

	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := m["name"].(string)
		// 数字を抽出する
		re := regexp.MustCompile(`(\d+)`)
		nums := re.FindAllString(name, -1)
		if len(nums) == 0 {
			continue
		}
		if len(nums) == 1 {
			// 例: "～500円" または "5000円以上" のような表現
			var v int
			fmt.Sscanf(nums[0], "%d", &v)
			if strings.Contains(name, "～") || strings.HasPrefix(name, "～") {
				if amount <= v {
					if code, ok := m["code"].(string); ok {
						return code
					}
				}
			} else {
				// 単一数字は下限とみなす
				if amount >= v {
					if code, ok := m["code"].(string); ok {
						return code
					}
				}
			}
		} else if len(nums) >= 2 {
			var low, high int
			fmt.Sscanf(nums[0], "%d", &low)
			fmt.Sscanf(nums[1], "%d", &high)
			if amount >= low && amount <= high {
				if code, ok := m["code"].(string); ok {
					return code
				}
			}
		}
	}
	return ""
}
