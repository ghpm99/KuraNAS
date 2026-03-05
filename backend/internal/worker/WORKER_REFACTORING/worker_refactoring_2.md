# Task 2 - Persistencia de Jobs/Steps (migracao)

## Objetivo
Introduzir persistencia de estado de execucao para jobs e steps no banco.

## Contexto atual
- Nao existem tabelas de jobs/steps.
- Progresso e estado dependem do fluxo em memoria.

## Escopo
- Criar novas migracoes SQL para tabelas de job e step.
- Incluir colunas minimas:
  - jobs: `id`, `type`, `priority`, `scope_json`, `status`, `created_at`, `started_at`, `ended_at`, `cancel_requested`, `last_error`
  - steps: `id`, `job_id`, `type`, `status`, `depends_on_json`, `attempts`, `max_attempts`, `last_error`, `progress`, `payload_json`, `created_at`, `started_at`, `ended_at`
- Criar indices para consulta de jobs ativos/recentes e steps por job.
- Registrar migracoes em `migrations.go`.

## Arquivos alvo
- `backend/pkg/database/migrations/queries/*`
- `backend/pkg/database/migrations/migrations.go`

## Criterios de pronto
- Migracoes aplicam sem erro em ambiente limpo.
- Reexecucao de migracao nao causa inconsistencias.
- Estrutura suporta restart sem perder estado.

## Dependencias
- Task 1.
