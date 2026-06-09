# Fase 4 — Mover vídeo de `files` para a extensão `video/`

**Risco:** médio · **Pré-requisito:** Fases 0–3 · **Natureza:** reconciliação de duplicação (o pacote `video/` já existe)

## Objetivo

Remover de `files` a navegação/stream de vídeo e a metadata de vídeo, levando para o pacote `video/` que já existe (playback, smart playlists, catálogo). Ao final, o `metadata_repository.go` genérico de `files` deixa de existir.

## Por quê

- Mesmo padrão da Fase 3: `files` tem `repository_video_queries.go` (GetVideos) e handlers de vídeo (`GetVideosHandler`, `StreamVideoHandler`, thumbnails/preview de vídeo) **em paralelo** ao pacote `video/` completo.
- `VideoMetadataModel` está no `metadata_repository.go` genérico de `files`.

## O que precisa ser feito

### Mapeamento (sai de `files/` → vai para `video/`)

| Fonte em `files/` | Destino em `video/` |
|---|---|
| `repository_video_queries.go` (GetVideos) | `video/repository_browse.go` |
| `VideoMetadataModel` (de `model.go`) + métodos `GetVideoMetadataByID`, `UpsertVideoMetadata`, `DeleteVideoMetadata` (de `metadata_repository.go`) | `video/model.go` + `video/repository.go` |
| Handlers `GetVideosHandler`, `StreamVideoHandler`, `GetVideoThumbnailHandler`, `GetVideoPreviewHandler` (de `handler_media_stream.go`) | `video/handler.go` |

### Rotas afetadas (preservar URLs)

Passam de `files` para `video` com **o mesmo path**:

- `GET /files/videos`
- `GET /files/video-stream/:id`
- `GET /files/video-thumbnail/:id`
- `GET /files/video-preview/:id`

> Consolidar em `RegisterVideoRoutes` (já existe), removendo de `RegisterFilesRoutes`. Sem mudança de contrato.

### Passos

1. Mover query de vídeo e metadata de vídeo para `video/`.
2. Mover handlers de vídeo para `video/handler.go`.
3. Realocar rotas para `RegisterVideoRoutes` preservando paths.
4. **Eliminar `metadata_repository.go`** de `files`: após Fases 2–4, ele estará vazio (imagem → image, áudio → music, vídeo → video). Confirmar e remover.
5. Pipeline de scan que grava metadata de vídeo passa a importar `video` (`worker → video → files`).

## Resultado esperado

- "Vídeo" num lugar só (`video/`).
- `metadata_repository.go` removido de `files` — a metadata por tipo agora vive em cada extensão.
- `files` reduzido ao núcleo genérico.

## Critério de aceite

- `make ci-backend` verde.
- Rotas de vídeo idênticas (incluindo `Range`/streaming).
- Sem ciclo `files ↔ video`.
- Validar streaming de vídeo de ponta a ponta (player no frontend).
