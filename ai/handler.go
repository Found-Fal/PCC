package ai

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatRequest struct {
	Pesan string `json:"pesan" binding:"required"`
}

func ChatHandler(c *gin.Context) {
	var req ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Format request tidak valid",
			"error":   err.Error(),
		})
		return
	}
	log.Println("AI Request:", req.Pesan)

	jawaban := TanyaAi("", req.Pesan)

	log.Println("AI Response berhasil diterima")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"pesan":   req.Pesan,
		"jawaban": jawaban,
	})
}
