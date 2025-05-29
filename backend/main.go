package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
	pdfcpu "rsc.io/pdf"
)

type NotaFiscal struct {
	ID         string  `json:"id"`
	CNPJ       string  `json:"cnpj"`
	NumeroNota string  `json:"numeroNota"`
	Valor      float64 `json:"valor"`
}

type ComparisonResult struct {
	NotaFiscal    NotaFiscal `json:"notaFiscal"`
	DadosPlanilha NotaFiscal `json:"dadosPlanilha"`
	Match         bool       `json:"match"`
}

var (
	notasFiscais []NotaFiscal
	planilha     []NotaFiscal
)

var (
	cnpjRegex       = regexp.MustCompile(`InscMunicipal(\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2})`)
	numeroNotaRegex = regexp.MustCompile(`NúmerodaNFS-e\s*(\d+)`)
	valorRegex      = regexp.MustCompile(`\(-\) Desconto Condicionado\s*(\d+(?:\.\d{3})*,\d{2})?`)
)

// Função para desfragmentar texto (remover espaços entre caracteres)
func defragmentText(s string) string {
	// Junta caracteres separados por espaço, mas mantém palavras separadas
	return strings.ReplaceAll(s, " ", "")
}

func extractFromPDF(file *os.File) (NotaFiscal, error) {
	fileInfo, err := file.Stat()
	if err != nil {
		return NotaFiscal{}, fmt.Errorf("erro ao obter tamanho do arquivo: %v", err)
	}

	reader, err := pdf.NewReader(file, fileInfo.Size())
	if err != nil {
		return NotaFiscal{}, fmt.Errorf("erro ao ler PDF: %v", err)
	}

	var content strings.Builder
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		content.WriteString(text)
	}

	text := content.String()
	fmt.Println("=== TEXTO EXTRAÍDO LEDONGTHUC/PDF ===")
	fmt.Println(text)
	fmt.Println("======================================")

	var cnpj string
	if matches := cnpjRegex.FindStringSubmatch(text); len(matches) > 1 {
		cnpj = strings.TrimSpace(matches[1])
	}
	fmt.Println("=== CNPJ EXTRAÍDO ===")
	fmt.Println(cnpj)
	fmt.Println("===========================")

	var numeroNota string
	if matches := numeroNotaRegex.FindStringSubmatch(text); len(matches) > 1 {
		numeroNota = strings.TrimSpace(matches[1])
	}

	// Fallback para rsc.io/pdf se não encontrar CNPJ ou número da nota
	if cnpj == "" || numeroNota == "" {
		fmt.Println("Tentando extração com rsc.io/pdf...")
		file.Seek(0, 0)
		pdfReader, err := pdfcpu.Open(file.Name())
		if err == nil {
			var pdfText strings.Builder
			for i := 0; i < pdfReader.NumPage(); i++ {
				p := pdfReader.Page(i + 1)
				if p.V.IsNull() {
					continue
				}
				content := p.Content()
				for _, txt := range content.Text {
					pdfText.WriteString(txt.S)
					pdfText.WriteString(" ")
				}
			}
			fallbackText := pdfText.String()
			// Desfragmentar texto extraído
			fallbackText = defragmentText(fallbackText)
			fmt.Println("=== TEXTO DESFRAGMENTADO RSC.IO/PDF ===")
			fmt.Println(fallbackText)
			fmt.Println("========================================")
			if cnpj == "" {
				if matches := cnpjRegex.FindStringSubmatch(fallbackText); len(matches) > 1 {
					cnpj = strings.TrimSpace(matches[1])
				}
			}
			if numeroNota == "" {
				if matches := numeroNotaRegex.FindStringSubmatch(fallbackText); len(matches) > 1 {
					numeroNota = strings.TrimSpace(matches[1])
				}
			}
		}
	}

	var valor float64
	if matches := valorRegex.FindStringSubmatch(text); len(matches) > 1 && matches[1] != "" {
		valorStr := strings.ReplaceAll(matches[1], ".", "")
		valorStr = strings.ReplaceAll(valorStr, ",", ".")
		if v, err := strconv.ParseFloat(valorStr, 64); err == nil {
			valor = v
		}
	}

	notaFiscal := NotaFiscal{
		CNPJ:       cnpj,
		NumeroNota: numeroNota,
		Valor:      valor,
	}

	jsonData, err := json.MarshalIndent(notaFiscal, "", "  ")
	if err == nil {
		fmt.Println("\n=== DADOS DA NOTA FISCAL ===")
		fmt.Println(string(jsonData))
		fmt.Println("===========================")
	}

	return notaFiscal, nil
}

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	if err := os.MkdirAll("tmp", 0755); err != nil {
		log.Fatal("Erro ao criar diretório temporário:", err)
	}

	// Upload de nota fiscal (PDF ou Excel)
	r.POST("/api/notas", func(c *gin.Context) {
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
			nota, err := extractFromPDF(f)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao processar PDF: %v", err)})
				return
			}

			notasFiscais = append(notasFiscais, nota)

			c.JSON(http.StatusOK, gin.H{
				"message": "Nota fiscal PDF processada com sucesso",
				"nota":    nota,
			})
			return
		}

		// Excel: adiciona notas sem limpar as anteriores
		f.Seek(0, io.SeekStart)
		xlsx, err := excelize.OpenReader(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler arquivo Excel"})
			return
		}

		rows, err := xlsx.GetRows("Sheet1")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler planilha"})
			return
		}

		importadas := 0
		errosConversao := []string{}
		for i, row := range rows {
			if i == 0 {
				continue
			}
			// Ignorar linhas totalmente vazias
			if len(row) < 3 || (strings.TrimSpace(row[0]) == "" && strings.TrimSpace(row[1]) == "" && strings.TrimSpace(row[2]) == "") {
				continue
			}
			valorStr := strings.TrimSpace(row[2])
			valorStr = strings.ReplaceAll(valorStr, ".", "")  // Remove separador de milhar
			valorStr = strings.ReplaceAll(valorStr, ",", ".") // Troca vírgula por ponto
			valor, err := strconv.ParseFloat(valorStr, 64)
			if err != nil {
				errosConversao = append(errosConversao, fmt.Sprintf("Linha %d: valor inválido '%s'", i+1, row[2]))
				continue
			}
			notasFiscais = append(notasFiscais, NotaFiscal{
				CNPJ:       strings.TrimSpace(row[0]),
				NumeroNota: strings.TrimSpace(row[1]),
				Valor:      valor,
			})
			importadas++
		}
		if len(errosConversao) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Erros ao importar:", "detalhes": errosConversao})
			return
		}

		c.JSON(http.StatusOK, notasFiscais)
	})

	// Upload de planilha de comparação
	r.POST("/api/planilha", func(c *gin.Context) {
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

		xlsx, err := excelize.OpenReader(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler arquivo Excel"})
			return
		}

		planilha = []NotaFiscal{}
		rows, err := xlsx.GetRows("Sheet1")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler planilha"})
			return
		}

		importadas := 0
		errosConversao := []string{}
		for i, row := range rows {
			if i == 0 {
				continue
			}
			// Ignorar linhas totalmente vazias
			if len(row) < 3 || (strings.TrimSpace(row[0]) == "" && strings.TrimSpace(row[1]) == "" && strings.TrimSpace(row[2]) == "") {
				continue
			}
			valorStr := strings.TrimSpace(row[2])
			valorStr = strings.ReplaceAll(valorStr, ".", "")  // Remove separador de milhar
			valorStr = strings.ReplaceAll(valorStr, ",", ".") // Troca vírgula por ponto
			valor, err := strconv.ParseFloat(valorStr, 64)
			if err != nil {
				errosConversao = append(errosConversao, fmt.Sprintf("Linha %d: valor inválido '%s'", i+1, row[2]))
				continue
			}
			planilha = append(planilha, NotaFiscal{
				CNPJ:       strings.TrimSpace(row[0]),
				NumeroNota: strings.TrimSpace(row[1]),
				Valor:      valor,
			})
			importadas++
		}
		if len(errosConversao) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Erros ao importar:", "detalhes": errosConversao})
			return
		}

		c.JSON(http.StatusOK, planilha)
	})

	// Comparar notas fiscais com planilha
	r.GET("/api/comparar", func(c *gin.Context) {
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
					NotaFiscal: nf,
					DadosPlanilha: NotaFiscal{
						CNPJ:       "",
						NumeroNota: "",
						Valor:      0,
					},
					Match: false,
				})
			}
		}

		c.JSON(http.StatusOK, results)
	})

	// Nova rota para listar todas as notas fiscais importadas
	r.GET("/api/notas", func(c *gin.Context) {
		c.JSON(http.StatusOK, notasFiscais)
	})

	r.Run(":8080")
}
