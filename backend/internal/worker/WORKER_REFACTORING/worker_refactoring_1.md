# Task 1 - Base de dominio do Job System

## Objetivo
Criar a base de dominio para Jobs/Steps sem alterar o comportamento funcional atual dos workers.

## Contexto atual
- Hoje o sistema usa `utils.Task` com `TaskType` fixo e executa pipeline monolitica em `StartFileProcessingPipeline`.
- Nao existe entidade persistida de job/step, prioridade ou estado detalhado.

## Escopo
- Definir tipos de dominio para Job/Step em backend (status, tipo, prioridade, timestamps, erro, tentativas).
- Definir enums de status:
  - Job: `queued`, `running`, `partial_fail`, `failed`, `completed`, `canceled`
  - Step: `queued`, `running`, `completed`, `failed`, `canceled`, `skipped`
- Definir tipos iniciais de job: `startup_scan`, `upload_process`, `fs_event`, `reindex_folder`.
- Definir tipos iniciais de step: `scan_filesystem`, `diff_against_db`, `metadata`, `checksum`, `persist`, `thumbnail`, `playlist_index`, `mark_deleted`.
- Criar contrato de payload de escopo (file/path/root) para job/step.

## Arquivos alvo
- `backend/internal/worker/*` (novo modulo de dominio)
- `backend/pkg/utils/dto.go` (apenas se necessario para coexistencia)

## Criterios de pronto
- Dominio tipado criado e compilando sem quebrar fluxo atual.
- Sem texto hardcoded novo para usuario final.
- Sem regressao de rotas existentes.

## Dependencias
- Nenhuma.
