# NF Decoder AI

Sistema de decodificação de notas fiscais usando inteligência artificial (Google Gemini).

## 🚀 Funcionalidades

- Upload de arquivos PDF e XML de notas fiscais
- Extração automática de informações usando IA
- Interface moderna e responsiva
- Análise estruturada de dados fiscais
- Suporte a múltiplos formatos de arquivo

## 📋 Pré-requisitos

- Go 1.24 ou superior
- Node.js 16 ou superior
- Chave da API da OpenAI

## 🛠️ Instalação

### 1. Clone o repositório
```bash
git clone <url-do-repositorio>
cd nf-decoder-ai
```

### 2. Configure a Chave da API da OpenAI
Crie um arquivo `.env` na raiz do projeto e adicione sua chave:

```
OPENAI_API_KEY="sua-chave-secreta-da-openai-aqui"
```

Ou configure a variável de ambiente diretamente no seu terminal:

```bash
export OPENAI_API_KEY="sua-chave-secreta-da-openai-aqui"
```

Para obter uma chave da API da OpenAI:
1. Acesse [https://platform.openai.com/api-keys](https://platform.openai.com/api-keys)
2. Crie uma nova chave secreta
3. Copie a chave e configure no ambiente

### 3. Backend (Go)

```bash
cd backend
go mod tidy
go run main.go
```

O servidor estará rodando em `http://localhost:8080`

### 4. Frontend (React)

```bash
cd frontend
npm install
npm start
```

O frontend estará rodando em `http://localhost:3000`

## 📁 Estrutura do Projeto

```
nf-decoder-ai/
├── backend/
│   ├── main.go              # Servidor principal
│   ├── go.mod               # Dependências Go
│   └── handlers/
│       └── handler.go       # Handler da API
├── frontend/
│   ├── package.json         # Dependências Node.js
│   ├── public/
│   │   └── index.html       # HTML principal
│   └── src/
│       ├── App.js           # Componente principal
│       ├── index.js         # Ponto de entrada
│       └── components/
│           └── UploadForm.jsx # Formulário de upload
└── README.md
```

## 🔧 Configuração

### Variáveis de Ambiente

- `OPENAI_API_KEY`: Chave da API da OpenAI (obrigatória)

### Endpoints da API

- `POST /upload`: Upload e análise de arquivo de nota fiscal

## 🎯 Como Usar

1. Acesse `http://localhost:3000`
2. Clique na área de upload para selecionar um arquivo PDF ou XML
3. Clique em "Enviar para Análise"
4. Aguarde o processamento pela IA
5. Visualize o resultado estruturado
6. Use o botão "Copiar Resultado" para copiar o texto

## 📊 Informações Extraídas

O sistema extrai automaticamente:

- **Dados do Emitente**: Nome, CNPJ, Endereço
- **Dados do Destinatário**: Nome, CPF/CNPJ, Endereço
- **Dados da NF**: Número, Série, Data, Valor Total
- **Itens**: Descrição, Quantidades, Valores
- **Impostos**: ICMS, PIS, COFINS
- **Informações Adicionais**: Observações, Condições de pagamento

## 🛡️ Segurança

- API Key armazenada em variáveis de ambiente
- Validação de tipos de arquivo
- Limite de tamanho de arquivo (10MB)
- CORS configurado adequadamente

## 🐛 Solução de Problemas

### Erro de API Key
```
Error: A variável de ambiente OPENAI_API_KEY não está configurada.
```
**Solução**: Configure a variável de ambiente `OPENAI_API_KEY`.

### Erro de CORS
```
Access to fetch at 'http://localhost:8080/upload' from origin 'http://localhost:3000' has been blocked by CORS policy
```
**Solução**: Verifique se o backend está rodando e se o CORS está configurado corretamente

### Arquivo muito grande
```
Error: O arquivo deve ter no máximo 10MB
```
**Solução**: Use um arquivo menor ou comprima o arquivo

## 🤝 Contribuição

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## 📝 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## 📞 Suporte

Para suporte, abra uma issue no repositório ou entre em contato através do email. 