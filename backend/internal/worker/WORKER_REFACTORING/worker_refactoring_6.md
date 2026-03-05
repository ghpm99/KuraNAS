# Task 6 - Extrair executores de step atomicos

## Objetivo
Transformar logicas atuais em executores atomicos reutilizaveis por job system.

## Contexto atual
- Pipeline atual encadeia scanner, metadata, checksum e persistencia em um fluxo unico.

## Escopo
- Extrair/adapter para executores de step:
  - `metadata`
  - `checksum`
  - `persist`
  - `thumbnail`
  - `playlist_index`
- Garantir contratos claros de input/output por step.
- Garantir idempotencia basica e `skipped` quando up-to-date.
- Remover logica de "pipeline inteira" de cada executor.

## Arquivos alvo
- `backend/internal/worker/fileMetadata.go`
- `backend/internal/worker/fileChecksum.go`
- `backend/internal/worker/fileDatabasePersistence.go`
- `backend/internal/worker/thumbnail.go`
- `backend/internal/worker/videoPlaylist.go`

## Criterios de pronto
- Cada executor possui uma responsabilidade.
- Steps podem ser chamados isoladamente pelo scheduler.
- Sem regressao de comportamento para thumbnail/playlist.

## Dependencias
- Task 4.
