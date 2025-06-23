import React, { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';

const parseCurrency = (value) => {
  if (typeof value === 'number') return value;
  if (!value || typeof value !== 'string') return 0.0;

  const sanitizedValue = String(value)
    .trim()
    .replace(/R\$\s*/, '');

  const lastComma = sanitizedValue.lastIndexOf(',');
  const lastDot = sanitizedValue.lastIndexOf('.');

  // If comma is the last separator, treat as decimal (Brazilian format)
  if (lastComma > lastDot) {
    const numberString = sanitizedValue.replace(/\./g, '').replace(',', '.');
    return parseFloat(numberString) || 0.0;
  }

  // If dot is the last separator (or no comma), treat dot as decimal
  // and remove commas as thousand separators (Standard/US format)
  if (lastDot > lastComma) {
    const numberString = sanitizedValue.replace(/,/g, '');
    return parseFloat(numberString) || 0.0;
  }

  // No separators, just a number
  return parseFloat(sanitizedValue) || 0.0;
};

const NfValidator = () => {
  const [comparisonData, setComparisonData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [newNf, setNewNf] = useState({ cnpj: '', numero: '', valor: '', prestador: '', issRetido: '' });
  const [searchCompetencia, setSearchCompetencia] = useState('');
  const [searching, setSearching] = useState(false);
  const [spreadsheetData, setSpreadsheetData] = useState(null);

  const formatCurrency = (value) => {
    if (typeof value !== 'number' || isNaN(value)) return 'R$ 0,00';
    return new Intl.NumberFormat('pt-BR', {
      style: 'currency',
      currency: 'BRL',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(value);
  };
  
  const normalizeString = (str) => {
    if (!str) return '';
    return str.toString().replace(/[.\-/]/g, '');
  }

  const handleClearAll = () => {
    setComparisonData([]);
    setSpreadsheetData(null);
    setError('');
  };

  const handleSearchNotas = async () => {
    if (!searchCompetencia) {
      setError('Por favor, informe a competência para buscar.');
      return;
    }

    setSearching(true);
    setError('');

    try {
      const apiUrl = process.env.REACT_APP_API_URL || '';
      const response = await fetch(`${apiUrl}/buscar-notas-fiscais?competencia=${encodeURIComponent(searchCompetencia)}`);

      if (!response.ok) {
        let errMsg = 'Erro ao buscar notas fiscais.';
        try {
          const errData = await response.json();
          errMsg = errData.error || errMsg;
        } catch {
          const text = await response.text();
          errMsg = `Erro ${response.status}: ${text}`;
        }
        throw new Error(errMsg);
      }

      const result = await response.json();
      
      const notasConvertidas = (result.notas_fiscais || []).map(nota => ({
        nfCnpj: nota.cnpj || 'N/A',
        nfPrestador: nota.prestador || 'N/A',
        nfNumero: nota.numeroNota || 'N/A',
        nfValor: nota.valorServicos || 0.0,
        nfIssRetido: nota.issRetido || 0.0,
        plCnpj: 'N/A',
        plNumero: 'N/A',
        plValor: 0.0,
        status: 'Enviada',
        dataNota: nota.dataNota || 'N/A',
        competencia: nota.competencia || 'N/A'
      }));

      setComparisonData(prevData => {
        const notasExistentes = prevData.filter(item => 
          !notasConvertidas.some(nova => nova.nfNumero === item.nfNumero)
        );
        return [...notasConvertidas, ...notasExistentes];
      });
      
      if (result.total === 0) {
        setError(`Nenhuma nota fiscal encontrada para a competência ${searchCompetencia}`);
      } else {
        setError('');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setSearching(false);
    }
  };

  const onPdfDrop = useCallback(async (acceptedFiles) => {
    if (acceptedFiles.length === 0) return;

    setLoading(true);
    setError('');

    const formData = new FormData();
    acceptedFiles.forEach(file => {
      formData.append('files', file);
    });

    try {
      const apiUrl = process.env.REACT_APP_API_URL || '';
      const response = await fetch(`${apiUrl}/upload`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const errData = await response.json();
        throw new Error(errData.error || 'Erro no servidor ao processar PDFs.');
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      const processStream = async () => {
        while (true) {
          const { done, value } = await reader.read();
          if (done) {
            setLoading(false);
            break;
          }

          buffer += decoder.decode(value, { stream: true });
          const parts = buffer.split('\n---\n');
          
          parts.slice(0, -1).forEach(part => {
            if (part.trim() === '') return;
            try {
              const nf = JSON.parse(part);
              const newNfData = {
                cnpj: nf['CNPJ (NF)'],
                prestador: nf['Prestador de Serviços'],
                numero: nf['Número da Nota (NF)'],
                valor: parseCurrency(nf['Valor Líquido da Nota Fiscal']),
                issRetido: parseCurrency(nf['ISS Retido']),
              };

              setComparisonData(prevData => {
                const key = `${normalizeString(newNfData.cnpj)}-${normalizeString(newNfData.numero)}`;
                
                const matchIndex = prevData.findIndex(item => 
                    item.status === 'Aguardando PDF' &&
                    `${normalizeString(item.plCnpj)}-${normalizeString(item.plNumero)}` === key
                );

                if (matchIndex !== -1) {
                  const updatedData = [...prevData];
                  const existingItem = updatedData[matchIndex];
                  
                  existingItem.nfCnpj = newNfData.cnpj;
                  if (newNfData.prestador && newNfData.prestador !== 'N/A') {
                    existingItem.nfPrestador = newNfData.prestador;
                  }
                  existingItem.nfNumero = newNfData.numero;
                  existingItem.nfValor = newNfData.valor;
                  existingItem.nfIssRetido = newNfData.issRetido;

                  const nfValor = existingItem.nfValor;
                  const plValor = existingItem.plValor;
                  existingItem.status = Math.abs(nfValor - plValor) < 0.01 ? 'Validada' : 'Divergente';

                  return updatedData;
                } else {
                  const pdfExists = prevData.some(item => `${normalizeString(item.nfCnpj)}-${normalizeString(item.nfNumero)}` === key);
                  if (!pdfExists) {
                    const newRow = {
                      nfCnpj: newNfData.cnpj,
                      nfPrestador: newNfData.prestador,
                      nfNumero: newNfData.numero,
                      nfValor: newNfData.valor,
                      nfIssRetido: newNfData.issRetido,
                      plCnpj: 'N/A',
                      plNumero: 'N/A',
                      plValor: 0.0,
                      status: 'Importada',
                    };
                    return [...prevData, newRow];
                  }
                  return prevData;
                }
              });
            } catch (e) {
              console.error("Failed to parse JSON chunk", e);
            }
          });

          buffer = parts[parts.length - 1];
        }
      };

      processStream();

    } catch (err) {
      setError(err.message);
      setLoading(false);
    }
  }, [comparisonData]);

  const onSpreadsheetDrop = useCallback(async (acceptedFiles) => {
    if (acceptedFiles.length === 0) return;

    const file = acceptedFiles[0];
    setLoading(true);
    setError('');

    const formData = new FormData();
    formData.append('file', file);

    try {
      const apiUrl = process.env.REACT_APP_API_URL || '';
      const response = await fetch(`${apiUrl}/process-spreadsheet`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const errData = await response.json();
        throw new Error(errData.error || 'Erro ao processar planilha.');
      }

      const result = await response.json();

      if (result.success) {
        setSpreadsheetData(result.data);
        console.log('JSON da planilha:', JSON.stringify(result.data, null, 2));
        compareWithSpreadsheet(result.data);
        setError('');
      } else {
        throw new Error(result.error || 'Erro ao processar planilha');
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [comparisonData]);

  const compareWithSpreadsheet = (spreadsheetData) => {
    if (!spreadsheetData || !spreadsheetData.rows) return;

    const lowerCaseHeaders = spreadsheetData.headers.map(h => String(h).toLowerCase());
    const cnpjIndex = lowerCaseHeaders.findIndex(h => h.includes('cnpj'));
    const nfNumeroIndex = lowerCaseHeaders.findIndex(h => h.includes('nota') || h.includes('numero'));
    const valorIndex = lowerCaseHeaders.findIndex(h => h.includes('valor'));
    const prestadorIndex = lowerCaseHeaders.findIndex(h => h.includes('prestador') || h.includes('razão social') || h.includes('razao social') || h.includes('nome'));

    if (cnpjIndex === -1 || nfNumeroIndex === -1 || valorIndex === -1) {
      setError("A planilha deve conter colunas com 'CNPJ', 'Nota'/'Número' e 'Valor'.");
      return;
    }

    const excelDataMap = new Map();
    spreadsheetData.rows.forEach(row => {
      const originalHeaders = spreadsheetData.headers;
      const key = `${normalizeString(row[originalHeaders[cnpjIndex]])}-${normalizeString(row[originalHeaders[nfNumeroIndex]])}`;
      excelDataMap.set(key, {
        cnpj: row[originalHeaders[cnpjIndex]],
        numero: row[originalHeaders[nfNumeroIndex]],
        valor: parseCurrency(String(row[originalHeaders[valorIndex]])),
        prestador: prestadorIndex !== -1 ? row[originalHeaders[prestadorIndex]] : 'N/A',
      });
    });

    const updatedComparisonData = [...comparisonData];

    updatedComparisonData.forEach(item => {
      if (item.status === 'Importada' || item.status === 'Enviada') {
        const key = `${normalizeString(item.nfCnpj)}-${normalizeString(item.nfNumero)}`;
        const match = excelDataMap.get(key);

        if (match) {
          item.plCnpj = match.cnpj;
          item.plNumero = match.numero;
          item.plValor = match.valor;
          if(item.nfPrestador === 'N/A') item.nfPrestador = match.prestador;
          
          const nfValor = item.nfValor;
          const plValor = match.valor;
          item.status = Math.abs(nfValor - plValor) < 0.01 ? 'Validada' : 'Divergente';
          excelDataMap.delete(key);
        }
      }
    });

    excelDataMap.forEach((value, key) => {
      updatedComparisonData.push({
        nfCnpj: 'N/A',
        nfPrestador: value.prestador,
        nfNumero: 'N/A',
        nfValor: 0.0,
        nfIssRetido: 0.0,
        plCnpj: value.cnpj,
        plNumero: value.numero,
        plValor: value.valor,
        status: 'Aguardando PDF',
      });
    });

    setComparisonData(updatedComparisonData);
  };

  const { getRootProps: getPdfRootProps, getInputProps: getPdfInputProps } = useDropzone({
    onDrop: onPdfDrop,
    accept: { 'application/pdf': ['.pdf'] },
    multiple: true
  });
  
  const { getRootProps: getSpreadsheetRootProps, getInputProps: getSpreadsheetInputProps } = useDropzone({
    onDrop: onSpreadsheetDrop,
    accept: { 
      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': ['.xlsx'], 
      'application/vnd.ms-excel': ['.xls'],
      'text/csv': ['.csv']
    },
    multiple: false
  });

  const getStatusClass = (status) => {
    switch (status) {
      case 'Validada': return 'bg-green-100 text-green-800';
      case 'Divergente': return 'bg-red-100 text-red-800';
      case 'Enviada': return 'bg-blue-100 text-blue-800';
      case 'Importada': return 'bg-yellow-100 text-yellow-800';
      case 'Aguardando PDF': return 'bg-purple-100 text-purple-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="bg-white p-8 rounded-lg shadow-lg">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Validador de Notas Fiscais</h1>
        <div className="flex gap-4">
          <button 
            onClick={handleClearAll}
            className="bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded transition-colors"
          >
            Limpar Tudo
          </button>
        </div>
      </div>
      
      {loading && (
        <div className="fixed top-0 left-0 w-full h-full bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="text-white text-xl">Aguarde enquanto processamos os arquivos...</div>
        </div>
      )}

      {error && <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">{error}</div>}

      <div className="mb-6 flex gap-4 items-end">
        <div className="flex-1 max-w-xs">
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Buscar Notas Enviadas por Competência
          </label>
          <input
            type="text"
            value={searchCompetencia}
            onChange={(e) => setSearchCompetencia(e.target.value)}
            placeholder="Ex: 06/2025"
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            disabled={searching}
          />
        </div>
        <button
          onClick={handleSearchNotas}
          disabled={searching || !searchCompetencia}
          className="bg-blue-500 text-white py-3 px-6 rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed font-semibold"
        >
          {searching ? 'Buscando...' : 'Buscar'}
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        <div {...getPdfRootProps()} className="border-2 border-dashed border-gray-300 rounded-lg p-6 text-center cursor-pointer hover:border-blue-500 bg-gray-50 transition-colors">
          <input {...getPdfInputProps()} />
          <p className="text-blue-600 font-semibold">IMPORTAR NOTAS FISCAIS (PDF)</p>
          <p className="text-sm text-gray-500 mt-1">Arraste e solte os PDFs aqui, ou clique para selecionar</p>
        </div>
        <div {...getSpreadsheetRootProps()} className="border-2 border-dashed border-gray-300 rounded-lg p-6 text-center cursor-pointer hover:border-purple-500 bg-gray-50 transition-colors">
          <input {...getSpreadsheetInputProps()} />
          <p className="text-purple-600 font-semibold">IMPORTAR PLANILHA (Excel/CSV)</p>
          <p className="text-sm text-gray-500 mt-1">Arraste e solte Excel ou CSV aqui, ou clique para selecionar</p>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full bg-white">
          <thead className="bg-gray-200">
            <tr>
              <th className="py-3 px-4 text-left text-sm font-semibold text-gray-600">Prestador de Serviços</th>
              <th className="py-3 px-4 text-left text-sm font-semibold text-gray-600">CNPJ (NF)</th>
              <th className="py-3 px-4 text-left text-sm font-semibold text-gray-600">Número da Nota (NF)</th>
              <th className="py-3 px-4 text-left text-sm font-semibold text-gray-600">Valor Líquido</th>
              <th className="py-3 px-4 text-left text-sm font-semibold text-gray-600">ISS Retido</th>
              <th className="py-3 px-4 text-left text-sm font-semibold text-gray-600">CNPJ (Planilha)</th>
              <th className="py-3 px-4 text-left text-sm font-semibold text-gray-600">Número da Nota (Planilha)</th>
              <th className="py-3 px-4 text-left text-sm font-semibold text-gray-600">Valor (Planilha)</th>
              <th className="py-3 px-4 text-center text-sm font-semibold text-gray-600">Status</th>
            </tr>
          </thead>
          <tbody>
            
            {comparisonData.length > 0 ? (
              comparisonData.map((item, index) => (
                <tr key={index} className="border-b hover:bg-gray-50">
                  <td className="py-3 px-4">{item.nfPrestador}</td>
                  <td className="py-3 px-4">{item.nfCnpj}</td>
                  <td className="py-3 px-4">{item.nfNumero}</td>
                  <td className="py-3 px-4">{formatCurrency(item.nfValor)}</td>
                  <td className="py-3 px-4">{formatCurrency(item.nfIssRetido)}</td>
                  <td className="py-3 px-4">{item.plCnpj}</td>
                  <td className="py-3 px-4">{item.plNumero}</td>
                  <td className="py-3 px-4">{formatCurrency(item.plValor)}</td>
                  <td className="py-3 px-4 text-center">
                    <span className={`px-3 py-1 text-xs font-medium rounded-full ${getStatusClass(item.status)}`}>
                      {item.status}
                    </span>
                  </td>
                </tr>
              ))
            ) : (
                <tr>
                  <td colSpan="9" className="text-center py-10 text-gray-500">
                    Aguardando importação dos arquivos...
                  </td>
                </tr>
            )}
          </tbody>
        </table>
      </div>
      {comparisonData.length > 0 && 
        <div className="flex justify-end items-center mt-4 text-sm text-gray-600">
            {`Mostrando 1-${comparisonData.length} de ${comparisonData.length}`}
        </div>
      }
    </div>
  );
};

export default NfValidator; 