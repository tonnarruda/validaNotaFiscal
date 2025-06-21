import React, { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';

const EnviaNotaFiscal = () => {
  const [email, setEmail] = useState('');
  const [numeroNota, setNumeroNota] = useState('');
  const [competencia, setCompetencia] = useState('');
  const [notaFiscal, setNotaFiscal] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [extractedData, setExtractedData] = useState(null);

  const onDrop = useCallback((acceptedFiles) => {
    if (acceptedFiles.length > 0) {
      setNotaFiscal(acceptedFiles[0]);
      setError('');
    }
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      'application/pdf': ['.pdf']
    },
    multiple: false
  });

  const handleSubmitNota = async () => {
    if (!email || !email.includes('@')) {
      setError('Por favor, informe um email válido.');
      return;
    }

    if (!numeroNota || !competencia || !notaFiscal) {
      setError('Por favor, preencha todos os campos e anexe a nota fiscal.');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const formData = new FormData();
      formData.append('email', email);
      formData.append('numeroNota', numeroNota);
      formData.append('competencia', competencia);
      formData.append('notaFiscal', notaFiscal);

      const apiUrl = process.env.REACT_APP_API_URL || '';
      const response = await fetch(`${apiUrl}/save-nota-fiscal`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const errData = await response.json();
        throw new Error(errData.error || 'Erro ao salvar nota fiscal.');
      }

      const result = await response.json();
      setSuccess('Nota fiscal enviada e salva com sucesso!');
      setExtractedData(result.extracted_data);
      
      // Reset form
      setEmail('');
      setNumeroNota('');
      setCompetencia('');
      setNotaFiscal(null);
      setExtractedData(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const formatCompetencia = (value) => {
    // Formato MM/AAAA
    const cleaned = value.replace(/\D/g, '');
    if (cleaned.length <= 2) {
      return cleaned;
    }
    return `${cleaned.slice(0, 2)}/${cleaned.slice(2, 6)}`;
  };

  const handleCompetenciaChange = (e) => {
    const formatted = formatCompetencia(e.target.value);
    setCompetencia(formatted);
  };

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-white rounded-lg shadow-lg p-8">
        <h1 className="text-3xl font-bold text-gray-800 mb-8 text-center">
          Enviar Nota Fiscal
        </h1>

        {/* Error and Success Messages */}
        {error && (
          <div className="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
            {error}
          </div>
        )}

        {success && (
          <div className="mb-4 p-4 bg-green-100 border border-green-400 text-green-700 rounded">
            {success}
          </div>
        )}

        {/* Form */}
        <div className="space-y-6">
          {/* Email */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="seu@email.com"
              disabled={loading}
            />
          </div>

          {/* Número da Nota */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Número da Nota Fiscal
            </label>
            <input
              type="text"
              value={numeroNota}
              onChange={(e) => setNumeroNota(e.target.value)}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="123456"
              disabled={loading}
            />
          </div>

          {/* Competência */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Competência (MM/AAAA)
            </label>
            <input
              type="text"
              value={competencia}
              onChange={handleCompetenciaChange}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="06/2025"
              maxLength="7"
              disabled={loading}
            />
          </div>

          {/* Upload PDF */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Nota Fiscal (PDF)
            </label>
            <div
              {...getRootProps()}
              className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors ${
                isDragActive
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-300 hover:border-blue-400'
              }`}
            >
              <input {...getInputProps()} />
              {notaFiscal ? (
                <div>
                  <p className="text-green-600 font-semibold">✓ Arquivo selecionado</p>
                  <p className="text-sm text-gray-500">{notaFiscal.name}</p>
                </div>
              ) : (
                <div>
                  <p className="text-blue-600 font-semibold">
                    {isDragActive ? 'Solte o arquivo aqui' : 'Clique ou arraste o PDF aqui'}
                  </p>
                  <p className="text-sm text-gray-500 mt-1">
                    Apenas arquivos PDF são aceitos
                  </p>
                </div>
              )}
            </div>
          </div>

          {/* Submit Button */}
          <button
            onClick={handleSubmitNota}
            disabled={loading}
            className="w-full bg-blue-500 text-white py-3 px-6 rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed font-semibold"
          >
            {loading ? 'Enviando...' : 'Enviar Nota Fiscal'}
          </button>
        </div>

        {/* Extracted Data Display */}
        {extractedData && (
          <div className="mt-8 p-6 bg-gray-50 rounded-lg">
            <h3 className="text-lg font-semibold text-gray-800 mb-4">
              Dados Extraídos da Nota Fiscal
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <span className="font-medium">Prestador:</span> {extractedData.PrestadorServicos}
              </div>
              <div>
                <span className="font-medium">CNPJ:</span> {extractedData.CNPJ}
              </div>
              <div>
                <span className="font-medium">Número da Nota:</span> {extractedData.NumeroNotaFiscal}
              </div>
              <div>
                <span className="font-medium">Data:</span> {extractedData.DataNotaFiscal}
              </div>
              <div>
                <span className="font-medium">Valor dos Serviços:</span> R$ {extractedData.ValorServicos}
              </div>
              <div>
                <span className="font-medium">ISS Retido:</span> R$ {extractedData.ISSRetido}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default EnviaNotaFiscal; 