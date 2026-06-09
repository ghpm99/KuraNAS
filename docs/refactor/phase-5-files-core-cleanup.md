# Fase 5 — Arrumar o núcleo `files` por arquivo

**Risco:** baixo · **Pré-requisito:** Fases 0–4 · **Natureza:** renomeio/agrupamento dentro do pacote `files` (sem novos pacotes)

## Objetivo

Com imagem/música/vídeo já fora, o que resta em `files` é genuinamente "arquivo". Padronizar os arquivos por responsabilidade (estilo `net/http`: um pacote coeso, muitos arquivos com nome claro), sem criar sub-pacotes.

## Por quê

- Depois das Fases 2–4, `files` deixa de ser god-package, mas ainda tem nomes herdados inconsistentes (`repository_listing.go`, `handler_reports.go`, `operations_service.go`…). Uniformizar fecha a sensação de "abri e me encontrei".
- Estes conceitos (listing, recent, reports, operations, blob) são **coesos com arquivo** — não passam no teste de "domínio próprio", então **ficam em `files` fatiados por arquivo**, não viram pacote.

## O que precisa ser feito

### Reagrupamento por arquivo (dentro de `files/`)

| Hoje | Vira | Conteúdo |
|---|---|---|
| `repository_listing.go` + `handler_listing.go` | `file_listing.go` (ou `listing.go`) | listagem/tree/filtro |
| `recent_service.go` + `recent_repository.go` + `recent_internal_test.go` | `recent.go` (+ teste) | acesso recente. *Opcional:* extrair sub-pacote `files/recent/` se crescer |
| `handler_reports.go` + `repository_reports.go` | `reports.go` | total space/files/dir, size-by-format, top-by-size, duplicates |
| `operations_handler.go` + `operations_service.go` | `operations.go` | upload/move/copy/rename/delete/folder |
| `handler_media_stream.go` (só `GetBlobFileHandler` + `GetFileThumbnailHandler` genéricos restantes) | `blob.go` | blob/thumbnail genérico |

### Estado final esperado de `model.go` / `dto.go`

Após as extrações, devem conter **só** o que é do arquivo genérico:

- `model.go`: `FileModel`, `RecentFileModel`, `SizeReportModel`, `DuplicateFilesModel`.
- `dto.go`: `FileDto`, `FileStat`, `FileFilter`, `FileBlob`, `RecentFileDto`, `SizeReportDto`, `DuplicateFileDto`, `DuplicateFileReportDto`.

(Os DTOs/Models de música, imagem e vídeo já saíram nas fases anteriores.)

### Passos

1. `git mv`/merge dos pares handler+repository por tema, ajustando para nomes consistentes.
2. Garantir prefixo de camada coerente nos arquivos restantes (`handler*`, `service*`, `repository*`).
3. Conferir que `files` não importa nenhuma extensão (`image`/`music`/`video`) — deve estar limpo.

## Resultado esperado

- `files/` legível: cada arquivo tem um tema óbvio, tudo `snake_case`, testes ao lado.
- Núcleo genérico de verdade — nada específico de tipo de mídia.

## Critério de aceite

- `make ci-backend` verde.
- `grep` confirma: nenhum `import` de `image`/`music`/`video` dentro de `files`.
- `make ci` (frontend + backend) verde antes de encerrar a refatoração.
