# Task 10 - Tratamento de deletados e reconciliacao

## Objetivo
Atualizar banco para arquivos removidos via step dedicado, sem pipeline completa.

## Contexto atual
- Logica de deletados esta acoplada em fluxo legado (`findFilesDeleted`).

## Escopo
- Criar step `mark_deleted` para refletir arquivos ausentes no filesystem.
- Integrar `mark_deleted` ao `startup_scan` e jobs de evento FS.
- Garantir consistencia de `deleted_at` e reversao quando arquivo reaparece.
- Opcional: criar job periodico de reconciliacao para faltas de thumbnail/checksum.

## Arquivos alvo
- `backend/internal/worker/files.go` (migracao/remocao logica antiga)
- novos executores de `mark_deleted/reconcile`

## Criterios de pronto
- Deletados atualizam estado no banco sem full reprocess.
- Fluxo tolera reinicio do backend sem perder consistencia.

## Dependencias
- Tasks 6 e 9.
