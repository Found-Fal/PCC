package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io" // Menggantikan ioutil
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"main/models"
)

func DriveUpload(c *gin.Context) {
	fileName := c.PostForm("fileName")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"pesan": "File tidak ditemukan"})
		return
	}

	mimeType := file.Header.Get("Content-Type")
	fileOpen, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"pesan": "Gagal membuka file"})
		return
	}
	defer fileOpen.Close()

	fileData, _ := io.ReadAll(fileOpen)

	// Encode ke base64
	data := base64.StdEncoding.EncodeToString(fileData)
	postBody, _ := json.Marshal(map[string]string{
		"fileName": fileName,
		"mimeType": mimeType,
		"data":     data,
	})
	requestBody := bytes.NewBuffer(postBody)

	// Post data ke Google Apps Script
	res, err := http.Post(
		"https://script.google.com/macros/s/AKfycbxMmMtbbcNWsxuYBMIdD2_XvkPuCCDHvn7GnWCHYDxNkDboMYjgtfEkSAwMsHwxX8wOWA/exec",
		"application/json; charset=UTF-8",
		requestBody,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"kode_error": "ERR-DRIVE",
			"pesan":      "Gagal Upload ke Drive",
		})
		return
	}
	defer res.Body.Close() // Pastikan ditutup di sini

	// Baca response data
	hasilBody, _ := io.ReadAll(res.Body)

	// Konversi JSON string ke Map
	var hasilJson map[string]interface{}
	if err := json.Unmarshal(hasilBody, &hasilJson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"pesan": "Gagal membaca response Drive"})
		return
	}

	// === PROSES DATABASE (Sekarang berada di dalam fungsi) ===
	db := c.MustGet("db").(*gorm.DB)

	// Menggunakan type assertion yang aman supaya tidak panic jika nil
	namaDokumen, _ := hasilJson["fileName"].(string) // Sesuaikan dengan response Google Script Anda (case-sensitive)
	fileID, _ := hasilJson["fileId"].(string)
	fileURL, _ := hasilJson["fileUrl"].(string)

	dokumenBaru := models.Dokumen{
		NamaDokumen: namaDokumen,
		FileId:      fileID,
		FileUrl:     fileURL,
	}

	// Membuat record baru di database
	if err := db.Create(&dokumenBaru).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": false,
			"pesan":  "Gagal menyimpan ke database",
			"error":  err.Error(),
		})
		return
	}

	// Response Sukses
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"pesan":  "Berhasil Upload dan Simpan Database",
		"data":   dokumenBaru,
	})
}

// tampil data dari tabel
func DriveTampil(c *gin.Context) {
	//ambil koneksi ke variabel db
	db := c.MustGet("db").(*gorm.DB)
	//membuat variabel dokumen berupa array dari model Dokumen
	var dokumen []models.Dokumen
	//ambil semua data dari tabel
	db.Find(&dokumen)
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"pesan":  "Berhasil Tampil",
		"data":   dokumen,
	})
}

// direct download dari google drive
func DriveUnduh(c *gin.Context) {
	id := c.Param("id")
	//get data
	res, err := http.Get(" https://script.google.com/macros/s/AKfycbxMmMtbbcNWsxuYBMIdD2_XvkPuCCDHvn7GnWCHYDxNkDboMYjgtfEkSAwMsHwxX8wOWA/exec?id=" + id)
	//apakah ada error
	if err != nil {
		c.JSON(500, gin.H{
			"status": false,
			"pesan":  "Gagal Unduh",
		})
		return
	}
	//baca response data
	hasilBody, _ := io.ReadAll(res.Body)
	hasilString := string(hasilBody)
	//konversi string json to object
	var hasilJson map[string]interface{}
	json.Unmarshal([]byte(hasilString), &hasilJson)
	//ambil data file dan mimeType
	fileBase64 := hasilJson["file"].(string)
	mimeType := hasilJson["mimeType"].(string)
	// konversi base64 ke file
	file, _ := base64.StdEncoding.DecodeString(fileBase64)
	//tulis dengan header sesuai dengan mimeType
	c.Writer.Header().Set("Content-Type", mimeType)
	c.Writer.Write(file)
}
