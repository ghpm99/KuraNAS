# Task 9 - Startup scan incremental como job

## Objetivo
Substituir startup scan monolitico por job incremental (`scan + diff + fan-out`).

## Contexto atual
- Startup enfileira `ScanFiles` e executa pipeline pesada.
- Arquivos unchanged podem ser reprocessados desnecessariamente.

## Escopo
- Criar job `startup_scan` no boot quando workers habilitados.
- Implementar steps:
  - `scan_filesystem`
  - `diff_against_db` (size + mtime como base)
  - fan-out para arquivos `new/modified`
- Definir prioridade `low` para startup scan.
- Garantir que `unchanged` nao avancem para steps pesados.

## Arquivos alvo
- `backend/internal/worker/worker.go`
- `backend/internal/app/app.go`
- novo(s) executor(es) de scan/diff

## Criterios de pronto
- Boot cria job `startup_scan` rastreavel por `job_id`.
- Apenas `new/modified` seguem para metadata/checksum/thumbnail.
- Sem reprocessamento total desnecessario.

## Dependencias
- Tasks 4-7.
