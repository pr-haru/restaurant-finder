package repository

import "restaurant-finder/Domain/entity"

//hotpepperのレスポンスをEntityに変換するインターフェイス
//interfaceはtypeから。
//リクエスト作成インターフェイス
type CreateResponse interface {
	GetRestaurants(request entity.HotPepperRequestParams) ([]entity.Shop, error)
}
