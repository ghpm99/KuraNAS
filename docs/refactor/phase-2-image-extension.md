# Fase 2 — Criar a extensão `image/`

**Status:** ✅ concluída (2026-06-09)

**Risco:** baixo · **Pré-requisito:** Fases 0–1 · **Natureza:** extração nova (pacote não existe ainda)

## Objetivo

Tirar tudo que é específico de **imagem** de dentro de `files` e colocar num novo pacote `internal/api/v1/image/`, que importa `files` e dá `JOIN` na tabela `files`.

## Por quê

- Hoje "imagem" está dissolvida em `files`: a classificação por IA, o modelo de metadata de imagem e o CRUD da metadata convivem com 30 outros arquivos. Não existe endereço óbvio para "lógica de imagem".
- É a primeira extração e a mais segura: o pacote `image` **não existe**, então criamos do zero sem reconciliar duplicação (diferente de música/vídeo).
- Valida o padrão supertype → extensão antes das fases mais arriscadas.

## O que precisa ser feito

### Mapeamento (sai de `files/` → vai para `image/`)

| Fonte em `files/` | Destino em `image/` |
|---|---|
| `image_classification.go` (+ `image_classification_test.go`) | `image/classification.go` (+ teste) |
| `ImageMetadataModel`, `ImageClassificationModel` (de `model.go`) | `image/model.go` |
| Métodos `GetImageMetadataByID`, `UpsertImageMetadata`, `DeleteImageMetadata` (de `metadata_repository.go`) | `image/repository.go` (com `JOIN` em `files`) |
| Handler/rota `GET /files/images` (`GetImagesHandler`) | `image/handler.go` |
| Queries `.sql` de imagem em `pkg/database/queries/file/` | nova pasta `pkg/database/queries/image/` |

### Passos

1. Criar `internal/api/v1/image/` com `model.go`, `repository.go`, `service.go`, `handler.go`, `interfaces.go` (template canônico).
2. Mover o código da tabela acima; `image` passa a `import .../files` para usar `files.FileModel`/`FileDto`.
3. Criar `pkg/database/queries/image/` e mover os `.sql` de imagem (com `//go:embed`).
4. Adicionar `newImageContext` em `internal/app/context.go` (`repository → service → handler`) e pendurar em `AppContext`.
5. Adicionar `RegisterImageRoutes` em `internal/app/routes.go`.
6. **Manter a URL existente** (`/files/images` continua respondendo igual) — só muda o pacote que serve. Se decidir promover para `/images`, é mudança de contrato → fora desta fase.

### Ponto de atenção: caminho de escrita (worker)

A pipeline de scan (`worker/scan/file_metadata.go`) que grava metadata de imagem passará a importar `image`. Isso é aceitável: o worker é orquestrador e pode depender de domínios (direção única `worker → image → files`). Confirmar que não cria ciclo.

### O que **não** vai para `image`

- `GetFileThumbnailHandler` genérico fica no núcleo `files` (serve thumbnail de qualquer arquivo).

## Resultado esperado

- Pacote `image/` coeso: tudo de imagem num lugar só, importando `files`.
- `files` deixa de conter `ImageMetadataModel`/`ImageClassificationModel` e os métodos de metadata de imagem.
- `files` continua sem conhecer `image` (direção correta).

## Critério de aceite

- `make ci-backend` verde (cobertura do novo pacote ≥80%).
- `GET /files/images` responde idêntico ao antes (mesmo path/shape).
- `go vet ./...` sem ciclo `files ↔ image`.
