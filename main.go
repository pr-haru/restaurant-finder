package main

import (
 "github.com/gin-gonic/gin")

func main(){
   router := gin.Default()
   router.LoadHTMLGlob("templates/*.html")
   router.GET("/",searchHandler)
   router.POST("/search",processSearchHandler)

   router.Run(":8080")
}

func searchHandler(c *gin.Context){
    c.Html(http.statusOK, "index.html", gin.H{})
}