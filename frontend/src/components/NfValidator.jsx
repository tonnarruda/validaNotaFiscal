import React, { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import * as XLSX from 'xlsx';

const NfValidator = () => {
  const [comparisonData, setComparisonData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [newNf, setNewNf] = useState({ cnpj: '', numero: '', valor: '', prestador: '', issRetido: '' });
  const [searchCompetencia, setSearchCompetencia] = useState('');
  const [searching, setSearching] = useState(false);

  const formatCurrency = (value) => {
    if (typeof value !== 'number') return 'R$ 0,00';
    return value.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
  };
  
  const normalizeString = (str) => {
    if (!str) return '';
    return str.toString().replace(/[.\-/]/g, '');
  }

  const handleClearAll = () => {
    setComparisonData([]);
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
      console.log('Buscando notas fiscais para competência:', searchCompetencia);
      console.log('URL da API:', `${apiUrl}/buscar-notas-fiscais?competencia=${encodeURIComponent(searchCompetencia)}`);
      
      const response = await fetch(`${apiUrl}/buscar-notas-fiscais?competencia=${encodeURIComponent(searchCompetencia)}`);

      console.log('Status da resposta:', response.status);
      console.log('Headers da resposta:', response.headers);

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

      const text = await response.text();
      console.log('Resposta bruta:', text);

      let result;
      try {
        result = JSON.parse(text);
      } catch (parseError) {
        console.error('Erro ao fazer parse do JSON:', parseError);
        throw new Error(`Resposta inválida do servidor: ${text.substring(0, 100)}...`);
      }

      console.log('Resultado parseado:', result);
      
      // Converter notas enviadas para o formato da tabela de comparação
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

      // Adicionar às notas existentes na tabela
      setComparisonData(prevData => {
        // Remover notas duplicadas baseado no número da nota
        const notasExistentes = prevData.filter(item => 
          !notasConvertidas.some(nova => nova.nfNumero === item.nfNumero)
        );
        return [...notasConvertidas, ...notasExistentes];
      });
      
      if (result.total === 0) {
        setError(`Nenhuma nota fiscal encontrada para a competência ${searchCompetencia}`);
      } else {
        setError(''); // Limpar erros anteriores
      }
    } catch (err) {
      console.error('Erro na busca:', err);
      setError(err.message);
    } finally {
      setSearching(false);
    }
  };

  const handleNewNfChange = (e) => {
    const { name, value } = e.target;
    setNewNf(prev => ({ ...prev, [name]: value }));
  };

  const handleSaveNewNf = () => {
    if (!newNf.cnpj || !newNf.numero || !newNf.valor || !newNf.prestador) {
      setError("Preencha todos os campos para salvar a nota.");
      return;
    }
    const newRow = {
      nfCnpj: newNf.cnpj,
      nfPrestador: newNf.prestador,
      nfNumero: newNf.numero,
      nfValor: parseFloat(newNf.valor.replace(',', '.')) || 0.0,
      nfIssRetido: parseFloat(newNf.issRetido.replace(',', '.')) || 0.0,
      plCnpj: 'N/A',
      plNumero: 'N/A',
      plValor: 0.0,
      status: 'Importada',
    };
    setComparisonData(prevData => [newRow, ...prevData]);
    setNewNf({ cnpj: '', numero: '', valor: '', prestador: '', issRetido: '' });
    setError('');
  };

  const handleCancelNewNf = () => {
    setNewNf({ cnpj: '', numero: '', valor: '', prestador: '', issRetido: '' });
    setError('');
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
              const newNf = {
                nfCnpj: nf['CNPJ (NF)'],
                nfPrestador: nf['Prestador de Serviços'],
                nfNumero: nf['Número da Nota (NF)'],
                nfValor: nf['Valor Líquido da Nota Fiscal'],
                nfIssRetido: nf['ISS Retido'],
                plCnpj: 'N/A',
                plNumero: 'N/A',
                plValor: 0.0,
                status: 'Importada',
              };

              setComparisonData(prevData => {
                const key = `${normalizeString(newNf.nfCnpj)}-${normalizeString(newNf.nfNumero)}`;
                const exists = prevData.some(item => `${normalizeString(item.nfCnpj)}-${normalizeString(item.nfNumero)}` === key);
                if (!exists) {
                  return [...prevData, newNf];
                }
                return prevData;
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
  }, []);

  const onExcelDrop = useCallback((acceptedFiles) => {
    if (acceptedFiles.length === 0 || comparisonData.length === 0) {
        if (comparisonData.length === 0) {
            setError("Por favor, importe os PDFs primeiro.");
        }
        return;
    }

    const file = acceptedFiles[0];
    const reader = new FileReader();

    reader.onload = (event) => {
        try {
            const bstr = event.target.result;
            const wb = XLSX.read(bstr, { type: 'binary' });
            const wsname = wb.SheetNames[0];
            const ws = wb.Sheets[wsname];
            const data = XLSX.utils.sheet_to_json(ws, { header: 1 });
            
            const header = data[0];
            const cnpjIndex = header.findIndex(h => h.toLowerCase().includes('cnpj'));
            const nfNumeroIndex = header.findIndex(h => h.toLowerCase().includes('nota'));
            const valorIndex = header.findIndex(h => h.toLowerCase().includes('valor'));

            if(cnpjIndex === -1 || nfNumeroIndex === -1 || valorIndex === -1) {
                setError("A planilha Excel deve conter as colunas 'CNPJ', 'Nota' e 'Valor'.");
                return;
            }

            const excelDataMap = new Map();
            data.slice(1).forEach(row => {
                const key = `${normalizeString(row[cnpjIndex])}-${normalizeString(row[nfNumeroIndex])}`;
                excelDataMap.set(key, {
                    cnpj: row[cnpjIndex],
                    numero: row[nfNumeroIndex],
                    valor: parseFloat(row[valorIndex]) || 0.0,
                });
            });

            const updatedComparisonData = comparisonData.map(item => {
                const key = `${normalizeString(item.nfCnpj)}-${normalizeString(item.nfNumero)}`;
                const match = excelDataMap.get(key);

                if (match) {
                    const nfValor = parseFloat(item.nfValor);
                    const plValor = match.valor;
                    const status = Math.abs(nfValor - plValor) < 0.01 ? 'Convergente' : 'Divergente';

                    return { ...item, plCnpj: match.cnpj, plNumero: match.numero, plValor: plValor, status: status };
                }
                return item;
            });

            setComparisonData(updatedComparisonData);

        } catch (err) {
            setError("Erro ao processar o arquivo Excel.");
        }
    };
    reader.readAsBinaryString(file);
  }, [comparisonData]);

  const { getRootProps: getPdfRootProps, getInputProps: getPdfInputProps } = useDropzone({
    onDrop: onPdfDrop,
    accept: { 'application/pdf': ['.pdf'] },
    multiple: true
  });
  
  const { getRootProps: getExcelRootProps, getInputProps: getExcelInputProps } = useDropzone({
    onDrop: onExcelDrop,
    accept: { 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': ['.xlsx'], 'application/vnd.ms-excel': ['.xls'] },
    multiple: false
  });

  const getStatusClass = (status) => {
    switch (status) {
      case 'Validada': return 'bg-green-100 text-green-800';
      case 'Divergente': return 'bg-red-100 text-red-800';
      case 'Enviada': return 'bg-blue-100 text-blue-800';
      case 'Importada': return 'bg-yellow-100 text-yellow-800';
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
            <div className="text-white text-xl">Aguarde enquanto processamos as notas fiscais...</div>
        </div>
      )}

      {error && <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded relative mb-4" role="alert">{error}</div>}

      {/* Campo de Busca de Notas Enviadas */}
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
        <div {...getExcelRootProps()} className="border-2 border-dashed border-gray-300 rounded-lg p-6 text-center cursor-pointer hover:border-purple-500 bg-gray-50 transition-colors">
          <input {...getExcelInputProps()} />
          <p className="text-purple-600 font-semibold">IMPORTAR PLANILHA EXCEL</p>
          <p className="text-sm text-gray-500 mt-1">Arraste e solte o Excel aqui, ou clique para selecionar</p>
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
                  <td colSpan="7" className="text-center py-10 text-gray-500">
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