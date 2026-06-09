# Fase 1 — Quebrar `internal/worker` em sub-pacotes

**Risco:** médio · **Pré-requisito:** Fase 0 · **Importadores externos:** só `internal/app/app.go` e `internal/app/context_run_test.go`

## Objetivo

Separar o pacote único `worker` (38 arquivos, tudo no mesmo nível) em sub-pacotes por responsabilidade: `job`, `engine`, `steps`, `scan`. E consolidar o watcher duplicado.

## Por quê

- Hoje `worker` junta três coisas distintas no mesmo nível: o **motor** (pool/orquestrador/scheduler), os **steps** de job, e a **pipeline de scan** de arquivo. Abrir a pasta não diz o que é o quê.
- Existe `internal/worker/watcher.go` **e** o pacote `internal/watcher/` — duplicação que confunde.
- O blast radius é pequeno (só `app.go` + 1 teste importam `worker`), então é seguro reorganizar agora.

## O que precisa ser feito

### Sub-pacotes-alvo

| Sub-pacote | Recebe (fonte) |
|---|---|
| `worker/job/` | `job_domain.go` (+`job_domain_test.go`) — enums/tipos de Job/Step. **Pacote neutro: `engine` e `steps` importam daqui, ele não importa ninguém → sem ciclo.** |
| `worker/engine/` | `worker.go`, `job_orchestrator.go`, `job_scheduler.go`, `step_executors.go` (+ testes: `workers_internal_test.go`, `more_workers_internal_test.go`, `job_orchestrator_scheduler_test.go`, `job_scheduler_loop_test.go`, `scheduler_observers_test.go`, `step_executors_test.go`, `step_executors_additional_test.go`) |
| `worker/steps/` | `checksum.go` (+`checksum_pipeline_internal_test.go`), `thumbnail.go`, `takeout_step.go`, `takeout_step_payload.go` (+`takeout_step_test.go`), `ollama_step.go` (+`ollama_step_test.go`), `ai_playlist_cluster_step.go` (+`ai_playlist_cluster_step_test.go`), `diff_step_pg_integration_test.go`, `mark_deleted_pg_integration_test.go`, `ai_settings_gate_test.go` |
| `worker/scan/` | pipeline de arquivo: `directory_walker.go`, `dir.go`, `file_checksum.go`, `file_metadata.go`, `file_database_persistence.go`, `file_dto_converter.go`, `file_processing_pipeline.go`, `file_result_monitor.go`, `files.go`, `video_playlist.go` |

### Passos

1. Criar `worker/job/` primeiro (pacote folha, sem dependências internas).
2. Mover `engine`, depois `steps`, depois `scan` — corrigindo imports a cada passo.
3. Ajustar `internal/app/app.go` (que chama `StartWorkers` etc.) e `context_run_test.go` para os novos paths.
4. **Consolidar o watcher:** mover/mesclar `worker/watcher.go` (+`watcher_test.go`) para `internal/watcher/`, eliminando a duplicação. Conferir quem registra o poll de 60s em `app.go`.

### Cuidados

- `worker/scan/` importa `internal/api/v1/files` (usa `files.FileModel` e conversões) — direção válida, mantém.
- Vigiar **ciclos de import** ao separar `engine` ↔ `steps`: o que ambos compartilham (tipos de job) tem que estar em `job/`. Se aparecer ciclo, é sinal de que falta empurrar um tipo para `job/`.

## Resultado esperado

- `internal/worker/` com 4 sub-pacotes coesos, cada um com responsabilidade óbvia.
- Watcher num lugar só (`internal/watcher/`).
- Comportamento idêntico (mesmos workers, mesmos jobs/steps).

## Critério de aceite

- `make ci-backend` verde (testes de worker incluídos).
- `go vet ./...` sem ciclos.
- Subir o servidor e confirmar que workers iniciam (`ENABLE_WORKERS`) e o watcher roda.
