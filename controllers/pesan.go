package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"main/models"
)

// Menampilkan semua pesan
func PesanTampil(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var pesan []models.Pesan

	db.Find(&pesan)

	c.JSON(http.StatusOK, gin.H{
		"data": pesan,
	})
}

// Menambah pesan baru
func PesanTambah(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var pesan models.Pesan

	if err := c.ShouldBindJSON(&pesan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	db.Create(&pesan)

	c.JSON(http.StatusOK, gin.H{
		"message": "Data pesan berhasil ditambahkan",
		"data":    pesan,
	})
}

// Mengubah pesan
func PesanUbah(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var pesan models.Pesan

	kode := c.Query("kode")

	if err := db.Where("kode = ?", kode).First(&pesan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Data tidak ditemukan",
		})
		return
	}

	var input models.Pesan

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	pesan.Balasan = input.Balasan

	db.Save(&pesan)

	c.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil diubah",
		"data":    pesan,
	})
}

// Menghapus pesan
func PesanHapus(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var pesan models.Pesan

	kode := c.Query("kode")

	if err := db.Where("kode = ?", kode).First(&pesan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Data tidak ditemukan",
		})
		return
	}

	db.Delete(&pesan)

	c.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil dihapus",
	})
}
