# Task 4 - Servico de orquestracao (Planner) e scheduler base

## Objetivo
Introduzir orquestrador que monta jobs/steps com dependencias, mantendo coexistencia com workers atuais durante migracao.

## Contexto atual
- `StartFileProcessingPipeline` mistura planejamento e execucao.
- Nao existe componente central de planejamento de DAG de steps.

## Escopo
- Criar `JobOrchestrator/Planner` para construir planos por tipo de job.
- Criar scheduler simples que seleciona steps elegiveis por dependencia/status.
- Integrar com startup de app (`backend/internal/app`) sem quebrar queue atual.
- Garantir que worker executor execute apenas step atomico.

## Arquivos alvo
- `backend/internal/worker/*` (novo orchestrator/scheduler)
- `backend/internal/app/app.go` e `context.go` (wiring)

## Criterios de pronto
- E possivel criar job com steps persistidos e iniciar execucao.
- Nenhum step roda antes das dependencias.
- Fluxo legado ainda funcional ate fim da migracao.

## Dependencias
- Tasks 1-3.
