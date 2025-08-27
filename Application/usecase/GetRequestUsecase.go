package usecase

import (
	"fmt"
	api "restaurant-finder/Infrastructure/api"
	"restaurant-finder/Domain/entity"
)

// GetRestaurantUsecase HotPepperAPIを使用してレストラン検索を行うユースケース
type GetRestaurantUsecase struct {
	hotPepperClient *api.HotPepperAPIClient
}

// GetRestaurant ユーザーの入力からレストランを検索する
func (u *GetRestaurantUsecase) GetRestaurant(prompt string) (*entity.HotPepperResponse, error) {
	// OpenAIを使用してHotPepperAPIのリクエストパラメータを生成
 //newしてその下でメソッド呼び出している感じ
	generator := &api.OpenAIGenerator{} 
	 params, err := generator.GenerateSearchQuery(prompt)
	if params == nil {
		return nil,err
	}

	// HotPepperAPIを呼び出してレストラン情報を取得
	response, err := u.hotPepperClient.GetRestaurants(params)
	if err != nil {
		return nil,err
	}

	return response, nil
}