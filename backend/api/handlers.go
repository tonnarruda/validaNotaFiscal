package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	notasFiscais []NotaFiscal
	planilha     []NotaFiscal
)

func RegisterRoutes(r *gin.Engine) {
	r.POST("/api/notas", UploadNotasHandler)
	r.POST("/api/planilha", UploadPlanilhaHandler)
	r.GET("/api/comparar", CompararHandler)
	r.GET("/api/notas", ListarNotasHandler)
}

func UploadNotasHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não encontrado"})
		return
	}
	tmpPath := filepath.Join("tmp", file.Filename)
	if err := c.SaveUploadedFile(file, tmpPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo"})
		return
	}
	defer os.Remove(tmpPath)

	f, err := os.Open(tmpPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao abrir arquivo"})
		return
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == ".pdf" {
		nota, err := ExtractFromPDF(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao processar PDF: %v", err)})
			return
		}
		notasFiscais = append(notasFiscais, nota)
		c.JSON(http.StatusOK, gin.H{"message": "Nota fiscal PDF processada com sucesso", "nota": nota})
		return
	}

	notas, erros := ParseExcelNotas(f)
	if len(erros) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erros ao importar:", "detalhes": erros})
		return
	}
	for _, nota := range notas {
		notasFiscais = append(notasFiscais, nota)
	}
	c.JSON(http.StatusOK, notasFiscais)
}

func UploadPlanilhaHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não encontrado"})
		return
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao abrir arquivo"})
		return
	}
	defer f.Close()

	tmpFile, err := os.CreateTemp("tmp", "planilha-*.xlsx")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar arquivo temporário"})
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao copiar arquivo"})
		return
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao reposicionar arquivo"})
		return
	}

	notas, erros := ParseExcelNotas(tmpFile)
	if len(erros) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erros ao importar:", "detalhes": erros})
		return
	}
	planilha = notas
	c.JSON(http.StatusOK, planilha)
}

func CompararHandler(c *gin.Context) {
	if len(notasFiscais) == 0 || len(planilha) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Importe as notas fiscais e a planilha primeiro"})
		return
	}
	results := []ComparisonResult{}
	for _, nf := range notasFiscais {
		found := false
		for _, pl := range planilha {
			if nf.CNPJ == pl.CNPJ && nf.NumeroNota == pl.NumeroNota {
				results = append(results, ComparisonResult{
					NotaFiscal:    nf,
					DadosPlanilha: pl,
					Match:         nf.Valor == pl.Valor,
				})
				found = true
				break
			}
		}
		if !found {
			results = append(results, ComparisonResult{
				NotaFiscal:    nf,
				DadosPlanilha: NotaFiscal{},
				Match:         false,
			})
		}
	}
	c.JSON(http.StatusOK, results)
}

func ListarNotasHandler(c *gin.Context) {
	c.JSON(http.StatusOK, notasFiscais)
}
