package api

import (
	"encoding/json"
	"fmt"

	"restaurant-finder/Domain/entity"
)

// DebugMergeFromJSON は AI の出力 JSON テキストを受け取り、format.json を読み込んで
// mergeAIParamsWithCodes を実行して変換結果を返します。テスト用です。
func DebugMergeFromJSON(aiJSON string) (*entity.HotPepperRequestParams, error) {
	var ai aiOutput
	if err := json.Unmarshal([]byte(aiJSON), &ai); err != nil {
		return nil, fmt.Errorf("ai json unmarshal error: %w", err)
	}

	fmtData, err := loadFormatJSON()
	if err != nil {
		return nil, fmt.Errorf("load format.json error: %w", err)
	}

	params, err := mergeAIParamsWithCodes(&ai, fmtData)
	if err != nil {
		return nil, fmt.Errorf("merge error: %w", err)
	}
	return params, nil
}
