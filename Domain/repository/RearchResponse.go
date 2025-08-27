//hotpepperのレスポンスをEntityに変換するインターフェイス
//interfaceはtypeから。
//リクエスト作成インターフェイス
type CreateResponse interface {
	GetRestaurants(request HotPepperRequestParams) ([]Shop, error)
}
