# Fase 0 — Convenção de nomes + regra no CLAUDE.md

**Risco:** mínimo · **Pré-requisito:** nenhum · **Status:** ✅ concluída (2026-06-09)

> Renomeados 10 arquivos para `snake_case` (8 em `internal/worker/` + `pkg/database/dbContext.go`/`dbContext_test.go`). `make ci-backend` verde, cobertura 81%.

## Objetivo

Estabelecer a regra de organização como padrão do projeto e uniformizar o nome dos arquivos para `snake_case`, sem tocar em imports nem em comportamento.

## Por quê

- Hoje o `worker/` mistura `camelCase` (`directoryWalker.go`, `fileChecksum.go`) com `snake_case` (`job_domain.go`), o que quebra a ordenação visual e a previsibilidade.
- A regra de estrutura (núcleo + extensões, "pacote não é dono de tabela", direção de dependência) precisa estar escrita **antes** de migrar código, para que as fases seguintes apenas a realizem.
- Renomear arquivo dentro do **mesmo pacote** não muda import: é a mudança mais barata e segura possível.

## O que precisa ser feito

### 1. Regra no CLAUDE.md (✅ já aplicado)

- Raiz `CLAUDE.md` → seção "Backend domains: a generic file core + type extensions".
- `backend/CLAUDE.md` → seção "Domain package organization — generic file core + type extensions" + nota de `snake_case` no "add a feature" + layout-alvo na seção "Worker subsystem".

### 2. Renomear arquivos do worker para `snake_case`

Usar `git mv` (mesmo pacote `worker` → zero mudança de import):

| De | Para |
|---|---|
| `directoryWalker.go` | `directory_walker.go` |
| `fileChecksum.go` | `file_checksum.go` |
| `fileMetadata.go` | `file_metadata.go` |
| `fileDatabasePersistence.go` | `file_database_persistence.go` |
| `fileDtoConverter.go` | `file_dto_converter.go` |
| `fileProcessingPipeline.go` | `file_processing_pipeline.go` |
| `fileResultMonitor.go` | `file_result_monitor.go` |
| `videoPlaylist.go` | `video_playlist.go` |

> Varrer o resto do backend por outros `camelCase` em nomes de arquivo `.go` e normalizar junto, se houver.

## Resultado esperado

- Todos os arquivos `.go` em `snake_case`.
- Regra de estrutura documentada e canônica nos `CLAUDE.md`.
- Nenhuma mudança de código, import ou rota.

## Critério de aceite

- `git mv` apenas (diff mostra renomes, não edição de conteúdo).
- `make ci-backend` verde.
