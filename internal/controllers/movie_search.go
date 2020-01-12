package controllers

import (
	"github.com/gin-gonic/gin"
	"media-web/internal/moviedb"
)

func GetMoveSearchHandler() (func (ctx *gin.Context)) {
	return func(ctx *gin.Context) {
		query := ctx.Query("query")

		if query != "" {
			result, err := moviedb.Search(query)

			if err != nil {
				ctx.JSON(500, gin.H{
					"message" : "Failed to search the movie api",
				})
				return
			}
			ctx.JSON(200, result)
		} else {
			ctx.JSON(400, gin.H{
				"message" : "query string is required",
			})
		}
	}
}
