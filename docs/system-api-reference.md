# Referência Completa de Rotas e Navegação do KuraNAS

## 1. Rotas frontend

| Rota | Papel |
| --- | --- |
| `/` | Redireciona para `/home` |
| `/home` | Dashboard principal |
| `/files/*` | Explorador de arquivos baseado em path |
| `/favorites` | Favoritos |
| `/starred` | Legado, redireciona para favoritos |
| `/settings` | Configurações persistidas |
| `/internal/activity-diary` | Diário de atividade |
| `/activity-diary` | Legado, redireciona para diário |
| `/analytics/*` | Analytics |
| `/about` | Sobre/runtime/build info |
| `/images/*` | Domínio de imagens |
| `/music/*` | Domínio de música |
| `/videos/*` | Domínio de vídeos |
| `/video/:id` | Player de vídeo dedicado |

## 2. Rotas backend

Prefixo comum: `/api/v1`

## 2.1 Files

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/files/` | Lista arquivos por filtro simples |
| `GET` | `/files/tree` | Lista árvore/contexto atual do explorador |
| `GET` | `/files/:id` | Busca item por ID e carrega filhos pelo path |
| `GET` | `/files/recent` | Lista acessos recentes |
| `GET` | `/files/recent/:id` | Histórico de acessos do arquivo |
| `GET` | `/files/path` | Resolve arquivo por path relativo |
| `GET` | `/files/path/:path` | Variante adicional de path |
| `GET` | `/files/thumbnail/:id` | Thumbnail de arquivo |
| `GET` | `/files/video-thumbnail/:id` | Thumbnail de vídeo |
| `GET` | `/files/video-preview/:id` | Preview GIF de vídeo |
| `GET` | `/files/blob/:id` | Conteúdo binário do arquivo |
| `POST` | `/files/update` | Agenda reindexação manual |
| `POST` | `/files/upload` | Upload e criação de job assíncrono |
| `POST` | `/files/folder` | Cria pasta |
| `POST` | `/files/move` | Move arquivo/pasta |
| `POST` | `/files/copy` | Copia arquivo/pasta |
| `POST` | `/files/rename` | Renomeia arquivo/pasta |
| `DELETE` | `/files/path` | Remove arquivo/pasta do disco |
| `POST` | `/files/starred/:id` | Alterna favorito |
| `GET` | `/files/total-space-used` | Total de espaço usado |
| `GET` | `/files/total-files` | Total de arquivos |
| `GET` | `/files/total-directory` | Total de diretórios |
| `GET` | `/files/report-size-by-format` | Relatório por formato |
| `GET` | `/files/top-files-by-size` | Top maiores arquivos |
| `GET` | `/files/duplicate-files` | Duplicados |
| `GET` | `/files/images` | Biblioteca de imagens |
| `GET` | `/files/music` | Biblioteca musical crua |
| `GET` | `/files/videos` | Biblioteca de vídeos crua |
| `GET` | `/files/stream/:id` | Stream de áudio |
| `GET` | `/files/video-stream/:id` | Stream de vídeo |

### Subgrupo `/files/music`

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/files/music/artists` | Lista artistas a partir de arquivos |
| `GET` | `/files/music/artists/:name` | Faixas por artista |
| `GET` | `/files/music/albums` | Lista álbuns |
| `GET` | `/files/music/albums/:name` | Faixas por álbum |
| `GET` | `/files/music/genres` | Lista gêneros |
| `GET` | `/files/music/genres/:name` | Faixas por gênero |
| `GET` | `/files/music/folders` | Lista pastas musicais |

## 2.2 Diary

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/diary/` | Lista entradas do diário |
| `GET` | `/diary/summary` | Resumo recente |
| `POST` | `/diary/` | Cria entrada |
| `PUT` | `/diary/:id` | Atualiza entrada |
| `POST` | `/diary/copy` | Duplica entrada |

## 2.3 Music playlists

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/music/playlists/` | Lista playlists |
| `POST` | `/music/playlists/` | Cria playlist |
| `GET` | `/music/playlists/now-playing` | Obtém ou cria fila now-playing |
| `GET` | `/music/playlists/system` | Playlists automáticas |
| `GET` | `/music/playlists/:id` | Busca playlist por ID |
| `PUT` | `/music/playlists/:id` | Atualiza playlist |
| `DELETE` | `/music/playlists/:id` | Remove playlist |
| `GET` | `/music/playlists/:id/tracks` | Lista tracks da playlist |
| `POST` | `/music/playlists/:id/tracks` | Adiciona track |
| `DELETE` | `/music/playlists/:id/tracks/:fileId` | Remove track |
| `PUT` | `/music/playlists/:id/tracks/reorder` | Reordena tracks |

## 2.4 Music library

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/music/library` | Lista faixa a faixa |
| `GET` | `/music/library/` | Variante equivalente |
| `GET` | `/music/library/home` | Home catalog musical |
| `GET` | `/music/library/artists` | Catálogo de artistas |
| `GET` | `/music/library/artists/:key/tracks` | Faixas por artista |
| `GET` | `/music/library/albums` | Catálogo de álbuns |
| `GET` | `/music/library/albums/:key/tracks` | Faixas por álbum |
| `GET` | `/music/library/genres` | Catálogo de gêneros |
| `GET` | `/music/library/genres/:key/tracks` | Faixas por gênero |
| `GET` | `/music/library/folders` | Catálogo por pasta |
| `GET` | `/music/library/folders/:key/tracks` | Faixas por pasta |

## 2.5 Music player-state

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/music/player-state/` | Estado do player por cliente/IP |
| `PUT` | `/music/player-state/` | Atualiza estado do player |

