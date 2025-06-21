package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenData representa um token de validação
type TokenData struct {
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NotaFiscalData representa os dados da nota fiscal a ser salva
type NotaFiscalData struct {
	Email         string  `json:"email"`
	NumeroNota    string  `json:"numeroNota"`
	Competencia   string  `json:"competencia"`
	Prestador     string  `json:"prestador"`
	CNPJ          string  `json:"cnpj"`
	ValorServicos float64 `json:"valorServicos"`
	DataNota      string  `json:"dataNota"`
	ISSRetido     float64 `json:"issRetido"`
}

// Map para armazenar tokens temporariamente (em produção, use um banco de dados)
var tokenStore = make(map[string]TokenData)

// generateToken gera um token aleatório de 6 caracteres
func generateToken() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:6]
}

// SendValidationToken envia um token de validação para o email
func SendValidationToken(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email inválido"})
		return
	}

	// Gerar token
	token := generateToken()

	// Armazenar token (expira em 10 minutos)
	tokenData := TokenData{
		Email:     request.Email,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	tokenStore[request.Email] = tokenData

	// Em produção, aqui você enviaria o email
	// Por enquanto, apenas logamos o token
	log.Printf("Token de validação para %s: %s", request.Email, token)

	c.JSON(http.StatusOK, gin.H{
		"message": "Token de validação enviado com sucesso",
		"token":   token, // Em produção, remova esta linha
	})
}

// ValidateToken valida o token enviado pelo usuário
func ValidateToken(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Buscar token armazenado
	tokenData, exists := tokenStore[request.Email]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token não encontrado"})
		return
	}

	// Verificar se o token expirou
	if time.Now().After(tokenData.ExpiresAt) {
		delete(tokenStore, request.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expirado"})
		return
	}

	// Verificar se o token está correto
	if tokenData.Token != request.Token {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	// Token válido - remover do store
	delete(tokenStore, request.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Token validado com sucesso",
	})
}

// SaveNotaFiscal salva a nota fiscal no sistema
func SaveNotaFiscal(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar formulário"})
		return
	}

	// Extrair dados do formulário
	email := c.PostForm("email")
	numeroNota := c.PostForm("numeroNota")
	competencia := c.PostForm("competencia")

	if email == "" || numeroNota == "" || competencia == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados obrigatórios não fornecidos"})
		return
	}

	// Processar arquivo PDF
	file, header, err := c.Request.FormFile("notaFiscal")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo PDF não fornecido"})
		return
	}
	defer file.Close()

	// Verificar se é um PDF
	if filepath.Ext(header.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Apenas arquivos PDF são aceitos"})
		return
	}

	// Ler o conteúdo do arquivo PDF
	pdfBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Erro ao ler arquivo PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler arquivo PDF"})
		return
	}

	// Obter a chave da API OpenAI
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "A variável de ambiente OPENAI_API_KEY não está configurada."})
		return
	}

	// Extrair dados da nota fiscal usando OpenAI
	log.Printf("Iniciando extração de dados da nota fiscal para email: %s", email)
	nfseDataList, err := callOpenAI(pdfBytes, apiKey)
	if err != nil {
		log.Printf("Erro ao extrair dados da nota fiscal: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar nota fiscal: " + err.Error()})
		return
	}

	log.Printf("Dados extraídos da OpenAI: %+v", nfseDataList)

	if len(nfseDataList) == 0 {
		log.Printf("Nenhum dado foi extraído da nota fiscal")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Não foi possível extrair dados da nota fiscal"})
		return
	}

	// Usar o primeiro resultado (ou o que corresponder ao número da nota)
	var notaFiscalExtraida NFSeData
	found := false

	for _, nf := range nfseDataList {
		if nf.NumeroNotaFiscal == numeroNota {
			notaFiscalExtraida = nf
			found = true
			break
		}
	}

	// Se não encontrou correspondência, usar o primeiro
	if !found {
		notaFiscalExtraida = nfseDataList[0]
	}

	// Criar diretório para salvar os arquivos se não existir
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Erro ao criar diretório: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	// Gerar nome único para o arquivo
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.pdf", email, numeroNota, timestamp)
	filePath := filepath.Join(uploadDir, filename)

	// Salvar arquivo PDF
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Erro ao criar arquivo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo"})
		return
	}
	defer dst.Close()

	// Copiar conteúdo do arquivo
	if _, err := dst.Write(pdfBytes); err != nil {
		log.Printf("Erro ao copiar arquivo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo"})
		return
	}

	// Criar registro completo da nota fiscal com dados extraídos
	notaFiscal := NotaFiscalData{
		Email:         email,
		NumeroNota:    notaFiscalExtraida.NumeroNotaFiscal,
		Competencia:   competencia,
		Prestador:     notaFiscalExtraida.PrestadorServicos,
		CNPJ:          notaFiscalExtraida.CNPJ,
		ValorServicos: notaFiscalExtraida.ValorServicos,
		DataNota:      notaFiscalExtraida.DataNotaFiscal,
		ISSRetido:     notaFiscalExtraida.ISSRetido,
	}

	// Salvar dados em JSON (em produção, use um banco de dados)
	notaFiscalJSON, err := json.Marshal(notaFiscal)
	if err != nil {
		log.Printf("Erro ao serializar dados: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	// Salvar dados em arquivo JSON
	jsonFilename := fmt.Sprintf("%s_%s_%s.json", email, numeroNota, timestamp)
	jsonFilepath := filepath.Join(uploadDir, jsonFilename)

	if err := os.WriteFile(jsonFilepath, notaFiscalJSON, 0644); err != nil {
		log.Printf("Erro ao salvar dados JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar dados"})
		return
	}

	log.Printf("Nota fiscal salva: %s", filename)
	log.Printf("Dados extraídos: %+v", notaFiscal)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Nota fiscal salva com sucesso",
		"filename":       filename,
		"data":           notaFiscal,
		"extracted_data": notaFiscalExtraida,
	})
}

