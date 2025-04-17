# Estágio 1: Builder
FROM golang:1.23-alpine AS builder

# Instalar git e ferramentas C para compilação (necessário para algumas dependências Go)
RUN apk update && apk add --no-cache git build-base

WORKDIR /app

# Copiar arquivos de módulo e baixar dependências primeiro para aproveitar o cache do Docker
COPY go.mod go.sum ./
RUN go mod download

# Copiar o restante do código fonte
COPY . .

# Compilar a aplicação para Linux AMD64. Desabilitar CGO para um build estático, se possível.
# O ponto de entrada é cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/lta-results-api ./cmd/api

# Estágio 2: Runner
FROM debian:12-slim

# Instalar dependências do sistema, incluindo fontes, Google Chrome e dnsutils (para nslookup)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    fonts-liberation \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libc6 \
    libcairo2 \
    libcups2 \
    libdbus-1-3 \
    libexpat1 \
    libfontconfig1 \
    libgbm1 \
    libgcc1 \
    libglib2.0-0 \
    libgtk-3-0 \
    libnspr4 \
    libnss3 \
    libpango-1.0-0 \
    libpangocairo-1.0-0 \
    libstdc++6 \
    libx11-6 \
    libx11-xcb1 \
    libxcb1 \
    libxcomposite1 \
    libxcursor1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxi6 \
    libxrandr2 \
    libxrender1 \
    libxss1 \
    libxtst6 \
    lsb-release \
    wget \
    xdg-utils \
    dnsutils \
    && wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
    && apt-get install -y ./google-chrome-stable_current_amd64.deb \
    # Limpar cache do apt
    && rm -rf /var/lib/apt/lists/* \
    && rm google-chrome-stable_current_amd64.deb

WORKDIR /app

# Copiar o executável compilado do estágio builder
COPY --from=builder /app/lta-results-api /app/lta-results-api

# Copiar o arquivo .env (se existir e for necessário em produção - NÃO recomendado para segredos)
# No Render, é melhor usar as variáveis de ambiente do serviço.
# COPY .env .env

# Criar diretório para dados do Chrome ANTES de mudar o usuário
RUN mkdir /app/chrome-data

# Criar um usuário não-root para executar a aplicação
RUN groupadd -r appuser && useradd --no-log-init -r -g appuser appuser
# Garantir que o diretório de trabalho e o diretório de dados do chrome pertençam ao novo usuário
RUN chown -R appuser:appuser /app
USER appuser

# Expor a porta que a aplicação usa (obtida da variável PORT)
# Defina PORT no ambiente do Render (ou localmente via .env/docker-compose)
# Usando 4444 como padrão se PORT não for definido, mas é melhor definir explicitamente.
EXPOSE ${PORT:-4444}

# Comando para iniciar a aplicação
ENTRYPOINT ["/app/lta-results-api"]
