# KuraNAS

Sistema NAS pessoal com backend em Go e frontend em React, focado em gerenciamento de arquivos, mídia e monitoramento de uso.

## Visão Geral

- Backend: API HTTP, regras de negócio, acesso a banco, workers e i18n.
- Frontend: SPA React + TypeScript consumindo API `/api/v1`.
- Build integrado: geração de artefatos em `build/`.

## Estrutura do Repositório

```text
.
├── backend/            # API, workers, banco, i18n e scripts de backend
├── frontend/           # Aplicação web React + Vite + TypeScript
├── build/              # Saída de build integrada (gerado)
├── docs/               # Padrões e documentação de engenharia
├── Makefile            # Build integrado do projeto
└── AGENTS.md           # Regras de colaboração para agentes
```

## Pré-requisitos

- Go (recomendado: 1.24+)
- Node.js (recomendado: 20+)
- Yarn 1.x
- Make

## Desenvolvimento

1. Instale dependências do frontend:

```bash
cd frontend && yarn
```

2. Suba o backend em modo dev (porta `8000`):

```bash
make -C backend run
```

3. Em outro terminal, suba o frontend:

```bash
cd frontend && yarn dev
```

## Build Integrado

Gera frontend + backend e organiza saída em `build/`:

```bash
make
```

Limpar artefatos:

```bash
make clean
```

## Testes e Qualidade

- Backend cobertura geral:
```bash
cd backend && go test ./... -cover
```
- Backend testes com tag `dev`:
```bash
make -C backend test
```
- Frontend lint:
```bash
cd frontend && yarn lint
```
- Frontend testes:
```bash
cd frontend && yarn test --watchAll=false
```
- Frontend cobertura:
```bash
cd frontend && yarn coverage
```

## Internacionalização (Obrigatória)

Este projeto usa JSON como fonte única de tradução.

- Não hardcode texto visível para usuário em backend ou frontend.
- Novas chaves devem ser adicionadas em `backend/translations`.
- O frontend consome traduções via endpoint de configuração do backend.

## Documentação por Módulo

- Frontend: `frontend/README.md`
- Backend: `backend/README.md`
- Padrões: `docs/standards/frontend-standards.md` e `docs/standards/backend-standards.md`
