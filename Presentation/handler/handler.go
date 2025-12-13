package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"restaurant-finder/Application/usecase"
)

// SearchHandler 検索ページを表示
func SearchHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "search.html", nil)
}

// ProcessSearchHandler 検索リクエストを処理し、結果を表示
func ProcessSearchHandler(c *gin.Context) {
	// フォームから検索クエリを取得
	prompt := c.PostForm("search_query")
	if prompt == "" {
		c.HTML(http.StatusBadRequest, "search.html", gin.H{
			"error": "検索クエリを入力してください",
		})
		return
	}

	// ユースケースを作成して検索を実行、usecaseのメソッド呼び出し
	usecase := usecase.NewGetRestaurantUsecase()
	result, err := usecase.GetRestaurantWithNaturalLanguage(prompt)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "search.html", gin.H{
			"error": "検索中にエラーが発生しました: " + err.Error(),
		})
		return
	}

	// 検索結果をテンプレートに渡す
	//検索ワード、検索件数、自然言語での説明
	c.HTML(http.StatusOK, "search.html", gin.H{
		"restaurants":        result.Response.Results.Shop,
		"query":              prompt,
		"count":              result.Response.Results.ResultsReturned,
		"naturalDescription": result.NaturalDescription,
	})
}
