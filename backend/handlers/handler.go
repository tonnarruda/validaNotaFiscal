package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// NFSeData struct holds the extracted data from the invoice.
type NFSeData struct {
	CNPJ                   string  `json:"CNPJ (NF)"`
	NumeroNotaFiscal       string  `json:"Número da Nota (NF)"`
	ValorServicos          float64 `json:"Valor dos Serviços"`
	ValorLiquidoNotaFiscal float64 `json:"Valor Líquido da Nota Fiscal"`
	DataNotaFiscal         string  `json:"Data da Nota Fiscal"`
	CompetenciaNotaFiscal  string  `json:"Competência da Nota Fiscal"`
	PrestadorServicos      string  `json:"Prestador de Serviços"`
	ISSRetido              float64 `json:"ISS Retido"`
}

// Structs for OpenAI API
type OpenAIRequest struct {
	Model     string          `json:"model"`
	Messages  []OpenAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

type OpenAIMessage struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type MessageContent struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// callOpenAI sends the invoice image to OpenAI API for processing.
func callOpenAI(pdfBytes []byte, apiKey string) ([]NFSeData, error) {
	var nfseDataList []NFSeData

	// Create a temporary file for the PDF
	tmpPdfFile, err := os.CreateTemp("", "invoice-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp pdf file: %v", err)
	}
	defer os.Remove(tmpPdfFile.Name())

	if _, err := tmpPdfFile.Write(pdfBytes); err != nil {
		return nil, fmt.Errorf("failed to write to temp pdf file: %v", err)
	}
	tmpPdfFile.Close()

	// Convert PDF to image using pdftoppm (from poppler-utils)
	// We'll just process the first page.
	outputImagePath := strings.TrimSuffix(tmpPdfFile.Name(), ".pdf")
	cmd := exec.Command("pdftoppm", "-png", "-f", "1", "-l", "1", tmpPdfFile.Name(), outputImagePath)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to convert pdf to image: %v. Make sure poppler-utils is installed", err)
	}

	imageFilePath := outputImagePath + "-1.png"
	defer os.Remove(imageFilePath)

	// Read the image file
	imageBytes, err := os.ReadFile(imageFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %v", err)
	}

	// Encode the image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)
	imageURL := fmt.Sprintf("data:image/png;base64,%s", base64Image)

	systemPrompt := `Você é um especialista em extração de dados de Notas Fiscais de Serviço Eletrônicas (NFS-e) de diferentes prefeituras do Brasil. Sua tarefa é analisar a imagem de uma nota fiscal e retornar **APENAS** um JSON válido com a seguinte estrutura:

{
  "Prestador de Serviços": "Razão Social ou nome do prestador",
  "CNPJ (NF)": "CNPJ do prestador de serviços",
  "Número da Nota (NF)": "número da nota fiscal",
  "Valor dos Serviços": 0.0,
  "Data da Nota Fiscal": "DD/MM/AAAA",
  "Competência da Nota Fiscal": "MM/AAAA",
  "ISS Retido": 0.0
}

### INSTRUÇÕES OBRIGATÓRIAS:

1. **FOCO EXCLUSIVO NO PRESTADOR**: Todos os dados de identificação (Prestador de Serviços, CNPJ) devem ser **exclusivamente** do **PRESTADOR DE SERVIÇOS**. É o erro mais crítico a ser evitado.

2. **PROCESSO DE EXTRAÇÃO**:
   - **PASSO 1: LOCALIZAR O BLOCO DO PRESTADOR**: Antes de extrair qualquer dado, encontre a seção da nota fiscal intitulada **"DADOS DO PRESTADOR DE SERVIÇOS"** ou "EMITENTE".
   - **PASSO 2: EXTRAIR DADOS DO BLOCO**: Todos os campos a seguir devem ser extraídos **APENAS DE DENTRO DESTE BLOCO**.
   - **IGNORE COMPLETAMENTE O TOMADOR**: Qualquer informação na seção "DADOS DO TOMADOR DE SERVIÇOS" deve ser ignorada.

3. **Prestador de Serviços**:
   - Dentro do bloco do **PRESTADOR**, encontre e extraia a "Razão Social/Nome".

4. **CNPJ (NF)**:
   - Dentro do mesmo bloco do **PRESTADOR**, encontre e extraia o "CPF/CNPJ".

5. **Número da Nota (NF)**:
   - Busque por "Número da NFS-e" ou "Número da Nota Fiscal". Priorize o número da NFS-e.

6. **Valor dos Serviços**:
   - Use o campo **"Valor do Serviço"** ou **"Valor Total"**.
   - O número deve ser puro (sem aspas e sem R$), ex: 2380.89.

7. **Data da Nota Fiscal**:
   - Extraia do campo "Data de Emissão" ou similar. Use o formato DD/MM/AAAA.

8. **Competência da Nota Fiscal**:
   - Busque pelo campo "Competência". Se não existir, use o mês/ano da data de emissão.

9. **ISS Retido**:
   - Busque por "ISS Retido" ou "(-) ISS Retido". Se não houver, o valor é 0.

10. **Se algum campo não for encontrado**:
    - Use string vazia "" (exceto para campos de valor, que devem ser 0).

11. **Se houver mais de uma nota fiscal no mesmo texto**, retorne um array com um objeto JSON para cada uma.`

	userPrompt := "Extraia os dados da imagem desta nota fiscal e retorne apenas o JSON."

	reqBody := OpenAIRequest{
		Model: "gpt-4o",
		Messages: []OpenAIMessage{
			{
				Role: "system",
				Content: []interface{}{
					MessageContent{Type: "text", Text: systemPrompt},
				},
			},
			{
				Role: "user",
				Content: []interface{}{
					MessageContent{Type: "text", Text: userPrompt},
					MessageContent{Type: "image_url", ImageURL: &ImageURL{URL: imageURL, Detail: "high"}},
				},
			},
		},
		MaxTokens: 3000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nfseDataList, fmt.Errorf("erro ao criar JSON para OpenAI: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nfseDataList, fmt.Errorf("erro ao criar requisição para OpenAI: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nfseDataList, fmt.Errorf("erro ao chamar a API OpenAI: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nfseDataList, fmt.Errorf("erro ao ler resposta da OpenAI: %v", err)
	}

	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		return nfseDataList, fmt.Errorf("erro da API OpenAI (status %d): %s", resp.StatusCode, string(respBody))
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nfseDataList, fmt.Errorf("erro ao decodificar resposta da OpenAI: %v. Resposta: %s", err, string(respBody))
	}

	if openAIResp.Error != nil {
		return nfseDataList, fmt.Errorf("erro da API OpenAI: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return nfseDataList, fmt.Errorf("resposta da OpenAI vazia")
	}

	// Limpar o conteúdo para garantir que seja um JSON válido
	jsonContent := strings.TrimSpace(openAIResp.Choices[0].Message.Content)

	// Remove markdown code blocks if present
	if strings.HasPrefix(jsonContent, "```json") {
		jsonContent = strings.TrimPrefix(jsonContent, "```json")
		jsonContent = strings.TrimSuffix(jsonContent, "```")
	} else if strings.HasPrefix(jsonContent, "```") {
		jsonContent = strings.TrimPrefix(jsonContent, "```")
		jsonContent = strings.TrimSuffix(jsonContent, "```")
	}

	// Clean up any remaining whitespace
	jsonContent = strings.TrimSpace(jsonContent)

	// Handle both single object and array of objects
	if strings.HasPrefix(jsonContent, "[") {
		// Response is a JSON array
		if err := json.Unmarshal([]byte(jsonContent), &nfseDataList); err != nil {
			return nil, fmt.Errorf("erro ao fazer unmarshal do array JSON da OpenAI: %v. Resposta: %s", err, jsonContent)
		}
	} else if strings.HasPrefix(jsonContent, "{") {
		// Response is a single JSON object
		var singleNfseData NFSeData
		if err := json.Unmarshal([]byte(jsonContent), &singleNfseData); err != nil {
			// Fallback for malformed single object
			startIdx := strings.Index(jsonContent, "{")
			endIdx := strings.LastIndex(jsonContent, "}")
			if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
				jsonSubstring := jsonContent[startIdx : endIdx+1]
				if err := json.Unmarshal([]byte(jsonSubstring), &singleNfseData); err != nil {
					return nil, fmt.Errorf("erro ao fazer unmarshal do JSON da OpenAI (substring): %v. Resposta: %s", err, jsonContent)
				}
			} else {
				return nil, fmt.Errorf("erro ao fazer unmarshal do JSON da OpenAI: %v. Resposta: %s", err, jsonContent)
			}
		}
		nfseDataList = append(nfseDataList, singleNfseData)
	} else {
		return nil, fmt.Errorf("formato de resposta inesperado da OpenAI: não é JSON nem array. Resposta: %s", jsonContent)
	}

	// Calculate the net value
	for i := range nfseDataList {
		nfseDataList[i].ValorLiquidoNotaFiscal = nfseDataList[i].ValorServicos - nfseDataList[i].ISSRetido
	}

	// Validate that we have at least some data
	if len(nfseDataList) > 0 && nfseDataList[0].NumeroNotaFiscal == "" && nfseDataList[0].ValorServicos == 0 {
		// Fallback cannot be text-based anymore.
		// For now, we'll just log if the response seems empty.
		log.Printf("Warning: OpenAI response for an image seems empty or invalid: %+v", nfseDataList)
	}

	return nfseDataList, nil
}

// DecodeNotaFiscal handles multi-file upload and processing using a streaming response.
func DecodeNotaFiscal(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	if c.Request.Method == "OPTIONS" {
		c.Status(http.StatusOK)
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "A variável de ambiente OPENAI_API_KEY não está configurada."})
		return
	}

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar formulário."})
		return
	}

	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum arquivo enviado."})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		log.Println("Streaming unsupported!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening file %s: %v", fileHeader.Filename, err)
			continue
		}

		content, err := io.ReadAll(file)
		file.Close() // Close file immediately after reading
		if err != nil {
			log.Printf("Error reading file %s: %v", fileHeader.Filename, err)
			continue
		}

		nfseDataList, err := callOpenAI(content, apiKey)
		if err != nil {
			log.Printf("Error processing %s with OpenAI: %v", fileHeader.Filename, err)
			continue
		}

		for _, nfseData := range nfseDataList {
			jsonData, err := json.Marshal(nfseData)
			if err != nil {
				log.Printf("Error marshalling NFSe data: %v", err)
				continue
			}
			// Use a separator to distinguish between JSON objects
			fmt.Fprintf(c.Writer, "%s\n---\n", jsonData)
			flusher.Flush()
		}
	}
}
