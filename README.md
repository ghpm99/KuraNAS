# KuraNAS

Sistema NAS pessoal com backend em Go e frontend em React para gerenciamento de arquivos, mídia e recursos de organização.

## Visão Geral

- Backend (`backend/`): API HTTP, regras de negócio, workers e i18n.
- Frontend (`frontend/`): SPA React + TypeScript consumindo `/api/v1`.
- Build integrado: empacotamento final em `build/`.

## Estrutura

```text
.
├── backend/            # API, workers, banco, i18n e scripts
├── frontend/           # Aplicação web (Vite + React + TypeScript)
├── docs/               # Padrões de engenharia
├── build/              # Saída do build integrado (gerado)
├── Makefile            # Pipeline local de build/qualidade
└── AGENTS.md           # Regras de colaboração para agentes
```

## Pré-requisitos

- Go 1.24+
- Node.js 20+
- Yarn 1.x
- Make

## Setup Rápido (Desenvolvimento)

1. Instale dependências do frontend:

```bash
cd frontend && yarn
```

2. Configure variáveis do backend em `backend/.env` (detalhes completos em `backend/README.md`).

3. Inicie o backend (modo `dev`, porta `8000`):

```bash
make -C backend run
```

4. Em outro terminal, inicie o frontend:

```bash
cd frontend && yarn dev
```

## Build Integrado

Gera frontend + backend e organiza artefatos em `build/`:

```bash
make
```

Limpeza:

```bash
make clean
```

## Testes e Qualidade

Backend:

```bash
cd backend && go test ./... -cover
make -C backend test
```

Frontend:

```bash
cd frontend && yarn lint
cd frontend && yarn test --watchAll=false
cd frontend && yarn coverage
```

Pipeline local completa:

```bash
make ci
```

## Internacionalização

- Não hardcode texto visível para usuário.
- Backend e frontend devem usar as mesmas chaves em `backend/translations`.
- O frontend obtém traduções via endpoint de configuração do backend.

## Documentação por Módulo

- [README do backend](/home/server/Documentos/Projetos/KuraNAS/backend/README.md)
- [README do frontend](/home/server/Documentos/Projetos/KuraNAS/frontend/README.md)
- [Padrão backend](/home/server/Documentos/Projetos/KuraNAS/docs/standards/backend-standards.md)
- [Padrão frontend](/home/server/Documentos/Projetos/KuraNAS/docs/standards/frontend-standards.md)
