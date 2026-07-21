package controllers

import (
	"crypto/sha1"
	"fmt"
	"main/models"
	"net/http"

	jwtV3 "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Binding dari POST JSON
type StrukturUserTambah struct {
	Nama     string `json:"nama" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type StrukturUserUbah struct {
	Id       uint   `binding:"required"`
	Nama     string `binding:"required"`
	Username string `binding:"required"`
	Password string `binding:"required"`
}

type StrukturUserHapus struct {
	Id uint `binding:"required"`
}

type StrukturLogin struct {
	Username string `binding:"required"`
	Password string `binding:"required"`
}

func UserTampil(c *gin.Context) {
	//ambil koneksi variabel db dari main
	db := c.MustGet("db").(*gorm.DB)
	//buat variabel array dari model User
	var modeluser []models.User
	hasil := db.Find(&modeluser)
	kesalahan := hasil.Error
	if hasil.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":    true,
			"pesan":     "Berhasil Tampil data",
			"kesalahan": nil,
			"data":      modeluser,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":    false,
			"pesan":     "Gagal Tampil Data",
			"kesalahan": kesalahan.Error(),
			"data":      nil,
		})
	}
}
func UserTambah(c *gin.Context) {
	//ambil koneksi variabel db dari main
	db := c.MustGet("db").(*gorm.DB)
	//membuat variabel data User dengan struktur user
	// dan menangkap data dari request
	var datauser StrukturUserTambah
	if err := c.ShouldBindJSON(&datauser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"pesan":     "Gagal membaca Data",
			"kesalahan": err.Error(),
		})
		return
	}

	//enkripsi password dengan sha1
	var sha = sha1.New()
	sha.Write([]byte(datauser.Password))
	var encrypted = sha.Sum(nil)
	var encryptedString = fmt.Sprintf("%x", encrypted)

	//membuat data baru dengan model user
	modeluser := models.User{
		Nama:     datauser.Nama,
		Username: datauser.Username,
		Password: encryptedString,
	}
	hasil := db.Create(&modeluser)

	if hasil.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"status": true,
			"pesan":  "Berhasil tambah data",
			"data":   modeluser,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":    false,
			"pesan":     "Gagal tambah data",
			"kesalahan": hasil.Error.Error(),
		})
	}
}

func UserUbah(c *gin.Context) {
	//ambil koneksi variabel db dari main
	db := c.MustGet("db").(*gorm.DB)
	//membuat variabel data User dengan struktur user
	//dan menangkap data dari request
	var datauser StrukturUserUbah
	if err := c.ShouldBindJSON(&datauser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"pesan":     "Gagal membaca Data",
			"kesalahan": err.Error(),
		})
		return
	}
	//membuat variabel model user
	var modeluser models.User
	//mencari data user dan merubah datanya
	cekUser := db.First(&modeluser, datauser.Id)
	if cekUser.Error == nil {
		//ekripsi password dengan sha1
		var sha = sha1.New()
		sha.Write([]byte(datauser.Password))
		var encrypted = sha.Sum(nil)
		var encryptedString = fmt.Sprintf("%x", encrypted)

		modeluser.Nama = datauser.Nama
		modeluser.Username = datauser.Username
		modeluser.Password = encryptedString
		hasil := db.Save(&modeluser)

		kesalahan := hasil.Error
		if hasil.Error == nil {
			c.JSON(http.StatusOK, gin.H{
				"status":    true,
				"pesan":     "Berhasil ubah data",
				"kesalahan": nil,
				"data":      modeluser,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":    false,
				"pesan":     "Gagal ubah Data",
				"kesalahan": kesalahan.Error(),
				"data":      modeluser,
			})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":    false,
			"pesan":     "Data tidak ditemukan",
			"kesalahan": cekUser.Error.Error(),
			"data":      modeluser,
		})
	}
}

func UserHapus(c *gin.Context) {
	//ambil koneksi variabel db dari main
	db := c.MustGet("db").(*gorm.DB)
	//membuat variabel data user dengan struktur user
	//dan menangkap data dari request
	var datauser StrukturUserHapus
	if err := c.ShouldBindJSON(&datauser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":    false,
			"pesan":     "Gagal membaca Data",
			"kesalahan": err.Error(),
		})
		return
	}
	//membuat variabel model user
	var modeluser models.User
	//menghapus data user berdasarkan Id yang dikirim
	hasil := db.Delete(&modeluser, datauser.Id)
	kesalahan := hasil.Error

	if hasil.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":    true,
			"pesan":     "Berhasil hapus data",
			"kesalahan": nil,
			"data":      datauser,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":    false,
			"pesan":     "Gagal hapus Data",
			"kesalahan": kesalahan.Error(),
			"data":      datauser,
		})
	}
}

func UserLogin(c *gin.Context) (any, error) {
	//ambil koneksi variabel db dari main
	db := c.MustGet("db").(*gorm.DB)
	//membuat variabel data user dengan struktur user dan menangkap data dari request
	var dataLogin StrukturLogin
	if err := c.ShouldBindJSON(&dataLogin); err != nil {
		//kembalikan data kosong dan eror input login
		return nil, jwtV3.ErrMissingLoginValues
	}

	//enkripsi password dengan sha1
	var sha = sha1.New()
	sha.Write([]byte(dataLogin.Password))
	var encrypted = sha.Sum(nil)
	var encryptedString = fmt.Sprintf("%x", encrypted)
	//membuat variabel model user
	var modelUser models.User
	//mencari data user berdasarkan username dan password
	cekUser := db.Where("username = ? AND password = ?", dataLogin.Username, encryptedString).First(&modelUser)
	if cekUser.Error == nil {
		//kembalikan data user dan eror=nil
		return &modelUser, nil
	} else {
		//kembalikan data kosong dan eror gagal login
		return nil, jwtV3.ErrFailedAuthentication
	}
}
