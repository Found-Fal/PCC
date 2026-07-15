package main

import (
	"log"
	"main/fungsi"
	"main/models"
	"main/wa"
	"os"
	"time"

	"main/controllers"
	"main/koneksi"

	jwtV3 "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"

	"main/ai"
)

func main() {
	godotenv.Load()

	koneksi.Connect()
	db := koneksi.DB

	// 🔥 HARUS DIBUAT DULU
	r := gin.Default()

	// root route
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Server Gin berjalan",
		})
	})

	db.AutoMigrate(&models.Suhu{})
	db.AutoMigrate(&models.Informasi{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Pesan{})
	db.AutoMigrate(&models.Penggajian{})
	db.AutoMigrate(&models.Dokumen{})

	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	key_jwt := os.Getenv("KEY_JWT")
	authMiddleware, err := jwtV3.New(&jwtV3.GinJWTMiddleware{
		Realm:       "PCC",
		Key:         []byte(key_jwt),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour * 24,
		IdentityKey: "id",
		PayloadFunc: func(data any) jwt.MapClaims {
			value, ok := data.(models.User)
			if ok {
				return jwt.MapClaims{
					"id":   value.ID,
					"nama": value.Nama,
				}
			}
			return jwt.MapClaims{}
		},
		Authenticator: controllers.UserLogin,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	errInit := authMiddleware.MiddlewareInit()
	if errInit != nil {
		log.Fatal("JWT MiddlewareInit Error:" + errInit.Error())
	}

	r.POST("/login", authMiddleware.LoginHandler)

	auth := r.Group("/backend", authMiddleware.MiddlewareFunc())

	auth.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": true,
			"pesan":  "Berhasil tampil",
		})
	})

	auth.POST("/programstudi", fungsi.BacaDataProdi)

	auth.GET("/suhu", controllers.Tampil)
	auth.POST("/suhu", controllers.Tambah)
	auth.PUT("/suhu", controllers.Ubah)
	auth.DELETE("/suhu", controllers.Hapus)

	auth.GET("/informasi", controllers.InformasiTampil)
	auth.POST("/informasi", controllers.InformasiTambah)
	auth.PUT("/informasi", controllers.InformasiUbah)
	auth.DELETE("/informasi", controllers.InformasiHapus)

	auth.GET("/user", controllers.UserTampil)
	auth.POST("/user", controllers.UserTambah)
	auth.PUT("/user", controllers.UserUbah)
	auth.DELETE("/user", controllers.UserHapus)

	auth.POST("/penggajian", controllers.CreatePenggajian)
	auth.GET("/penggajian", controllers.GetPenggajian)
	auth.GET("/penggajian/:id", controllers.GetPenggajianByID)
	auth.DELETE("/penggajian/:id", controllers.DeletePenggajian)

	//METHOD DRIVE
	auth.POST("/drive", controllers.DriveUpload)
	auth.GET("/drive", controllers.DriveTampil)
	auth.GET("/drive/:id", controllers.DriveUnduh)

	auth.GET("/pesan", controllers.PesanTampil)
	auth.POST("/pesan", controllers.PesanTambah)
	auth.PUT("/pesan", controllers.PesanUbah)
	auth.DELETE("/pesan", controllers.PesanHapus)

	port := os.Getenv("PORT")
	go r.Run(":" + port)
	//ai.MulaiChatAi()
	ai.InitAi()
	wa.KonekWa(db)
}
