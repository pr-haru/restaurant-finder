package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"restaurant-finder/Presentation/handler"
)

var hotpepperAPIKey string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	hotpepperAPIKey = os.Getenv("HOTPEPPER_API_KEY")
	if hotpepperAPIKey == "" {
		log.Fatal("HOTPEPPER_API_KEY is not set")
	}
	
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}
	
	log.Printf("Environment variables loaded successfully")

	router := gin.Default()

	// 静的ファイルの設定
	router.Static("/CSS", "./CSS")
	router.LoadHTMLGlob("templates/*.html")
	
	// ルートの設定
	router.GET("/", handler.SearchHandler)
	router.POST("/search", handler.ProcessSearchHandler)

	router.Run(":8080")
}