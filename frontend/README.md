# Frontend - Sistema de Notas Fiscais

Este é o frontend do sistema de validação e envio de notas fiscais.

## Funcionalidades

### 1. Processar Nota Fiscal (`/processa-nota-fiscal`)
- Upload de arquivos PDF de notas fiscais
- Upload de planilha Excel para comparação
- Validação e comparação de dados
- Interface para adicionar notas manualmente
- Visualização de resultados com status de convergência/divergência

### 2. Enviar Nota Fiscal (`/envia-nota-fiscal`)
- Sistema de validação por email em 3 etapas:
  1. **Informar email**: Usuário informa o email para receber token de validação
  2. **Validar token**: Usuário digita o token recebido por email
  3. **Enviar nota**: Usuário preenche dados da nota e anexa o arquivo PDF

## Rotas Disponíveis

- `/processa-nota-fiscal` - Tela de processamento e validação de notas fiscais
- `/envia-nota-fiscal` - Tela de envio de notas fiscais com validação por email
- `/` - Redireciona automaticamente para `/processa-nota-fiscal`

## Tecnologias Utilizadas

- React 18
- React Router DOM
- React Dropzone
- XLSX (para processamento de planilhas Excel)
- Tailwind CSS (para estilização)

## Como Executar

1. Instale as dependências:
```bash
npm install
```

2. Configure a variável de ambiente `REACT_APP_API_URL` no arquivo `.env`:
```
REACT_APP_API_URL=http://localhost:8000
```

3. Execute o projeto:
```bash
npm start
```

O projeto estará disponível em `http://localhost:3000`

## Estrutura de Arquivos

```
src/
├── components/
│   ├── NfValidator.jsx      # Componente de processamento de notas
│   ├── EnviaNotaFiscal.jsx  # Componente de envio de notas
│   └── Navigation.jsx       # Componente de navegação
├── App.js                   # Componente principal com rotas
├── index.js                 # Ponto de entrada da aplicação
└── index.css               # Estilos globais e Tailwind CSS
```

## APIs Utilizadas

### Para Processamento de Notas (`/processa-nota-fiscal`)
- `POST /upload` - Upload de arquivos PDF
- `POST /send-validation-token` - Envio de token de validação
- `POST /validate-token` - Validação do token
- `POST /save-nota-fiscal` - Salvamento da nota fiscal

## Dados da Nota Fiscal

O sistema processa e armazena os seguintes dados:
- **Prestador de Serviços**: Razão Social ou nome do prestador
- **CNPJ (NF)**: CNPJ do prestador de serviços
- **Número da Nota (NF)**: Número da nota fiscal
- **Valor dos Serviços**: Valor dos serviços prestados
- **Data da Nota Fiscal**: Data da emissão (DD/MM/AAAA)
- **Competência da Nota Fiscal**: Período de competência (MM/AAAA)
- **ISS Retido**: Valor do ISS retido 