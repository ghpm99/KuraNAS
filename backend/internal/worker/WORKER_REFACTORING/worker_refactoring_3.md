# Task 3 - Repository e queries de Job/Step

## Objetivo
Adicionar camada repository para CRUD e atualizacao de estado de jobs/steps seguindo padrao do projeto.

## Contexto atual
- Nao ha repositorio para jobs.
- Projeto exige `handler -> service -> repository` e queries em `backend/pkg/database/queries`.

## Escopo
- Criar SQL em `backend/pkg/database/queries/jobs`.
- Implementar loader de queries com `go:embed`.
- Implementar repository com operacoes:
  - criar job e steps
  - buscar job por id
  - listar jobs por status/tipo/prioridade
  - buscar steps por job
  - transicao de status (queued->running->completed/failed etc)
  - registrar erro/tentativas/progresso por step
- Garantir `fmt.Errorf(...: %w)` em erros de DB.

## Arquivos alvo
- `backend/pkg/database/queries/jobs/*`
- `backend/internal/api/v1/jobs/*` ou modulo equivalente de backend para repositorio

## Criterios de pronto
- Repository testavel por interface.
- Queries desacopladas do handler.
- Cobertura de cenarios de erro de persistencia.

## Dependencias
- Task 2.