## 2.6 Configuration

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/configuration/translation` | Arquivo JSON de tradução atual |
| `GET` | `/configuration/about` | Informações de runtime/build |
| `GET` | `/configuration/settings` | Preferências persistidas |
| `PUT` | `/configuration/settings` | Atualiza preferências |

## 2.7 Video playback

| Método | Rota | Uso real |
| --- | --- | --- |
| `POST` | `/video/playback/start` | Inicia sessão de reprodução |
| `GET` | `/video/playback/state` | Retorna sessão atual |
| `PUT` | `/video/playback/state` | Atualiza progresso/estado |
| `POST` | `/video/playback/next` | Avança |
| `POST` | `/video/playback/previous` | Volta |
| `POST` | `/video/playback/behavior` | Registra evento comportamental |

## 2.8 Video catalog

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/video/catalog/home` | Seções do home de vídeo |

## 2.9 Video library

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/video/library/files` | Lista paginada da biblioteca de vídeo |

## 2.10 Video playlists

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/video/playlists/` | Lista playlists |
| `GET` | `/video/playlists` | Variante equivalente |
| `GET` | `/video/playlists/memberships` | Relação playlist-vídeo |
| `POST` | `/video/playlists/rebuild` | Reconstrói playlists inteligentes |
| `GET` | `/video/playlists/unassigned` | Vídeos sem playlist |
| `PUT` | `/video/playlists/:id/reorder` | Reordena itens |
| `GET` | `/video/playlists/:id` | Detalhe da playlist |
| `PUT` | `/video/playlists/:id` | Renomeia playlist |
| `PUT` | `/video/playlists/:id/hidden` | Oculta/exibe playlist |
| `POST` | `/video/playlists/:id/videos` | Adiciona vídeo manualmente |
| `DELETE` | `/video/playlists/:id/videos/:videoId` | Remove vídeo |

## 2.11 Analytics

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/analytics/overview` | Painel consolidado de analytics |

## 2.12 Jobs

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/jobs/:id` | Detalhe do job |
| `GET` | `/jobs` | Lista jobs |
| `GET` | `/jobs/:id/steps` | Lista steps |
| `POST` | `/jobs/:id/cancel` | Cancela job |

## 2.13 Update

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/update/status` | Consulta release mais recente |
| `POST` | `/update/apply` | Baixa e aplica atualização |

## 2.14 Search

| Método | Rota | Uso real |
| --- | --- | --- |
| `GET` | `/search/global` | Busca global unificada |

## 3. Parâmetros e comportamentos mais relevantes

## 3.1 Files tree

Query params usuais:

- `page`
- `page_size`
- `file_parent`
- `category`

Categorias usadas no frontend:

- `all`
- `starred`
- `recent`

## 3.2 Files by path

`/files/path` usa path relativo do ponto de vista da UI e o backend converte para path absoluto com base em `ENTRY_POINT`.

## 3.3 Upload

Multipart fields:

- `files`
- `target_path`

Resposta:

- `message`
- `uploaded`
- `job_id`

## 3.4 Music player state

O backend usa `c.ClientIP()` como `client_id`.

Impacto:

- o estado é segmentado por IP do cliente, não por autenticação de usuário.

## 3.5 Video playback state

Também é segmentado por `c.ClientIP()`.

## 3.6 Search global

Query params:

- `q`
- `limit`

Limites reais:

- default `6`
- máximo `12`

## 3.7 Analytics

`period` aceito:

- `24h`
- `7d`
- `30d`
- `90d`

## 4. Serviços frontend por domínio

## 4.1 `src/service/files.ts`

Expõe:

- árvore;
- recentes;
- path;
- favorite toggle;
- reindexação;
- upload;
- criar/mover/copiar/renomear/deletar;
- blob;
- catálogos de música e imagem.

## 4.2 `src/service/music.ts`

Expõe:

- home musical;
- artistas;
- álbuns;
- gêneros;
- pastas;
- todas as faixas;
- tracks por agrupamento.

## 4.3 `src/service/playlist.ts`

Expõe:

- CRUD de playlists musicais;
- now playing;
- playlists automáticas;
- tracks por playlist;
- add/remove track.

## 4.4 `src/service/videoPlayback.ts`

Expõe:

- start/update/get state;
- next/previous;
- home catalog;
- playlists;
- memberships;
- rename;
- hide;
- reorder;
- add/remove;
- unassigned;
- biblioteca de vídeo.

## 4.5 `src/service/configuration.ts`

Expõe:

- about;
- translations;
- settings get/update.

## 4.6 `src/service/search.ts`

Expõe:

- busca global.

## 4.7 `src/service/analytics.ts`

Expõe:

- overview de analytics.

## 4.8 `src/service/update.ts`

Expõe:

- `getUpdateStatus`
- `applyUpdate`

Observação:

- o cliente existe, mas não foi encontrado uso evidente em tela da aplicação.

## 5. Atalhos e navegação derivada

## 5.1 Global search

- `Ctrl+K` ou `Cmd+K`

## 5.2 Music

Seções:

- `home`
- `playlists`
- `artists`
- `albums`
- `genres`
- `folders`

## 5.3 Images

Seções:

- `library`
- `recent`
- `captures`
- `photos`
- `folders`
- `albums`

## 5.4 Videos

Seções:

- `home`
- `continue`
- `series`
- `movies`
- `personal`
- `clips`
- `folders`

## 5.5 Analytics

Seções:

- `overview`
- `library`
