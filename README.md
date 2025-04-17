# 🏆 LTA Match Results API

<div align="center">
  <img src="https://ltafantasy.com/public/lta-fantasy-logo.svg" alt="LTA API Logo" width="300">
  <br>
  <h3>Uma API poderosa para obter resultados e estatísticas de jogos da LTA</h3>
  <p>
    
[![Go](https://github.com/Bulletdev/lta-results-api/actions/workflows/go.yml/badge.svg)](https://github.com/Bulletdev/lta-results-api/actions/workflows/go.yml)

  <img src="https://img.shields.io/badge/MongoDB-4.4+-47A248?style=for-the-badge&logo=mongodb&logoColor=white" alt="MongoDB">
    <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge" alt="License: MIT">
  </p>
</div> 
 
<br>

## 📋 Índice

- [Funcionalidades](#-funcionalidades)
- [Tecnologias](#-tecnologias)
- [Instalação](#-instalação)
- [Configuração](#-configuração)
- [Endpoints da API](#-endpoints-da-api)
- [Troubleshooting](#-troubleshooting)
- [Segurança](#-segurança)
- [Como Contribuir](#-como-contribuir)
- [Licença](#-licença)

<br>

##  Sobre

A **LTA Match Results API** é uma solução completa para extrair, armazenar e disponibilizar dados sobre os resultados de partidas da Liga de League of Legends (LTA).
Utilizando técnicas de web scraping, a API coleta automaticamente informações de jogos das regiões Sul e Norte, fornecendo dados detalhados sobre desempenho de equipes e jogadores.

Ideal para:
- Sites de estatísticas e análises de e-sports
- Aplicativos de fantasy league
- Dashboards para times e casters
- Integrações com plataformas de análise de jogos

<br>

## ✨ Funcionalidades

### ⚙️ Core
- **Extração Automática**: Coleta dados de partidas diariamente
- **API RESTful**: Endpoints intuitivos para consulta de dados
- **Filtros Flexíveis**: Consulta por região, time, jogador ou período
- **Paginação**: Controle sobre o volume de dados retornados

### 📊 Dados Disponíveis
- **Resultados de Partidas**: Placar, vencedor, data, duração
- **Estatísticas de Jogadores**: KDA, farm, participação em abates
- **Estatísticas de Times**: Winrate, desempenho por lado, campeões mais jogados
- **Histórico de Confrontos**: Performance histórica entre equipes

### 🛠️ Administração
- **Painel Admin**: Interface para gerenciar dados manualmente
- **Autenticação Segura**: Proteção de rotas sensíveis
- **Logs Detalhados**: Acompanhamento de operações do sistema

<br>

## 🔧 Tecnologias

### Backend
- [Go](https://golang.org/) - Linguagem de programação eficiente
- [Gin](https://github.com/gin-gonic/gin) - Framework web rápido
- [ChromeDP](https://github.com/chromedp/chromedp) - Automação de navegador
- [MongoDB](https://www.mongodb.com/) - Banco de dados NoSQL

### Ferramentas
- [Docker](https://www.docker.com/) - Containerização
- [Cron](https://github.com/robfig/cron) - Agendamento de tarefas
- [Go Modules](https://blog.golang.org/using-go-modules) - Gerenciamento de dependências

<br>

## 🚀 Instalação

### Pré-requisitos
- Go 1.20 ou superior
- MongoDB 4.4 ou superior
- Chrome/Chromium (para web scraping)

### Usando Go

```bash
# Clonar o repositório
git clone https://github.com/bulletdev/lta-results-api.git
cd lta-results-api

# Configurar variáveis de ambiente
cp .env .env
# Edite o arquivo .env com suas configurações

# Instalar dependências
go mod download

# Compilar
go build -o lta-api ./cmd/api

# Executar
./lta-api
```

### Usando Docker

```bash
# Clonar o repositório
git clone https://github.com/bulletdev/lta-results-api.git
cd lta-results-api

# Configurar variáveis de ambiente
cp .env .env
# Edite o arquivo .env com suas configurações

# Construir e iniciar containers
docker-compose up -d

# A API estará disponível em http://localhost:8080
```

## 🔐 Configuração

### Variáveis de Ambiente

O projeto requer as seguintes variáveis de ambiente:

```env
# Configurações do Servidor
PORT=8080

# Configurações do MongoDB
MONGODB_USERNAME=
MONGODB_PASSWORD=
MONGODB_CLUSTER=
MONGODB_DATABASE=

# Configurações de Segurança
ADMIN_API_KEY=
```

### Configuração do MongoDB

1. Crie um cluster no MongoDB Atlas
2. Configure o acesso à rede (IP whitelist)
3. Crie um usuário com permissões de leitura/escrita
4. Obtenha a string de conexão:
   ```
   mongodb+srv://<username>:<password>@<cluster>/<database>?retryWrites=true&w=majority
   ```

<br>

## 📍 Endpoints da API

### Resultados de Partidas

#### `GET /api/v1/results`
Listar resultados de partidas com filtros opcionais.

**Parâmetros de consulta:**
- `region` - Filtrar por região (sul, norte)
- `team` - Filtrar por time
- `limit` - Número de resultados por página (padrão: 10)
- `page` - Número da página (padrão: 1)

**Exemplo de resposta:**
```json
{
  "results": [
    {
      "id": "6457e2eb7ac0b2a4f86c2d3a",
      "matchId": "sul-123",
      "date": "2025-04-10T13:00:00Z",
      "teamA": "PAIN",
      "teamB": "RED",
      "scoreA": 1,
      "scoreB": 0,
      "region": "sul",
      "winner": "PAIN",
      "duration": "32:15"
    }
  ],
  "pagination": {
    "total": 24,
    "page": 1,
    "limit": 10,
    "pages": 3
  }
}
```

#### `GET /api/v1/results/:matchId`
Obter detalhes de uma partida específica.

**Exemplo de resposta:**
```json
{
  "id": "6457e2eb7ac0b2a4f86c2d3a",
  "matchId": "sul-123",
  "date": "2025-04-10T13:00:00Z",
  "teamA": "PAIN",
  "teamB": "RED",
  "scoreA": 1,
  "scoreB": 0,
  "region": "sul",
  "players": [
    {
      "name": "Wizer",
      "team": "PAIN",
      "position": "TOP",
      "champion": "Aatrox",
      "kills": 5,
      "deaths": 1,
      "assists": 8,
      "cs": 215,
      "damageDealt": 18500,
      "visionScore": 32
    }
  ],
  "winner": "PAIN",
  "duration": "32:15",
  "tournamentStage": "Regular Season"
}
```

### Estatísticas de Jogadores

#### `GET /api/v1/players/:playerName/stats`
Obter estatísticas agregadas de um jogador.

**Exemplo de resposta:**
```json
{
  "playerName": "Wizer",
  "totalGames": 12,
  "wins": 8,
  "losses": 4,
  "winRate": 66.67,
  "averageKills": 3.5,
  "averageDeaths": 2.1,
  "averageAssists": 6.4,
  "averageCS": 201.3,
  "kda": "4.71"
}
```

### Estatísticas de Times

#### `GET /api/v1/teams/:teamName/stats`
Obter estatísticas agregadas de um time.

**Exemplo de resposta:**
```json
{
  "teamName": "PAIN",
  "totalGames": 16,
  "wins": 10,
  "losses": 6,
  "winRate": 62.50,
  "averageGameDuration": 31.24,
  "mostPlayedChampions": [
    {
      "champion": "Aatrox",
      "games": 8,
      "wins": 6,
      "winRate": 75.0
    }
  ]
}
```

### Endpoints Administrativos

> ⚠️ **Nota:** Todos os endpoints administrativos requerem autenticação via header `X-API-Key`.

#### `POST /api/v1/admin/scrape`
Iniciar processo de scraping manualmente.

#### `POST /api/v1/admin/results`
Adicionar um resultado manualmente.

#### `PUT /api/v1/admin/results/:matchId`
Atualizar um resultado existente.

#### `DELETE /api/v1/admin/results/:matchId`
Excluir um resultado.

<br>

## 🐛 Troubleshooting

### Problemas comuns e soluções:

1. **Erro de conexão com o MongoDB**
   - Verifique se as credenciais estão corretas
   - Confirme se o IP está na whitelist
   - Verifique se o cluster está online

2. **Erro no scraping**
   - Verifique se o Chrome/Chromium está instalado
   - Confirme se as URLs estão acessíveis
   - Verifique os logs para mais detalhes

3. **Erro de autenticação**
   - Verifique se a chave de API está correta
   - Confirme se o header `X-API-Key` está sendo enviado
   - Verifique se a chave está configurada no `.env`

### Logs
Os logs são exibidos no console e incluem:
- Conexão com o banco de dados
- Operações de scraping
- Erros e exceções
- Requisições à API

<br>

## 🔒 Segurança

### Autenticação da API
- Todas as rotas administrativas requerem uma chave de API
- Configure a variável `ADMIN_API_KEY` no arquivo `.env`
- Inclua a chave no header `X-API-Key` das requisições

### Boas Práticas
- Nunca compartilhe sua chave de API
- Use HTTPS em produção
- Mantenha as dependências atualizadas
- Monitore os logs regularmente
- Implemente rate limiting em produção

<br>

## 💡 Como Contribuir

Contribuições são bem-vindas! Aqui estão algumas maneiras de ajudar:

1. **Reportar bugs**: Abra issues descrevendo problemas encontrados
2. **Sugerir melhorias**: Compartilhe ideias para novos recursos ou aprimoramentos
3. **Enviar pull requests**: Implemente correções ou novos recursos
4. **Melhorar a documentação**: Ajude a tornar a documentação mais clara e completa

Para contribuir com código:

```bash
# 1. Faça um fork do repositório
# 2. Clone seu fork
git clone https://github.com/bulletdev/lta-results-api.git

# 3. Crie uma branch para sua feature
git checkout -b feature/nova-funcionalidade

# 4. Faça suas alterações e commit
git commit -m "Adiciona nova funcionalidade"

# 5. Envie para o seu fork
git push origin feature/nova-funcionalidade

# 6. Abra um Pull Request
```

<br>

## 📜 Licença

Este projeto está licenciado sob a [Licença MIT](LICENSE).

<br>

---

<div align="center">
  <p>
    Desenvolvido com ❤️ para a comunidade de League of Legends
  </p>
  <p>
    <a href="https://github.com/bulletdev">GitHub</a> •
    <a href="https://twitter.com/bulletonrails">Twitter</a> •
    <a href="https://linkedin.com/in/michael-bullet">LinkedIn</a>
  </p>
</div>
