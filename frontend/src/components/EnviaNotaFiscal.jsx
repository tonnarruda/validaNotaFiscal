import React, { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';

const EnviaNotaFiscal = () => {
  const [step, setStep] = useState(1); // 1: email, 2: token, 3: dados da nota
  const [email, setEmail] = useState('');
  const [token, setToken] = useState('');
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

  const handleSendEmail = async () => {
    if (!email || !email.includes('@')) {
      setError('Por favor, informe um email válido.');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const apiUrl = process.env.REACT_APP_API_URL || '';
      console.log('API URL:', apiUrl);
      console.log('URL completa:', `${apiUrl}/send-validation-token`);
      console.log('Email sendo enviado:', email);
      const response = await fetch(`${apiUrl}/send-validation-token`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email }),
      });

      if (!response.ok) {
        let errMsg = 'Erro ao enviar token de validação.';
        const text = await response.text();
        try {
          const errData = JSON.parse(text);
          errMsg = errData.error || errMsg;
        } catch {
          errMsg = text;
        }
        throw new Error(errMsg);
      }

      setSuccess('Token de validação enviado para seu email. Verifique sua caixa de entrada.');
      setStep(2);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleValidateToken = async () => {
    if (!token || token.length < 4) {
      setError('Por favor, informe o token de validação.');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const apiUrl = process.env.REACT_APP_API_URL || '';
      const response = await fetch(`${apiUrl}/validate-token`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, token }),
      });

      if (!response.ok) {
        const errData = await response.json();
        throw new Error(errData.error || 'Token inválido.');
      }

      setSuccess('Token validado com sucesso!');
      setStep(3);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmitNota = async () => {
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
      setToken('');
      setNumeroNota('');
      setCompetencia('');
      setNotaFiscal(null);
      setExtractedData(null);
      setStep(1);
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

        {/* Progress Steps */}
        <div className="flex justify-center mb-8">
          <div className="flex items-center">
            <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-semibold ${
              step >= 1 ? 'bg-blue-500 text-white' : 'bg-gray-300 text-gray-600'
            }`}>
              1
            </div>
            <div className={`w-16 h-1 ${step >= 2 ? 'bg-blue-500' : 'bg-gray-300'}`}></div>
            <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-semibold ${
              step >= 2 ? 'bg-blue-500 text-white' : 'bg-gray-300 text-gray-600'
            }`}>
              2
            </div>
            <div className={`w-16 h-1 ${step >= 3 ? 'bg-blue-500' : 'bg-gray-300'}`}></div>
            <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-semibold ${
              step >= 3 ? 'bg-blue-500 text-white' : 'bg-gray-300 text-gray-600'
            }`}>
              3
            </div>
          </div>
        </div>

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

        {/* Dados Extraídos */}
        {extractedData && (
          <div className="mb-6 p-6 bg-blue-50 border border-blue-200 rounded-lg">
            <h3 className="text-lg font-semibold text-blue-800 mb-4">
              Dados Extraídos da Nota Fiscal
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-blue-700">Prestador de Serviços</label>
                <p className="text-blue-900">{extractedData['Prestador de Serviços'] || 'Não encontrado'}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-blue-700">CNPJ</label>
                <p className="text-blue-900">{extractedData['CNPJ (NF)'] || 'Não encontrado'}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-blue-700">Número da Nota</label>
                <p className="text-blue-900">{extractedData['Número da Nota (NF)'] || 'Não encontrado'}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-blue-700">Valor dos Serviços</label>
                <p className="text-blue-900">
                  {extractedData['Valor dos Serviços'] 
                    ? `R$ ${extractedData['Valor dos Serviços'].toFixed(2).replace('.', ',')}` 
                    : 'Não encontrado'}
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-blue-700">Data da Nota</label>
                <p className="text-blue-900">{extractedData['Data da Nota Fiscal'] || 'Não encontrado'}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-blue-700">Competência</label>
                <p className="text-blue-900">{extractedData['Competência da Nota Fiscal'] || 'Não encontrado'}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-blue-700">ISS Retido</label>
                <p className="text-blue-900">
                  {extractedData['ISS Retido'] 
                    ? `R$ ${extractedData['ISS Retido'].toFixed(2).replace('.', ',')}` 
                    : 'R$ 0,00'}
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-blue-700">Valor Líquido</label>
                <p className="text-blue-900">
                  {extractedData['Valor Líquido da Nota Fiscal'] 
                    ? `R$ ${extractedData['Valor Líquido da Nota Fiscal'].toFixed(2).replace('.', ',')}` 
                    : 'Não encontrado'}
                </p>
              </div>
            </div>
            <button
              onClick={() => setExtractedData(null)}
              className="mt-4 bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 font-semibold"
            >
              Fechar
            </button>
          </div>
        )}

        {/* Step 1: Email */}
        {step === 1 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-700 mb-4">
                Passo 1: Informe seu email
              </h2>
              <p className="text-gray-600 mb-4">
                Digite seu email para receber um token de validação.
              </p>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="seu@email.com"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={loading}
              />
            </div>
            <button
              onClick={handleSendEmail}
              disabled={loading}
              className="w-full bg-blue-500 text-white py-3 px-6 rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed font-semibold"
            >
              {loading ? 'Enviando...' : 'Enviar Token de Validação'}
            </button>
          </div>
        )}

        {/* Step 2: Token Validation */}
        {step === 2 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-700 mb-4">
                Passo 2: Valide o token
              </h2>
              <p className="text-gray-600 mb-4">
                Digite o token enviado para {email}
              </p>
              <input
                type="text"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="Digite o token"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={loading}
              />
            </div>
            <div className="flex space-x-4">
              <button
                onClick={() => setStep(1)}
                className="flex-1 bg-gray-500 text-white py-3 px-6 rounded-lg hover:bg-gray-600 font-semibold"
              >
                Voltar
              </button>
              <button
                onClick={handleValidateToken}
                disabled={loading}
                className="flex-1 bg-blue-500 text-white py-3 px-6 rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed font-semibold"
              >
                {loading ? 'Validando...' : 'Validar Token'}
              </button>
            </div>
          </div>
        )}

        {/* Step 3: Nota Fiscal Data */}
        {step === 3 && (
          <div className="space-y-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-700 mb-4">
                Passo 3: Informações da Nota Fiscal
              </h2>
              <p className="text-gray-600 mb-4">
                Preencha os dados da nota fiscal e anexe o arquivo PDF.
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Número da Nota Fiscal
                </label>
                <input
                  type="text"
                  value={numeroNota}
                  onChange={(e) => setNumeroNota(e.target.value)}
                  placeholder="Número da nota"
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  disabled={loading}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Competência (MM/AAAA)
                </label>
                <input
                  type="text"
                  value={competencia}
                  onChange={handleCompetenciaChange}
                  placeholder="MM/AAAA"
                  maxLength="7"
                  className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  disabled={loading}
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Anexar Nota Fiscal (PDF)
              </label>
              <div
                {...getRootProps()}
                className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors ${
                  isDragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400'
                }`}
              >
                <input {...getInputProps()} />
                {notaFiscal ? (
                  <div className="text-green-600">
                    <p className="font-semibold">✓ Arquivo selecionado</p>
                    <p className="text-sm">{notaFiscal.name}</p>
                  </div>
                ) : (
                  <div>
                    <p className="text-gray-600">
                      {isDragActive
                        ? 'Solte o arquivo aqui...'
                        : 'Arraste e solte um arquivo PDF aqui, ou clique para selecionar'}
                    </p>
                  </div>
                )}
              </div>
            </div>

            <div className="flex space-x-4">
              <button
                onClick={() => setStep(2)}
                className="flex-1 bg-gray-500 text-white py-3 px-6 rounded-lg hover:bg-gray-600 font-semibold"
              >
                Voltar
              </button>
              <button
                onClick={handleSubmitNota}
                disabled={loading || !notaFiscal}
                className="flex-1 bg-green-500 text-white py-3 px-6 rounded-lg hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed font-semibold"
              >
                {loading ? 'Enviando...' : 'Enviar Nota Fiscal'}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default EnviaNotaFiscal; 