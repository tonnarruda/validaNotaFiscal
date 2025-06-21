package main

import (
	"NF-DECODER-AI/handlers"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis de ambiente do arquivo .env
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	router := gin.Default()

	// Configurar CORS global
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	router.POST("/upload", handlers.DecodeNotaFiscal)
	router.POST("/save-nota-fiscal", handlers.SaveNotaFiscal)
	router.GET("/buscar-notas-fiscais", handlers.BuscarNotasFiscais)

	log.Println("Servidor iniciado na porta 8080")
	log.Fatal(router.Run(":8080"))
}
