import React, { useState } from 'react';
import './App.css';

interface NotaFiscal {
  id?: string;
  cnpj: string;
  numeroNota: string;
  valor: number;
}

interface ComparisonResult {
  notaFiscal: NotaFiscal;
  dadosPlanilha: NotaFiscal;
  match: boolean;
}

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

function formatCurrency(valor: number) {
  return valor.toLocaleString('pt-BR', { 
    style: 'currency', 
    currency: 'BRL',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  });
}

function App() {
  const [pdfFiles, setPdfFiles] = useState<File[]>([]);
  const [excelFile, setExcelFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notasImportadas, setNotasImportadas] = useState<NotaFiscal[]>([]);
  const [notasPlanilha, setNotasPlanilha] = useState<NotaFiscal[]>([]);
  const [comparacoes, setComparacoes] = useState<ComparisonResult[]>([]);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  const handlePdfChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files) {
      setPdfFiles(Array.from(event.target.files));
      setError(null);
    }
  };

  const handleExcelChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      setExcelFile(event.target.files[0]);
      setError(null);
    }
  };

  const compararNotas = (notas: NotaFiscal[], planilha: NotaFiscal[]) => {
    const comparacoes: ComparisonResult[] = [];
    const planilhaMap = new Map(planilha.map(nota => [nota.numeroNota, nota]));
    notas.forEach(notaFiscal => {
      const notaPlanilha = planilhaMap.get(notaFiscal.numeroNota);
      if (notaPlanilha) {
        comparacoes.push({
          notaFiscal,
          dadosPlanilha: notaPlanilha,
          match: Math.abs(notaFiscal.valor - notaPlanilha.valor) < 0.01
        });
      }
    });
    return comparacoes.sort((a, b) => Number(a.notaFiscal.numeroNota) - Number(b.notaFiscal.numeroNota));
  };

  const handleUpload = async () => {
    if (pdfFiles.length === 0 && !excelFile) {
      setError('Selecione um ou mais arquivos PDF ou Excel');
      return;
    }
    setLoading(true);
    setError(null);
    setSuccessMsg(null);

    // Upload múltiplo de PDFs para /api/notas
    if (pdfFiles.length > 0) {
      let novasNotas: NotaFiscal[] = [];
      for (const file of pdfFiles) {
        const formData = new FormData();
        formData.append('file', file);
        try {
          const response = await fetch(`${API_URL}/api/notas`, {
            method: 'POST',
            body: formData,
          });
          if (!response.ok) {
            throw new Error('Erro ao processar arquivo');
          }
          const data = await response.json();
          if (data.nota) {
            novasNotas.push(data.nota);
          }
        } catch (err) {
          setError(err instanceof Error ? err.message : 'Erro ao processar arquivo');
        }
      }
      if (novasNotas.length > 0) {
        setNotasImportadas(prev => {
          // Remove duplicadas pelo numeroNota e adiciona as novas
          const semDuplicadas = prev.filter(
            n => !novasNotas.some(nv => nv.numeroNota === n.numeroNota)
          );
          const todas = [...semDuplicadas, ...novasNotas];
          return todas.sort((a, b) => Number(a.numeroNota) - Number(b.numeroNota));
        });
        setSuccessMsg('Notas fiscais importadas com sucesso!');
      }
    }

    // Upload Excel para /api/planilha
    if (excelFile) {
      const formData = new FormData();
      formData.append('file', excelFile);
      try {
        const response = await fetch(`${API_URL}/api/planilha`, {
          method: 'POST',
          body: formData,
        });
        if (!response.ok) {
          throw new Error('Erro ao processar planilha');
        }
        const data = await response.json();
        if (Array.isArray(data)) {
          setNotasPlanilha(data.sort((a, b) => Number(a.numeroNota) - Number(b.numeroNota)));
          setSuccessMsg('Planilha importada com sucesso!');
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Erro ao processar planilha');
      }
    }
    setLoading(false);
  };

  // Atualizar comparações sempre que notas ou planilha mudam
  React.useEffect(() => {
    console.log('Notas importadas:', notasImportadas);
    console.log('Notas planilha:', notasPlanilha);
    if (notasImportadas.length > 0 && notasPlanilha.length > 0) {
      const novasComparacoes = compararNotas(notasImportadas, notasPlanilha);
      setComparacoes(novasComparacoes);
    }
  }, [notasImportadas, notasPlanilha]);

  return (
    <div className="App">
      <div className="container">
        <h1>Validador de Notas Fiscais</h1>
        <div className="upload-section">
          <div>
            <input
              id="pdf-upload"
              type="file"
              accept=".pdf"
              style={{ display: 'none' }}
              onChange={handlePdfChange}
              multiple
            />
            <label htmlFor="pdf-upload">
              <button
                className="upload-btn-pdf"
                type="button"
                onClick={() => document.getElementById('pdf-upload')?.click()}
              >
                IMPORTAR NOTAS FISCAIS (PDF)
              </button>
            </label>
            {pdfFiles.length > 0 && (
              <div className="file-info">
                {pdfFiles.length} arquivo(s) selecionado(s)
              </div>
            )}
          </div>
          <div>
            <input
              id="excel-upload"
              type="file"
              accept=".xlsx,.xls"
              style={{ display: 'none' }}
              onChange={handleExcelChange}
            />
            <label htmlFor="excel-upload">
              <button
                className="upload-btn-excel"
                type="button"
                onClick={() => document.getElementById('excel-upload')?.click()}
              >
                IMPORTAR PLANILHA EXCEL
              </button>
            </label>
            {excelFile && (
              <div className="file-info">
                Arquivo: {excelFile.name}
              </div>
            )}
          </div>
          <button
            className="upload-btn-final"
            onClick={handleUpload}
            disabled={loading || (pdfFiles.length === 0 && !excelFile)}
          >
            {loading ? 'Processando...' : 'Validar'}
          </button>
        </div>
        {error && <div className="error-message">{error}</div>}
        {successMsg && <div className="success-message">{successMsg}</div>}

        {/* Tabela de Comparação */}
        {comparacoes.length > 0 && (
          <div className="results-section">
            <h2>Resultado da Validação</h2>
            <table className="nf-table">
              <thead>
                <tr>
                  <th>CNPJ (NF)</th>
                  <th>Número da Nota</th>
                  <th>Valor (NF)</th>
                  <th>Valor (Planilha)</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {comparacoes.map((comp, idx) => (
                  <tr key={idx} className={comp.match ? 'row-match' : 'row-mismatch'}>
                    <td>{comp.notaFiscal.cnpj}</td>
                    <td>{comp.notaFiscal.numeroNota}</td>
                    <td>{formatCurrency(comp.notaFiscal.valor)}</td>
                    <td>{formatCurrency(comp.dadosPlanilha.valor)}</td>
                    <td>
                      <span className={`status-badge ${comp.match ? 'status-ok' : 'status-divergent'}`}>
                        {comp.match ? 'OK' : 'Divergente'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {/* Mensagens de aviso se algum lado estiver vazio */}
        {notasImportadas.length === 0 && (
          <div style={{marginTop: '2rem', color: '#888', fontSize: '1.1rem'}}>Importe ao menos um PDF de nota fiscal.</div>
        )}
        {notasPlanilha.length === 0 && (
          <div style={{marginTop: '2rem', color: '#888', fontSize: '1.1rem'}}>Importe a planilha Excel para comparar.</div>
        )}

        {/* Notas sem comparação na planilha */}
        {notasImportadas.length > 0 && notasPlanilha.length > 0 && (
          <div className="results-section">
            <h2>Notas não encontradas na planilha</h2>
            <table className="nf-table">
              <thead>
                <tr>
                  <th>CNPJ</th>
                  <th>Número da Nota</th>
                  <th>Valor</th>
                </tr>
              </thead>
              <tbody>
                {notasImportadas
                  .filter(nota => !notasPlanilha.some(p => p.numeroNota === nota.numeroNota))
                  .map((nota, idx) => (
                    <tr key={idx}>
                      <td>{nota.cnpj}</td>
                      <td>{nota.numeroNota}</td>
                      <td>{formatCurrency(nota.valor)}</td>
                    </tr>
                  ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
