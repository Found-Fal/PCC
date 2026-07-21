package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"main/koneksi"
	"main/models"
)

// struct untuk menerima input dari client
type InputPenggajian struct {
	NamaPegawai string  `json:"nama_pegawai"`
	GajiPokok   float64 `json:"gaji_pokok"`
	JamLembur   int     `json:"jam_lembur"`
}

// CREATE Penggajian
func CreatePenggajian(c *gin.Context) {
	var input InputPenggajian

	// binding JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// hitung lembur
	upahLemburPerJam := 50000.0
	totalLembur := float64(input.JamLembur) * upahLemburPerJam

	// hitung gaji kotor
	gajiKotor := input.GajiPokok + totalLembur

	// hitung pajak
	var pajak float64
	if gajiKotor > 5000000 {
		pajak = gajiKotor * 0.05
	} else {
		pajak = 0
	}

	// hitung gaji bersih
	gajiBersih := gajiKotor - pajak

	// simpan ke database
	penggajian := models.Penggajian{
		NamaPegawai: input.NamaPegawai,
		GajiPokok:   input.GajiPokok,
		JamLembur:   input.JamLembur,
		GajiKotor:   gajiKotor,
		Pajak:       pajak,
		GajiBersih:  gajiBersih,
	}

	db := koneksi.DB

	if err := db.Create(&penggajian).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menyimpan data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data penggajian berhasil ditambahkan",
		"data":    penggajian,
	})
}

// GET semua data
func GetPenggajian(c *gin.Context) {
	var penggajian []models.Penggajian
	db := koneksi.DB

	if err := db.Find(&penggajian).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": penggajian,
	})
}

// GET by ID
func GetPenggajianByID(c *gin.Context) {
	id := c.Param("id")
	var penggajian models.Penggajian
	db := koneksi.DB

	if err := db.First(&penggajian, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Data tidak ditemukan",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": penggajian,
	})
}

// DELETE
func DeletePenggajian(c *gin.Context) {
	id := c.Param("id")
	db := koneksi.DB

	if err := db.Delete(&models.Penggajian{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghapus data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil dihapus",
	})
}
