# Task 5 - API de observabilidade de jobs

## Objetivo
Expor endpoints para UI acompanhar execucao por polling com base no estado persistido.

## Contexto atual
- Nao ha endpoints de job/step.
- Criterio exige consulta de job por id, listagem de jobs e steps por job.

## Escopo
- Criar modulo API `jobs` (handler/service/repository).
- Endpoints minimos:
  - `GET /api/v1/jobs/:id`
  - `GET /api/v1/jobs`
  - `GET /api/v1/jobs/:id/steps`
  - (opcional) `POST /api/v1/jobs/:id/cancel`
- Calcular progresso geral por agregacao de steps persistidos.
- Registrar novas rotas em `backend/internal/app/routes.go`.
- Adicionar chaves i18n para mensagens de erro/sucesso expostas ao usuario.

## Arquivos alvo
- `backend/internal/api/v1/jobs/*`
- `backend/internal/app/routes.go`
- `backend/translations/*.json`

## Criterios de pronto
- UI/API consumer consegue acompanhar progresso sem canal acoplado.
- Respostas usam codigos HTTP consistentes.
- Strings de usuario saem de i18n.

## Dependencias
- Tasks 3-4.
