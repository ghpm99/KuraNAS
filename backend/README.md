# Backend KuraNAS

Serviço backend do KuraNAS, implementado em Go, responsável por API HTTP, persistência, processamento assíncrono e internacionalização.

## Stack

- Go
- Gin
- `database/sql`
- SQLite/PostgreSQL (configuração por ambiente)
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

## Execução

Modo desenvolvimento (tag `dev`, porta `8000`):

```bash
make -C backend run
```

Build backend:

```bash
make -C backend build
```

## Testes

Testes com tag `dev`:

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

## Configuração

O backend carrega variáveis a partir de `.env` e ambiente do sistema.

Principais variáveis:

- `ENTRY_POINT`: diretório raiz monitorado para arquivos.
- `LANGUAGE`: idioma padrão (`pt-BR`, `en-US`, etc.).
- `ENABLE_WORKERS`: ativa/desativa workers (`true`/`false`).
- `DB_PATH`: caminho do SQLite local.
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: configuração alternativa de banco relacional.
- `ALLOWED_ORIGINS`: CORS (lista separada por vírgula).
- `ENV`: ambiente de execução.

## API

- Prefixo principal: `/api/v1`
- Domínios existentes: `files`, `music`, `video`, `analytics`, `diary`, `configuration`, `update`

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

- `docs/standards/backend-standards.md`

Pontos obrigatórios:

- fluxo em camadas `Handler -> Service -> Repository`;
- regras de negócio em service, SQL em repository;
- sem bypass de camadas.
