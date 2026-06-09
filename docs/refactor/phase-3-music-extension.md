# Fase 3 — Mover música de `files` para a extensão `music/`

**Risco:** médio · **Pré-requisito:** Fases 0–2 · **Natureza:** reconciliação de duplicação (o pacote `music/` já existe e já importa `files`)

## Objetivo

Remover de `files` toda a navegação de música e a metadata de áudio, levando para o pacote `music/` que já existe (playlists, catálogo, clustering por IA, player state).

## Por quê

- Há **duplicação real**: `files` contém `repository_music_queries.go` (GetMusic, GetMusicArtists, GetMusicByAlbum…) e DTOs de música, **em paralelo** ao pacote `music/` completo. Dois lugares para "música".
- `music/repository.go` **já importa `files`** (`scanMusicFile(rows, *files.FileModel)`, `getLibraryFiles() []files.FileModel`) — a direção supertype → extensão já existe; só falta mover o resto para lá.
- A metadata de áudio (`AudioMetadataModel`) está no `metadata_repository.go` genérico de `files`, junto com imagem e vídeo.

## O que precisa ser feito

### Mapeamento (sai de `files/` → vai para `music/`)

| Fonte em `files/` | Destino em `music/` |
|---|---|
| `repository_music_queries.go` (GetMusic, GetMusicArtists, GetMusicByArtist, GetMusicAlbums, GetMusicByAlbum, GetMusicGenres, GetMusicByGenre, GetMusicFolders) | `music/repository_browse.go` |
| DTOs `MusicArtistDto`, `MusicAlbumDto`, `MusicGenreDto`, `MusicFolderDto`, `IMusicMetadata` (de `dto.go`) | `music/dto.go` |
| `AudioMetadataModel` (de `model.go`) + métodos `GetAudioMetadataByID`, `UpsertAudioMetadata`, `DeleteAudioMetadata` (de `metadata_repository.go`) | `music/model.go` + `music/repository.go` |
| Handlers de navegação de música + `StreamAudioHandler` | `music/handler.go` |

### Rotas afetadas (preservar URLs)

Hoje servidas pelo handler de `files`, devem passar a ser servidas por `music` **com o mesmo path**:

- `GET /files/music`, `GET /files/videos`→ (vídeo é Fase 4)
- `GET /files/stream/:id` (`StreamAudioHandler`)
- grupo `/files/music/*`: `/artists`, `/artists/:name`, `/albums`, `/albums/:name`, `/genres`, `/genres/:name`, `/folders`

> Consolidar o registro dessas rotas dentro de `RegisterMusicRoutes` (que já existe), removendo-as de `RegisterFilesRoutes`. **Sem mudar path/método/shape** — frontend, Android e plugin dependem disso.

### Passos

1. Mover queries de música (`repository_music_queries.go`) e DTOs para `music/`.
2. Mover metadata de áudio para `music/` (model + métodos do repositório).
3. Mover handlers de música e `StreamAudioHandler` para `music/handler.go`.
4. Migrar `.sql` de música, se aplicável, mantendo `pkg/database/queries/music/`.
5. Realocar as rotas para `RegisterMusicRoutes` preservando os paths.
6. Ajustar `context.go` se a injeção mudar (o `music` context já existe).

### Ponto de atenção

A pipeline de scan que grava metadata de áudio passará a importar `music` (igual ao caso de imagem). Direção única `worker → music → files`.

## Resultado esperado

- "Música" num lugar só (`music/`). `files` perde as queries/DTOs/metadata de música.
- `metadata_repository.go` fica menor (só imagem já saiu na Fase 2; vídeo sai na Fase 4).

## Critério de aceite

- `make ci-backend` verde.
- Todas as rotas de música respondem idênticas (mesmos paths/shapes).
- Sem ciclo `files ↔ music`.