// BuscarNotasFiscais busca notas fiscais por competência
func BuscarNotasFiscais(c *gin.Context) {
	competencia := c.Query("competencia")

	log.Printf("Busca de notas fiscais solicitada para competência: %s", competencia)

	if competencia == "" {
		log.Printf("Erro: Competência não fornecida")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Competência é obrigatória"})
		return
	}

	// Criar diretório de uploads se não existir
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Erro ao criar diretório: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	// Listar todos os arquivos JSON no diretório
	files, err := os.ReadDir(uploadDir)
	if err != nil {
		log.Printf("Erro ao ler diretório: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar notas fiscais"})
		return
	}

	log.Printf("Encontrados %d arquivos no diretório uploads", len(files))

	var notasFiscais []NotaFiscalData

	// Filtrar arquivos JSON e ler conteúdo
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			log.Printf("Processando arquivo JSON: %s", file.Name())
			filePath := filepath.Join(uploadDir, file.Name())

			// Ler conteúdo do arquivo JSON
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Erro ao ler arquivo %s: %v", file.Name(), err)
				continue
			}

			log.Printf("Conteúdo do arquivo %s: %s", file.Name(), string(content))

			// Deserializar dados da nota fiscal
			var notaFiscal NotaFiscalData
			if err := json.Unmarshal(content, &notaFiscal); err != nil {
				log.Printf("Erro ao deserializar arquivo %s: %v", file.Name(), err)
				continue
			}

			// Filtrar por competência
			if strings.Contains(notaFiscal.Competencia, competencia) {
				log.Printf("Nota fiscal %s corresponde à competência %s", file.Name(), competencia)
				notasFiscais = append(notasFiscais, notaFiscal)
			}
		}
	}

	log.Printf("Encontradas %d notas fiscais para competência %s", len(notasFiscais), competencia)

	response := gin.H{
		"notas_fiscais": notasFiscais,
		"total":         len(notasFiscais),
		"competencia":   competencia,
	}

	log.Printf("Resposta JSON: %+v", response)

	c.JSON(http.StatusOK, response)
}
