//interfaceはtypeから。
//リクエスト作成インターフェイス
type CreateRequest interface {
	GenerateSearchQuery(prompt string) (*HotPepperRequestParams, error)
}