package usecase

import (
	"os"
	"restaurant-finder/Domain/entity"
	api "restaurant-finder/Infrastructure/api"
)

// GetRestaurantUsecase HotPepperAPIを使用してレストラン検索を行うユースケース
type GetRestaurantUsecase struct {
	hotPepperClient *api.HotPepperAPIClient
	openaiGenerator *api.OpenAIGenerator
}

// GetRestaurantResult は検索結果と自然言語説明を含む構造体です
type GetRestaurantResult struct {
	Response           *entity.HotPepperResponse
	NaturalDescription string
	SearchParams       *entity.HotPepperRequestParams
}

// NewGetRestaurantUsecase GetRestaurantUsecaseのコンストラクタ
func NewGetRestaurantUsecase() *GetRestaurantUsecase {
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	return &GetRestaurantUsecase{
		hotPepperClient: &api.HotPepperAPIClient{},
		openaiGenerator: api.NewOpenAIGenerator(openaiAPIKey),
	}
}

// GetRestaurant ユーザーの入力からレストランを検索する
func (u *GetRestaurantUsecase) GetRestaurant(prompt string) (*entity.HotPepperResponse, error) {
	// OpenAIを使用してHotPepperAPIのリクエストパラメータを生成
	params, err := u.openaiGenerator.GenerateSearchQuery(prompt)
	if params == nil {
		return nil, err
	}

	// HotPepperAPIを呼び出してレストラン情報を取得
	response, err := u.hotPepperClient.GetRestaurants(params)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetRestaurantWithNaturalLanguage ユーザーの入力からレストランを検索し、自然言語での説明も返す
func (u *GetRestaurantUsecase) GetRestaurantWithNaturalLanguage(prompt string) (*GetRestaurantResult, error) {
	// OpenAIを使用してHotPepperAPIのリクエストパラメータを生成
	params, err := u.openaiGenerator.GenerateSearchQuery(prompt)
	if params == nil {
		return nil, err
	}

	// HotPepperAPIを呼び出してレストラン情報を取得
	response, err := u.hotPepperClient.GetRestaurants(params)
	if err != nil {
		return nil, err
	}

	// 検索結果を自然言語で説明
	var naturalDesc string
	if len(response.Results.Shop) > 0 {
		naturalDesc, err = u.openaiGenerator.GenerateNaturalLanguageResponse(prompt, response.Results.Shop, params)
		if err != nil {
			// 自然言語説明の生成に失敗しても検索結果は返す
			naturalDesc = ""
		}
	}

	return &GetRestaurantResult{
		Response:           response,
		NaturalDescription: naturalDesc,
		SearchParams:       params,
	}, nil
}
