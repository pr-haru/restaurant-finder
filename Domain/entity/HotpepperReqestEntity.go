package entity

// HotPepperRequestParams は、Hot Pepper グルメサーチAPIのリクエストパラメータを定義します。
type HotPepperRequestParams struct {
	Key         string  `json:"key"`
	Format      string  `json:"format"`
	Keyword     string  `json:"keyword,omitempty"`
	Lat         float64 `json:"lat,omitempty"`
	Lng         float64 `json:"lng,omitempty"`
	Range       int     `json:"range,omitempty"`
	LargeArea   string  `json:"large_area,omitempty"`
	MiddleArea  string  `json:"middle_area,omitempty"`
	SmallArea   string  `json:"small_area,omitempty"`
	Genre       string  `json:"genre,omitempty"`
	Budget      string  `json:"budget,omitempty"`
	Lunch       int     `json:"lunch,omitempty"`
	party_capacity int  `json:"private_room,omitempty"`
	PrivateRoom int     `json:"private_room,omitempty"`
	Count       int     `json:"count,omitempty"`
	Start       int     `json:"start,omitempty"`
}
