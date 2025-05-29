package api

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
