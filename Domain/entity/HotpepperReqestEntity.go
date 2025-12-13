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
	PrivateRoom int     `json:"private_room,omitempty"`
	Count       int     `json:"count,omitempty"`
	Start       int     `json:"start,omitempty"`
	Free_food   int     `json:"free_food,omitempty"`
	Free_drink  int     `json:"free_drink,omitempty"`
	Midnight    int     `json:"midnight,omitempty"`
	Cacktail    int     `json:"cacktail,omitempty"`
	Sake        int     `json:"sake,omitempty"`
	Wine        int     `json:"wine,omitempty"`
}
