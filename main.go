package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
)

// --- Shop構造体の定義を追加 ---
type Shop struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Photo     struct {
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
	Access      string `json:"access"`
	Open        string `json:"open"`
	Close       string `json:"close"`
	CouponURLs  struct {
		PC     string `json:"pc"`
		SP     string `json:"sp"`
	} `json:"coupon_urls"`
	// 必要に応じて他のフィールドも追加
}

type HotPepperResponse struct {
	Results struct {
		APIVersion       string `json:"api_version"`
		ResultsAvailable int    `json:"results_available"`
		ResultsReturned  string    `json:"results_returned"`
		ResultsStart     int    `json:"results_start"`
		Shop             []Shop `json:"shop"` // ここが店舗情報の配列
		Error            []struct { // エラーレスポンスも考慮
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	} `json:"results"`
}

var hotpepperAPIKey string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	hotpepperAPIKey = os.Getenv("HOTPEPPER_API_KEY")
	if hotpepperAPIKey == "" {
		log.Fatal("HOTPEPPER_API_KEY is not set")
	}

	router := gin.Default()
	
	// 静的ファイルの設定
	router.Static("/static", "./static")
	
	// テンプレートの設定
	router.LoadHTMLGlob("templates/*.html")
	
	// ルートの設定
	router.GET("/", searchHandler)
	router.GET("/search", processSearchHandler)
	// 重複したGET /searchルートを削除

	router.Run(":8080")
}

func searchHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "search.html", gin.H{
        "Keyword": "",
        "Shops":   []Shop{}, // 空のスライス
        "Found":   false,    // 結果が見つかったかどうかのフラグ
        "Error":   "",       // エラーメッセージも空にする
    })
}

func processSearchHandler(c *gin.Context) {
	keyword := c.PostForm("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "キーワードが入力されていません"})
		return
	}

	// HotPepper APIを呼び出し
	shops, err := searchHotPepperAPI(keyword)
	if err != nil {
		log.Printf("API呼び出しエラー: %v", err)
		c.HTML(http.StatusOK, "search.html", gin.H{
			"Keyword": keyword,
			"Shops":   []Shop{},
			"Found":   false,
			"Error":   "API呼び出しに失敗しました",
		})
		return
	}

	c.HTML(http.StatusOK, "search.html", gin.H{
		"Keyword": keyword,
		"Shops":   shops,
		"Found":   len(shops) > 0,
		"Error":   "",
	})
}

func searchHotPepperAPI(keyword string) ([]Shop, error) {

	baseURL := "https://webservice.recruit.co.jp/hotpepper/gourmet/v1/"
	
	// クエリパラメータを構築
	params := url.Values{}
	params.Add("key", hotpepperAPIKey)
	params.Add("keyword", keyword)
	params.Add("format", "json")
	params.Add("count", "20") // 最大20件取得

	// URLを構築
	requestURL := baseURL + "?" + params.Encode()
	
	// HTTPリクエストを実行
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエストエラー: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りエラー: %v", err)
	}

	// JSONをパース
	var response HotPepperResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("JSONパースエラー: %v", err)
	}
    log.Printf("HotPepper APIレスポンス - ResultsAvailable: %d", response.Results.ResultsAvailable)
    log.Printf("HotPepper APIレスポンス - ResultsReturned: %s", response.Results.ResultsReturned)
    log.Printf("HotPepper APIレスポンス - Shops found: %d", len(response.Results.Shop))
	// エラーチェック
	if len(response.Results.Error) > 0 {
		errorMsg := response.Results.Error[0].Message
		return nil, fmt.Errorf("APIエラー: %s", errorMsg)
	}
	// 店舗情報を返す
	return response.Results.Shop, nil
}