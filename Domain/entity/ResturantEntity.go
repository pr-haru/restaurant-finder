package entity

// HotPepperAPIのレスポンスのEntity
// Shop構造体の定義
type Shop struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URLs struct {
		PC string `json:"pc"`
	} `json:"urls"`
	Address string `json:"address"`
	Photo   struct {
		PC struct {
			L string `json:"l"`
			M string `json:"m"`
			S string `json:"s"`
		} `json:"pc"`
		Mobile struct {
			L string `json:"l"`
			S string `json:"s"`
		} `json:"mobile"`
	} `json:"photo"`
	Access string `json:"access"`
	Open   string `json:"open"`
	Close  string `json:"close"`
	Genre  struct {
		Name string `json:"name"`
	} `json:"genre"`
	Catch  string `json:"catch"`
	Budget struct {
		Name string `json:"name"`
	} `json:"budget"`
}

// HotPepperAPiのレスポンスの構造体
type HotPepperResponse struct {
	Results struct {
		APIVersion       string      `json:"api_version"`
		ResultsAvailable int         `json:"results_available"`
		ResultsReturned  interface{} `json:"results_returned"` // 文字列 or int
		ResultsStart     int         `json:"results_start"`
		Shop             []Shop      `json:"shop"`
		Error            []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	} `json:"results"`
}
