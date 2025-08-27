package Infrastructure

import (
	"os"
	"fmt"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"restaurant-finder/Domain/entity"
	"strconv" 
)

type HotPepperAPIClient struct{}

func (c *HotPepperAPIClient) GetRestaurants(params *entity.HotPepperRequestParams) (*entity.HotPepperResponse, error) {
	hotpepperAPIKey := os.Getenv("HOTPEPPER_API_KEY")
	if hotpepperAPIKey == "" {
		return nil, fmt.Errorf("HotPepperAPI通信エラーです。")
	}
	fmt.Printf("HotPepper API request params: %+v\n", params)
	// HotPepperAPIのベースURL
	baseURL := "https://webservice.recruit.co.jp/hotpepper/gourmet/v1/"
	
	// クエリパラメータを構築
	//baseURLの?以降の部分がクエリパラメータ作成のため
	queryParams := url.Values{}
	//APIきー。?key以降
	queryParams.Set("key", hotpepperAPIKey)
	queryParams.Set("format", "json")
	
	// パラメータが設定されている場合のみ追加
	if params.Keyword != "" {
		queryParams.Set("keyword", params.Keyword)
	}
	if params.Lat != 0 {
		queryParams.Set("lat", strconv.FormatFloat(params.Lat, 'f', -1, 64))
	}
	if params.Lng != 0 {
		queryParams.Set("lng", fmt.Sprintf("%f", params.Lng))
	}
	if params.Range != 0 {
		queryParams.Set("range", fmt.Sprintf("%d", params.Range))
	}
	if params.LargeArea != "" {
		queryParams.Set("large_area", params.LargeArea)
	}
	if params.MiddleArea != "" {
		queryParams.Set("middle_area", params.MiddleArea)
	}
	if params.SmallArea != "" {
		queryParams.Set("small_area", params.SmallArea)
	}
	if params.Genre != "" {
		queryParams.Set("genre", params.Genre)
	}
	if params.Budget != "" {
		queryParams.Set("budget", params.Budget)
	}
	if params.Lunch != 0 {
		queryParams.Set("lunch", fmt.Sprintf("%d", params.Lunch))
	}
	if params.PrivateRoom != 0 {
		queryParams.Set("private_room", fmt.Sprintf("%d", params.PrivateRoom))
	}
	if params.Count != 0 {
		queryParams.Set("count",  strconv.Itoa(params.Count))
	}
	if params.Start != 0 {
		queryParams.Set("start", fmt.Sprintf("%d", params.Start))
	}

	// APIリクエストを送信
	fullURL := baseURL + "?" + queryParams.Encode()
	fmt.Printf("HotPepper API URL: %s\n", fullURL)
	
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %v", err)
	}
	defer resp.Body.Close()
	
	fmt.Printf("HotPepper API response status: %s\n", resp.Status)

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// JSONをパース
	var hotPepperResponse entity.HotPepperResponse
	if err := json.Unmarshal(body, &hotPepperResponse); err != nil {
		fmt.Printf("Failed to parse JSON response: %v\n", err)
		fmt.Printf("Response body: %s\n", string(body))
		return nil, fmt.Errorf("failed to parse JSON response: %v", err)
	}
	
	fmt.Printf("HotPepper API response parsed successfully. Found %d shops\n", len(hotPepperResponse.Results.Shop))

	return &hotPepperResponse, nil
}