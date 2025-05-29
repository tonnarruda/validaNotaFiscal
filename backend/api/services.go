package api

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
	pdfcpu "rsc.io/pdf"
)

var (
	cnpjRegex       = regexp.MustCompile(`InscMunicipal(\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2})`)
	numeroNotaRegex = regexp.MustCompile(`NúmerodaNFS-e\s*(\d+)`)
	valorRegex      = regexp.MustCompile(`\(-\) Desconto Condicionado\s*(\d+(?:\.\d{3})*,\d{2})?`)
)

func ExtractFromPDF(file *os.File) (NotaFiscal, error) {
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

	var cnpj string
	if matches := cnpjRegex.FindStringSubmatch(text); len(matches) > 1 {
		cnpj = strings.TrimSpace(matches[1])
	}

	var numeroNota string
	if matches := numeroNotaRegex.FindStringSubmatch(text); len(matches) > 1 {
		numeroNota = strings.TrimSpace(matches[1])
	}

	if cnpj == "" || numeroNota == "" {
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
			fallbackText = DefragmentText(fallbackText)
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

func ParseExcelNotas(f *os.File) ([]NotaFiscal, []string) {
	notas := []NotaFiscal{}
	errosConversao := []string{}

	xlsx, err := excelize.OpenReader(f)
	if err != nil {
		errosConversao = append(errosConversao, "Erro ao ler arquivo Excel")
		return nil, errosConversao
	}

	rows, err := xlsx.GetRows("Sheet1")
	if err != nil {
		errosConversao = append(errosConversao, "Erro ao ler planilha")
		return nil, errosConversao
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 3 || (strings.TrimSpace(row[0]) == "" && strings.TrimSpace(row[1]) == "" && strings.TrimSpace(row[2]) == "") {
			continue
		}
		valorStr := strings.TrimSpace(row[2])
		if strings.Count(valorStr, ",") == 1 {
			valorStr = strings.ReplaceAll(valorStr, ".", "")  // Remove milhar
			valorStr = strings.ReplaceAll(valorStr, ",", ".") // Troca decimal
		}
		valor, err := strconv.ParseFloat(valorStr, 64)
		if err != nil {
			errosConversao = append(errosConversao, fmt.Sprintf("Linha %d: valor inválido '%s'", i+1, row[2]))
			continue
		}
		notas = append(notas, NotaFiscal{
			CNPJ:       strings.TrimSpace(row[0]),
			NumeroNota: strings.TrimSpace(row[1]),
			Valor:      valor,
		})
	}
	return notas, errosConversao
}
