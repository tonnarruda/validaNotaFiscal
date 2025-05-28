# Validador de Notas Fiscais

Este projeto consiste em uma aplicação para validação de notas fiscais contra dados de uma planilha Excel.

## Estrutura do Projeto

- `backend/`: API em Go para processamento das notas fiscais e planilhas
- `frontend/`: Interface web em React para visualização e interação

## Requisitos

### Backend
- Go 1.21+
- [go-excel](https://github.com/360EntSecGroup-Skylar/excelize)

### Frontend
- Node.js 18+
- React 18+
- Yarn ou NPM

## Como executar

### Backend
```bash
cd backend
go mod tidy
go run main.go
```

### Frontend
```bash
cd frontend
yarn install
yarn start
```

A aplicação estará disponível em `http://localhost:3000` # validaNotaFiscal
