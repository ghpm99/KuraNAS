# Backend KuraNAS

Serviço backend do KuraNAS, implementado em Go, responsável por API HTTP, persistência, processamento assíncrono e internacionalização.

## Stack

- Go
- Gin
- `database/sql`
- PostgreSQL
- Workers para processamento em background

## Estrutura

```text
backend/
├── cmd/nas/                         # entrypoints (`main.go`, `main_windows.go`)
├── internal/
│   ├── api/v1/                      # handlers HTTP por domínio
│   ├── app/                         # bootstrap e rotas
│   ├── config/                      # carga de configuração e ambiente
│   └── worker/                      # orquestração de workers
├── pkg/
│   ├── database/
│   │   ├── migrations/              # migrations e registro
│   │   └── queries/                 # SQL por domínio
│   ├── i18n/                        # loader e resolução de traduções
│   ├── logger/                      # logging
│   └── utils/                       # utilitários compartilhados
├── tests/                           # suites de teste adicionais
├── translations/                    # arquivos JSON de tradução
└── Makefile
```

## Execução e Build

Modo desenvolvimento (tag `dev`, porta `8000`):

```bash
make -C backend run
```

Build backend:

```bash
make -C backend build
```

## Testes

Testes com tag `dev` (suite do backend):

```bash
make -C backend test
```

Cobertura via Makefile:

```bash
make -C backend coverage
```

Cobertura geral recomendada:

```bash
cd backend && go test ./... -cover
```

## Configuração de Ambiente

O backend tenta carregar variáveis de um arquivo `.env` e, se não encontrar, usa o ambiente do sistema.

Caminho esperado do `.env` por build:

- `dev`: `backend/.env`
- Linux (release): `/etc/kuranas/.env`
- Windows (release): `%ProgramFiles%/Kuranas/.env`

Atualmente o projeto não possui `backend/.env.example`. Use a tabela abaixo como referência oficial.

### Variáveis da aplicação

| Variável | Obrigatória | Padrão | Observações |
| --- | --- | --- | --- |
| `ENTRY_POINT` | Sim | - | Diretório raiz monitorado pelo NAS. |
| `LANGUAGE` | Sim | - | Idioma base (ex.: `pt-BR`, `en-US`). |
| `ENABLE_WORKERS` | Não | `false` | Ativa workers em background quando `true`. |
| `ENV` | Não | vazio | Nome do ambiente (`dev`, `test`, `prod` etc.). |
| `DB_HOST` | Sim | - | Host do PostgreSQL. |
| `DB_PORT` | Sim | - | Porta do PostgreSQL (ex.: `5432`). |
| `DB_USER` | Sim | - | Usuário do banco. |
| `DB_PASSWORD` | Sim | - | Senha do banco. |
| `DB_NAME` | Sim | - | Nome do banco. |
| `ALLOWED_ORIGINS` | Sim | - | Lista de origens CORS separadas por vírgula. |

### Variáveis de workers (opcionais)

| Variável | Padrão | Observações |
| --- | --- | --- |
| `WORKER_CONCURRENCY_CHECKSUM` | `3` | Concorrência para jobs de checksum. |
| `WORKER_CONCURRENCY_METADATA` | `3` | Concorrência para extração de metadados. |
| `WORKER_CONCURRENCY_THUMBNAIL` | `2` | Concorrência para thumbnails. |
| `WORKER_RETRY_BACKOFF_MS` | `500` | Backoff de retry em milissegundos. |
| `WORKER_SCHEDULER_POLL_MS` | `2000` | Intervalo do scheduler em milissegundos. |
| `WORKER_MAX_CONCURRENT_JOBS` | `4` | Limite total de jobs concorrentes. |

### Variáveis de IA (opcionais)

Se nenhuma chave de IA for definida, o serviço de IA é desativado automaticamente.

| Variável | Padrão | Observações |
| --- | --- | --- |
| `AI_OPENAI_API_KEY` | vazio | Chave da OpenAI. |
| `AI_OPENAI_MODEL` | `gpt-4o-mini` | Modelo padrão OpenAI. |
| `AI_OPENAI_BASE_URL` | `https://api.openai.com/v1` | URL base da OpenAI. |
| `AI_ANTHROPIC_API_KEY` | vazio | Chave da Anthropic. |
| `AI_ANTHROPIC_MODEL` | `claude-sonnet-4-20250514` | Modelo padrão Anthropic. |
| `AI_TIMEOUT_SECONDS` | `30` | Timeout das chamadas de IA. |
| `AI_MAX_RETRIES` | `2` | Número de tentativas por chamada. |
| `AI_RETRY_BACKOFF_MS` | `500` | Backoff entre retries. |

### Exemplo de `backend/.env`

```dotenv
# App
ENTRY_POINT=/mnt/storage
LANGUAGE=pt-BR
ENABLE_WORKERS=true
ENV=dev
ALLOWED_ORIGINS=http://localhost:5173,http://127.0.0.1:5173

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=kuranas
DB_PASSWORD=secret
DB_NAME=kuranas

# Worker tuning (opcional)
WORKER_CONCURRENCY_CHECKSUM=3
WORKER_CONCURRENCY_METADATA=3
WORKER_CONCURRENCY_THUMBNAIL=2
WORKER_RETRY_BACKOFF_MS=500
WORKER_SCHEDULER_POLL_MS=2000
WORKER_MAX_CONCURRENT_JOBS=4

# IA (opcional)
# AI_OPENAI_API_KEY=
# AI_OPENAI_MODEL=gpt-4o-mini
# AI_OPENAI_BASE_URL=https://api.openai.com/v1
# AI_ANTHROPIC_API_KEY=
# AI_ANTHROPIC_MODEL=claude-sonnet-4-20250514
# AI_TIMEOUT_SECONDS=30
# AI_MAX_RETRIES=2
# AI_RETRY_BACKOFF_MS=500
```

## API

- Prefixo principal: `/api/v1`
- Domínios existentes: `files`, `music`, `video`, `analytics`, `diary`, `configuration`, `jobs`, `search`, `notifications`, `update`

## Banco, SQL e Migrations

- Queries SQL: `pkg/database/queries/<feature>`
- Migrations: `pkg/database/migrations/queries`
- Registro de migrations: `pkg/database/migrations/migrations.go`

Não alterar migrations antigas de forma incompatível; criar nova migration para mudanças de schema.

## Internacionalização (Obrigatória)

- Não hardcode mensagens visíveis ao usuário.
- Toda nova chave de texto deve ser adicionada em `backend/translations`.
- O frontend consome as mesmas chaves via endpoint de configuração de tradução.

## Padrões de Implementação

Antes de alterar código backend, seguir:

- `/docs/standards/backend-standards.md`

Pontos obrigatórios:

- fluxo em camadas `Handler -> Service -> Repository`;
- regras de negócio em service, SQL em repository;
- sem bypass de camadas.
