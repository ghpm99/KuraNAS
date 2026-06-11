# 07 — Remover o pipeline legado de scan (código morto)

**Tipo:** dívida técnica · **Prioridade:** P2

## Contexto

O `backend/CLAUDE.md` documenta "dois modelos de execução coexistindo": o canal legado de tasks (`chan utils.Task`) e o orquestrador de jobs (preferido). Na prática, os ramos legados de **scan** são código morto: em `internal/worker/engine/worker.go` (`handleTask`, `startWorkersScheduler`) e `watcher.go`, o caminho legado só executa quando `context.JobOrchestrator == nil` — e o orquestrador é sempre construído quando há `JobsRepository` (`StartWorkers`).

Custo real dessa dualidade: **o único código que indexava diretórios morava no legado** (`scan.ScanFilesWorker` faz `filepath.Walk` incluindo pastas; o pipeline novo as pula) — foi assim que o bug da task 01 nasceu e passou despercebido. Código morto que "parece cobrir" um caso esconde que o caminho vivo não o cobre.

Inventário do legado de scan:

- `scan.ScanFilesWorker` + `UpdateFileDto`/`CreateFileDto` (`internal/worker/scan/files.go`) — inclusive dispara **uma goroutine por arquivo** para checksum, sem limite.
- `scan.FindFilesDeleted` (mesmo arquivo) — inócua na prática (consulta `deleted_at = <zero>`, nunca casa; ver task 02).
- `scan.StartFileProcessingPipeline` e o ramo `else` de `utils.ScanFiles` em `handleTask`.
- `scan.ScanDirWorker` e o ramo `else` de `utils.ScanDir`.
- O fallback `context.Tasks <- utils.Task{...}` em `startEntryPointWatcher` e o ramo sem orquestrador de `startWorkersScheduler`.

**Não** é legado (continua em uso pelo canal de tasks): `utils.UpdateCheckSum`, `utils.CreateThumbnail`, `utils.GenerateVideoPlaylists` — avaliar migração para jobs em outra ocasião.

## Objetivo

Um único pipeline de indexação (o orquestrador de jobs), com o orquestrador tratado como dependência obrigatória do subsistema de workers. Menos superfície para bugs se esconderem.

## O que fazer

1. Remover os ramos `else` (sem orquestrador) de `handleTask`, `startWorkersScheduler` e `startEntryPointWatcher`.
2. Remover as funções legadas de scan e seus testes.
3. Tornar o orquestrador obrigatório: se `JobsRepository` for nil em `StartWorkers`, falhar alto (log/erro claro) em vez de degradar silenciosamente para um caminho sem manutenção.

## Como fazer

- **Pré-requisito**: task 01 concluída — só remover `ScanFilesWorker` depois que o pipeline novo indexa diretórios, senão perde-se a única referência do comportamento correto.
- Fazer a remoção em passos compiláveis: primeiro os ramos de fallback (deixando o orquestrador obrigatório), depois as funções de `scan/files.go` que ficarem sem chamadores, depois os testes correspondentes (`scan_test.go`, partes de `workers_internal_test.go` que testam o fallback).
- `FindFilesDeleted`: remover junto — o job `mark_deleted` do orquestrador é o substituto vivo.
- Verificar com `go vet` + busca por referências que nada além de testes apontava para os símbolos removidos.
- Atualizar `backend/CLAUDE.md` (seção "Worker subsystem"): remover a menção aos "dois modelos coexistindo" no que tange a scan, deixando claro o que resta no canal de tasks.
- A cobertura ≥ 80% tende a **subir** com a remoção de código morto, mas conferir — os testes do legado também saem.

## Critérios de aceite

- [x] Nenhuma referência a `ScanFilesWorker`, `StartFileProcessingPipeline`, `ScanDirWorker`, `FindFilesDeleted`, `UpdateFileDto`/`CreateFileDto` no código de produção. *(grep em `internal/`, `pkg/`, `cmd/` → 0 referências, inclusive aos workers de canal `Start*Worker` órfãos; −3362 linhas)*
- [x] `StartWorkers` sem `JobsRepository` falha de forma explícita (não degrada silenciosamente). *(loga ERROR e não inicia nada; `TestStartWorkersWithoutJobsRepositoryRefusesToStart`)*
- [x] Tasks `UpdateCheckSum`/`CreateThumbnail`/`GenerateVideoPlaylists` continuam funcionando pelo canal. *(`TestWorkerKnownTaskBranches`, `TestUpdateCheckSumWorker*`, `TestCreateThumbnailWorkerAndVideoPlaylistWorker`)*
- [x] Comportamento de scan inalterado (startup_scan, fs_event, watcher) — validado pelos testes de integração existentes. *(suite completa verde, incl. `diff_step_pg_integration_test` e `mark_deleted_pg_integration_test` contra Postgres real)*
- [x] `backend/CLAUDE.md` atualizado. *(seção "Worker subsystem": orquestrador único e obrigatório; canal só para checksum/thumbnail/playlists; watcher por eventos nativos)*
- [x] `make ci-backend` verde (cobertura ≥ 80%). *(2026-06-11, 80.0%, integração pg via `TEST_DB_PORT=54329`)*

## Fora de escopo

- Migrar checksum/thumbnail/playlist do canal de tasks para o orquestrador.
- Remover o canal `Tasks` em si (ainda tem consumidores legítimos).
- Qualquer mudança de comportamento do pipeline vivo.
