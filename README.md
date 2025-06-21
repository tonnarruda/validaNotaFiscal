# Sistema de Validação e Envio de Notas Fiscais

Sistema completo para processamento, validação e envio de notas fiscais com integração de IA para extração automática de dados.

## 🚀 Funcionalidades

### 1. Processar Nota Fiscal (`/processa-nota-fiscal`)
- **Upload de PDFs**: Processamento de múltiplos arquivos PDF de notas fiscais
- **Upload de Excel**: Comparação com planilhas Excel para validação
- **Extração automática**: Dados extraídos automaticamente usando IA (OpenAI GPT-4)
- **Busca de notas enviadas**: Filtro por competência para visualizar notas já enviadas
- **Validação cruzada**: Comparação entre dados do PDF e planilha Excel
- **Interface responsiva**: Design moderno com Tailwind CSS

### 2. Enviar Nota Fiscal (`/envia-nota-fiscal`)
- **Validação por email**: Sistema de 3 etapas com token de validação
- **Upload de PDF**: Anexo de arquivo PDF da nota fiscal
- **Extração automática**: Dados extraídos automaticamente usando IA
- **Salvamento completo**: Todos os dados são salvos no servidor
- **Visualização dos dados**: Interface mostra todos os dados extraídos

## 📋 Dados Extraídos Automaticamente

O sistema extrai automaticamente os seguintes dados das notas fiscais:

- **Prestador de Serviços** (Razão Social)
- **CNPJ** do prestador
- **Número da Nota Fiscal**
- **Valor dos Serviços**
- **Data da Nota Fiscal**
- **Competência**
- **ISS Retido**
- **Valor Líquido** (calculado automaticamente)

## 🔄 Fluxo de Trabalho

### Processamento de Notas
1. Upload de PDFs → Extração automática de dados
2. Upload de Excel → Comparação e validação
3. Busca por competência → Visualização de notas enviadas
4. Análise de convergência/divergência

### Envio de Notas
1. Informar email → Token enviado
2. Validar token → Acesso liberado
3. Preencher dados + anexar PDF → Envio
4. Dados extraídos automaticamente → Salvamento
5. Visualização dos dados extraídos

## 🛠️ Tecnologias

### Frontend
- **React 18** - Interface de usuário
- **React Router DOM** - Navegação entre páginas
- **React Dropzone** - Upload de arquivos
- **XLSX** - Processamento de planilhas Excel
- **Tailwind CSS** - Estilização moderna

### Backend
- **Go** - Servidor backend
- **Gin** - Framework web
- **OpenAI GPT-4** - Extração de dados de PDFs
- **Poppler-utils** - Conversão PDF para imagem

## 🚀 Como Executar

### 1. Configuração do Ambiente

Crie um arquivo `.env` na raiz do projeto:
```bash
# Configurações do Backend
OPENAI_API_KEY=sua_chave_openai_aqui

# Configurações do Frontend
REACT_APP_API_URL=http://localhost:8080
```

### 2. Instalação de Dependências

**Backend:**
```bash
cd backend
go mod download
```

**Frontend:**
```bash
cd frontend
npm install
```

### 3. Instalação do Poppler-utils (necessário para conversão PDF)

**macOS:**
```bash
brew install poppler
```

**Ubuntu/Debian:**
```bash
sudo apt-get install poppler-utils
```

### 4. Execução

**Backend:**
```bash
cd backend
go run main.go
```

**Frontend:**
```bash
cd frontend
npm start
```

## 📡 Endpoints da API

### Processamento de Notas
- `POST /upload` - Upload e processamento de PDFs
- `GET /buscar-notas-fiscais?competencia=MM/AAAA` - Busca notas por competência

### Envio de Notas
- `POST /send-validation-token` - Envio de token de validação
- `POST /validate-token` - Validação do token
- `POST /save-nota-fiscal` - Salvamento da nota fiscal

## 📁 Estrutura do Projeto

```
validaNotaFiscal/
├── backend/
│   ├── handlers/
│   │   ├── handler.go          # Processamento de PDFs
│   │   └── email_handlers.go   # Envio e busca de notas
│   ├── main.go                 # Servidor principal
│   └── uploads/                # Arquivos salvos
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── NfValidator.jsx     # Processamento de notas
│   │   │   ├── EnviaNotaFiscal.jsx # Envio de notas
│   │   │   └── Navigation.jsx      # Navegação
│   │   └── App.js              # Rotas principais
│   └── package.json
└── README.md
```

## 🔧 Configurações Importantes

### OpenAI API Key
- Obtenha sua chave em: https://platform.openai.com/api-keys
- Configure no arquivo `.env`

### Portas
- **Backend**: 8080
- **Frontend**: 3000

### CORS
- Configurado automaticamente no backend
- Aceita requisições de qualquer origem em desenvolvimento

## 📊 Funcionalidades de Busca

### Busca por Competência
- Filtro por período (MM/AAAA)
- Visualização em tabela organizada
- Contador de resultados
- Botão para limpar busca

### Dados Exibidos na Busca
- Email do remetente
- Número da nota fiscal
- Prestador de serviços
- CNPJ
- Valores (serviços e ISS)
- Data da nota

## 🎯 Casos de Uso

1. **Contabilidade**: Processamento em lote de notas fiscais
2. **Empresas**: Validação de notas recebidas
3. **Auditoria**: Comparação entre sistemas
4. **Compliance**: Verificação de dados fiscais

## 🔒 Segurança

- Validação por email obrigatória
- Tokens com expiração (10 minutos)
- Validação de tipos de arquivo (apenas PDF)
- Sanitização de dados de entrada

## 📈 Próximas Melhorias

- [ ] Banco de dados para persistência
- [ ] Sistema de usuários e permissões
- [ ] Relatórios e exportação
- [ ] Notificações por email
- [ ] API para integração com outros sistemas 