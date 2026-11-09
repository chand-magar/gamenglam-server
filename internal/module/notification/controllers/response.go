package controllers

import "github.com/gin-gonic/gin"

func errorResponse(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}
