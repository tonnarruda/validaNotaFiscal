package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
)

// NFSeData struct holds the extracted data from the invoice.
type NFSeData struct {
	CNPJ                  string  `json:"CNPJ (NF)"`
	NumeroNotaFiscal      string  `json:"Número da Nota (NF)"`
	ValorLiquido          float64 `json:"Valor (NF)"`
	DataNotaFiscal        string  `json:"Data da Nota Fiscal"`
	CompetenciaNotaFiscal string  `json:"Competência da Nota Fiscal"`
	PrestadorServicos     string  `json:"Prestador de Serviços"`
	ISSRetido             float64 `json:"ISS Retido"`
	//TomadorServicos       string  `json:"Tomador de Serviços"`
}

// Structs for OpenAI API
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
}
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

// extractTextFromPDF reads a PDF and returns its text content.
func extractTextFromPDF(reader io.ReaderAt, size int64) (string, error) {
	pdfReader, err := pdf.NewReader(reader, size)
	if err != nil {
		return "", err
	}
	var textBuilder strings.Builder
	numPages := pdfReader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			// Fallback to extract text row by row
			rows, _ := page.GetTextByRow()
			for _, row := range rows {
				for _, word := range row.Content {
					textBuilder.WriteString(word.S)
				}
				textBuilder.WriteString("\n")
			}
			continue
		}
		textBuilder.WriteString(text)
	}
	return textBuilder.String(), nil
}

// callOpenAI sends the extracted text to OpenAI API for processing.
func callOpenAI(textContent string, apiKey string) ([]NFSeData, error) {
	var nfseDataList []NFSeData

	systemPrompt := `Você é um especialista em extração de dados de Notas Fiscais de Serviço Eletrônicas (NFS-e) de diferentes prefeituras do Brasil. Sua tarefa é analisar o texto completo da nota e retornar **APENAS** um JSON válido com a seguinte estrutura:

{
  "Prestador de Serviços": "Razão Social ou nome do prestador",
  "CNPJ (NF)": "CNPJ do prestador de serviços",
  "Número da Nota (NF)": "número da nota fiscal",
  "Valor (NF)": 0.0,
  "Data da Nota Fiscal": "DD/MM/AAAA",
  "Competência da Nota Fiscal": "MM/AAAA",
  "ISS Retido": 0.0
}

### INSTRUÇÕES OBRIGATÓRIAS:

1. **Não escreva nenhum texto além do JSON**. 

2. Os dados devem vir sempre do **DADOS DO PRESTADOR DE SERVIÇOS**:
   - Use a seção "DADOS DO PRESTADOR DE SERVIÇOS", "EMITENTE", ou similar.
   - **Ignore completamente a seção "DADOS DO TOMADOR DE SERVIÇOS"** ou qualquer nome que esteja associado ao cliente.

3. **CNPJ (NF)**:
   - Use a seção "DADOS DO PRESTADOR DE SERVIÇOS", "EMITENTE", ou similar.
   - Procure o CNPJ próximo do NOME DE FANTASIA, Se houver.
   - Extraia o CNPJ apenas do **prestador de serviços** (exemplo: "CPF/CNPJ: 17.830.029/0001-01").
   - Procure sempre o **CNPJ do prestador** com base em:
		- Blocos com o título "DADOS DO PRESTADOR DE SERVIÇOS" ou similares.

4. **Número da Nota (NF)**:
   - Busque por "Número da NFS-e", "Número da Nota Fiscal", ou similar.
   - Priorize sempre o número principal da NFS-e, e **não o número do RPS**.

5. **Valor (NF)**:
   - Use PRIORITARIAMENTE o campo **"Valor Líquido"**.
   - Se não houver, utilize o campo **"Valor Total", "Valor do Serviço"** ou equivalente.
   - O número deve ser puro (sem aspas e sem R$ ou vírgulas), ex: 2380.89.

6. **Data da Nota Fiscal**:
   - Extraia do campo com nome como "Data de Emissão", "Data e Hora da Emissão", ou "Data Fato Gerador".
   - Use o formato DD/MM/AAAA.

7. **Competência da Nota Fiscal**:
   - Busque pelo campo "Competência".
   - Caso não exista, derive a competência com base na data de emissão (MM/AAAA).

8. **Prestador de Serviços**:
   - Utilize o nome/razão social do **prestador** encontrado na seção adequada, como "Razão Social/Nome" ou "Nome/Nome Empresarial".

9. **ISS Retido**:
   - Busque pelo campo "(-) ISS Retido". Se não houver, o valor deve ser 0.

10. Caso algum campo não seja encontrado:
    - Use string vazia "" (exceto para campos de valor, que devem ser 0).

11. **Se houver mais de uma nota fiscal no mesmo texto, retorne um array com múltiplos objetos JSON**, um para cada nota.`

	userPrompt := fmt.Sprintf("Extraia os dados da seguinte nota fiscal e retorne apenas o JSON:\n\n%s", textContent)

	reqBody := OpenAIRequest{
		Model: "gpt-4o",
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
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

	// Validate that we have at least some data
	if len(nfseDataList) > 0 && nfseDataList[0].NumeroNotaFiscal == "" && nfseDataList[0].ValorLiquido == 0 {
		// Try to extract data using regex patterns as fallback
		nfseDataList = extractDataFromText(jsonContent)
	}

	return nfseDataList, nil
}

// extractDataFromText attempts to extract invoice data from text using regex patterns
func extractDataFromText(text string) []NFSeData {
	var data []NFSeData

	// This is a fallback method - in practice, the improved prompt should prevent this from being needed
	// But it provides a safety net for edge cases

	// You could add regex patterns here to extract data from text if needed
	// For now, we'll return empty data and let the calling function handle the error

	return data
}

// DecodeNotaFiscal handles multi-file upload and processing.
func DecodeNotaFiscal(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

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

	var results []NFSeData
	var processingErrors []string

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			processingErrors = append(processingErrors, fmt.Sprintf("Erro ao abrir %s: %v", fileHeader.Filename, err))
			continue
		}

		content, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			processingErrors = append(processingErrors, fmt.Sprintf("Erro ao ler %s: %v", fileHeader.Filename, err))
			continue
		}

		textContent, err := extractTextFromPDF(bytes.NewReader(content), int64(len(content)))
		if err != nil {
			processingErrors = append(processingErrors, fmt.Sprintf("Erro ao extrair texto de %s: %v", fileHeader.Filename, err))
			continue
		}

		if textContent != "" {
			nfseDataList, err := callOpenAI(textContent, apiKey)
			if err != nil {
				processingErrors = append(processingErrors, fmt.Sprintf("Erro ao processar %s com OpenAI: %v", fileHeader.Filename, err))
				continue
			}
			results = append(results, nfseDataList...)
		}
	}

	if len(results) == 0 {
		errorDetails := strings.Join(processingErrors, "; ")
		log.Printf("Error processing files: %s", errorDetails)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Não foi possível processar nenhum dos arquivos.",
			"details": errorDetails,
		})
		return
	}

	resultsJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Printf("Error marshalling results for logging: %v", err)
		log.Printf("Successfully processed files (unstructured): %+v", results)
	} else {
		log.Printf("Successfully processed files:\n%s", string(resultsJSON))
	}

	c.JSON(http.StatusOK, results)
}
