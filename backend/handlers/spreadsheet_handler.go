package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// SpreadsheetData representa os dados da planilha
type SpreadsheetData struct {
	Headers []string                 `json:"headers"`
	Rows    []map[string]interface{} `json:"rows"`
	Total   int                      `json:"total"`
}

// ProcessSpreadsheet processa arquivos Excel (.xlsx, .xls) e CSV
func ProcessSpreadsheet(c *gin.Context) {
	// Obter o arquivo do formulário
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Erro ao obter arquivo: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Verificar extensão do arquivo
	ext := strings.ToLower(filepath.Ext(header.Filename))

	var data SpreadsheetData

	switch ext {
	case ".xlsx", ".xls":
		data, err = processExcelFile(file)
	case ".csv":
		data, err = processCSVFile(file)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Formato de arquivo não suportado. Use .xlsx, .xls ou .csv",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao processar arquivo: " + err.Error(),
		})
		return
	}

	// Retornar os dados como JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"message": fmt.Sprintf("Arquivo processado com sucesso. %d linhas encontradas.", data.Total),
	})
}

// processExcelFile processa arquivos Excel
func processExcelFile(file io.Reader) (SpreadsheetData, error) {
	var data SpreadsheetData

	// Ler o arquivo Excel
	xlFile, err := excelize.OpenReader(file)
	if err != nil {
		return data, fmt.Errorf("erro ao abrir arquivo Excel: %v", err)
	}
	defer xlFile.Close()

	// Obter a primeira planilha
	sheets := xlFile.GetSheetList()
	if len(sheets) == 0 {
		return data, fmt.Errorf("nenhuma planilha encontrada no arquivo")
	}

	sheetName := sheets[0]
	rows, err := xlFile.GetRows(sheetName)
	if err != nil {
		return data, fmt.Errorf("erro ao ler linhas da planilha: %v", err)
	}

	if len(rows) == 0 {
		return data, fmt.Errorf("planilha vazia")
	}

	// Primeira linha como cabeçalhos
	headers := rows[0]
	data.Headers = headers

	// Processar linhas de dados
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		rowData := make(map[string]interface{})

		// Mapear valores para cabeçalhos
		for j, value := range row {
			if j < len(headers) {
				rowData[headers[j]] = value
			}
		}

		// Preencher valores vazios com string vazia
		for _, header := range headers {
			if _, exists := rowData[header]; !exists {
				rowData[header] = ""
			}
		}

		data.Rows = append(data.Rows, rowData)
	}

	data.Total = len(data.Rows)
	return data, nil
}

// processCSVFile processa arquivos CSV
func processCSVFile(file io.Reader) (SpreadsheetData, error) {
	var data SpreadsheetData

	// Ler o arquivo CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return data, fmt.Errorf("erro ao ler arquivo CSV: %v", err)
	}

	if len(records) == 0 {
		return data, fmt.Errorf("arquivo CSV vazio")
	}

	// Primeira linha como cabeçalhos
	headers := records[0]
	data.Headers = headers

	// Processar linhas de dados
	for i := 1; i < len(records); i++ {
		record := records[i]
		rowData := make(map[string]interface{})

		// Mapear valores para cabeçalhos
		for j, value := range record {
			if j < len(headers) {
				rowData[headers[j]] = value
			}
		}

		// Preencher valores vazios com string vazia
		for _, header := range headers {
			if _, exists := rowData[header]; !exists {
				rowData[header] = ""
			}
		}

		data.Rows = append(data.Rows, rowData)
	}

	data.Total = len(data.Rows)
	return data, nil
}

// GetSpreadsheetPreview retorna uma prévia dos dados da planilha (primeiras 10 linhas)
func GetSpreadsheetPreview(c *gin.Context) {
	// Obter o arquivo do formulário
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Erro ao obter arquivo: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Verificar extensão do arquivo
	ext := strings.ToLower(filepath.Ext(header.Filename))

	var data SpreadsheetData

	switch ext {
	case ".xlsx", ".xls":
		data, err = processExcelFile(file)
	case ".csv":
		data, err = processCSVFile(file)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Formato de arquivo não suportado. Use .xlsx, .xls ou .csv",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao processar arquivo: " + err.Error(),
		})
		return
	}

	// Limitar a 10 linhas para prévia
	previewRows := data.Rows
	if len(previewRows) > 10 {
		previewRows = data.Rows[:10]
	}

	previewData := SpreadsheetData{
		Headers: data.Headers,
		Rows:    previewRows,
		Total:   data.Total,
	}

	// Retornar a prévia como JSON
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    previewData,
		"message": fmt.Sprintf("Prévia gerada. Total de %d linhas no arquivo.", data.Total),
	})
}
