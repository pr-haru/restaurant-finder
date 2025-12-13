package repository

import "restaurant-finder/Domain/entity"

//interfaceはtypeから。
//リクエスト作成インターフェイス
type CreateRequest interface {
	GenerateSearchQuery(prompt string) (*entity.HotPepperRequestParams, error)
}
