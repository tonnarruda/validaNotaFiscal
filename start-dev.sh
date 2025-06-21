#!/bin/bash

# Script para iniciar o ambiente de desenvolvimento
echo "🚀 Iniciando Validador de NF..."

# Verificar se a API key está configurada
if [ -z "$OPENAI_API_KEY" ]; then
    echo "❌ Erro: OPENAI_API_KEY não está configurada"
    echo "   Por favor, defina a variável de ambiente."
    echo ""
    echo "   Exemplo: export OPENAI_API_KEY='sua-chave-aqui'"
    echo "   Para obter uma chave: https://platform.openai.com/api-keys"
    exit 1
fi

echo "✅ Chave da API da OpenAI encontrada."

# Função para limpar processos ao sair
cleanup() {
    echo ""
    echo "🛑 Parando servidores..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null
    exit 0
}

# Capturar Ctrl+C
trap cleanup SIGINT

# Iniciar backend
echo "🔧 Iniciando backend..."
cd backend
go run main.go &
BACKEND_PID=$!
cd ..

# Aguardar backend inicializar
sleep 3

# Iniciar frontend
echo "🎨 Iniciando frontend..."
cd frontend
npm start &
FRONTEND_PID=$!
cd ..

echo ""
echo "✅ Servidores iniciados!"
echo "📱 Frontend: http://localhost:3000"
echo "🔧 Backend: http://localhost:8080"
echo ""
echo "Pressione Ctrl+C para parar os servidores"

# Aguardar indefinidamente
wait 