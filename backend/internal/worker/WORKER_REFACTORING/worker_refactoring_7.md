# Task 7 - Centralizacao de checksum (fonte unica)

## Objetivo
Garantir que checksum seja calculado apenas pelo step oficial de checksum.

## Contexto atual
- Checksum aparece em multiplos fluxos (`service.UpdateCheckSum`, worker dedicado e pipeline).
- Risco de inconsistencias entre caminhos diferentes.

## Escopo
- Definir step `checksum` como unica fonte de verdade.
- Remover/aposentar chamadas de checksum fora do job system.
- Ajustar upload/startup/watcher para nunca calcular checksum diretamente.
- Implementar regra de skip para checksum atualizado.

## Arquivos alvo
- `backend/internal/api/v1/files/service.go`
- `backend/internal/worker/checksum.go`
- `backend/internal/worker/files.go`
- `backend/internal/worker/fileProcessingPipeline.go` (durante transicao)

## Criterios de pronto
- Nao existem caminhos paralelos de checksum fora do step oficial.
- Reexecucao do mesmo step nao duplica nem corrompe dados.
- Estados `completed/skipped/failed` registrados por step.

## Dependencias
- Tasks 4 e 6.
