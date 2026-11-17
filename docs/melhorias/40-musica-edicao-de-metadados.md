# 40 — Música: corrigir metadados das faixas

**Tipo:** feature · **Prioridade:** P0 (N5) · **Spec:** [05](../especificacao/05-musica.md) `RN-02`, `RN-03`, `RN-04`, `RF-14`, `RF-15`, `RF-16`, `CA-11`, `CA-12`

## Contexto

O dono tem pastas de música sem gênero nas tags. Hoje o gênero, o artista e o álbum vêm só da tag do arquivo (`audio_metadata`) e não há como corrigir pelo aplicativo: a faixa fica em "Desconhecido" e não entra em "Gêneros → Rock". Não existe endpoint de edição (spec 05 §8, "A criar").

Conferido no código em 2026-11-17:

- **Agrupamento não lê a tag direto.** Artistas, álbuns e gêneros (`/music/library/*` — grupos, resumos, faixas e filas por grupo, cluster de IA) filtram e agrupam pelas colunas persistidas `catalog_artist_key/label`, `catalog_album_key/label/base_key/artist_label`, `catalog_genre_keys/labels` de `audio_metadata` (migrações `0059`, `0062`), calculadas em Go por `BuildCatalogGroupingKeys` (`music/catalog_grouping_keys.go`) no upsert da indexação (`upsert_audio_metadata.sql`, `ON CONFLICT ... DO UPDATE` regrava tags e chaves), no backfill (`step_audio_catalog_keys_backfill.go`) e no `reconcile_audio_album_groupings.sql` (compilações, álbuns homônimos). Um `COALESCE` só no SQL não tira a faixa de "Desconhecido": as chaves precisam ser recalculadas com os valores efetivos, e o upsert da reindexação precisa respeitar a correção.
- **Leitura de valores exibidos**: `get_library_tracks`, `get_library_files_by_ids` (hidrata quase todas as listas), `search_library_tracks`, `get_music` (`/files/music`, Android), filas (`get_library_queue_by_file_ids`, `get_playlist_queue`, `get_player_queue`), `get_playlist_tracks`, `get_audio_summary_by_file_id`, a navegação legada `/files/music/{artists,albums,genres}*` (agrupa pela tag crua; o frontend não usa, o Android só usa `/files/music`) e a busca global (`search/search_{tracks,artists,albums}.sql`, com índices trigram `kuranas_fold` das migrações `0054`–`0056`).
- **RN-04 hoje é parcial**: separadores de gênero são `;`, `,` e `|` — a `/` não separa (e há o caso especial `r&b/soul`); maiúsculas e espaços são ignorados, **acentos não** ("Música" ≠ "Musica"). Corrigir muda as chaves de todas as faixas e vira o sub-item 40.1.
- Mover/renomear mantém o `home_file.id` (`update_file.sql`), então a correção acompanha o arquivo; apagar de vez o arquivo apaga a correção (`ON DELETE CASCADE`).
- Próxima migração livre: `0073`.

## Objetivo

Corrigir título, artista, álbum, artista do álbum, gênero(s), ano, disco e número de uma faixa — ou de várias de uma vez — pelo aplicativo, com a correção valendo em todas as listas e sobrevivendo à reindexação.

## O que fazer

Uma alteração por item, cada uma em commit próprio, nesta ordem.

1. **Migração** `0073` `audio_metadata_override`: uma linha por `file_id` (FK para `home_file` com `ON DELETE CASCADE`) com as colunas dos oito campos, todas anuláveis (nulo = usa a tag). Junto, a view `audio_metadata_effective` com as mesmas colunas de `audio_metadata` e `COALESCE(override, tag)` nos oito campos. Nada é gravado no arquivo de áudio (decisão provisória `05/D-01`).
2. **Leitura e chaves**: as consultas de valores exibidos (lista do Contexto) leem `audio_metadata_effective`; a busca mantém o uso de índice (trigram também nas colunas da correção). As chaves `catalog_*` passam a ser calculadas dos valores efetivos por uma função única, chamada ao salvar a correção, no upsert da indexação, no backfill e antes do `reconcile` de álbuns — reindexar não desfaz a correção.
3. **`PUT /music/tracks/:file_id/metadata`** conforme o contrato de 05 §8 (campo ausente não muda; vazio ou `null` volta à tag; validações `400`; `404`).
4. **`PUT /music/tracks/metadata`** (lote) com `file_ids` | `album_key` | `folder_key`, `fields` e `?dry_run=true`, respondendo `matched`/`updated`/`failed`.
5. **`PUT /music/library/genres/:key`** — renomear e mesclar gênero.
6. **Web — "Editar informações"** no menu da faixa (`RF-14`): diálogo com os oito campos, sugestão de gêneros e artistas existentes, erro do servidor no campo. Ao salvar, as listas de música abertas refletem a mudança sem recarregar (`00/RN-35`).
7. **Web — edição em lote** (`RF-15`) a partir de seleção, álbum e pasta: só campos preenchidos são enviados; antes de confirmar, mostra a quantidade vinda de `dry_run`; ao fim, mostra o resultado por item (`00/RN-27`).
8. **Web — renomear/mesclar gênero** na tela do gênero (`RF-16`), com confirmação que diz quantas faixas mudam.

### Sub-itens

- [ ] 40.1. `RN-04` completa nas chaves de gênero: `/` também separa (decidir o caso `r&b/soul`) e acentos não diferenciam; recalcula as chaves de todas as faixas. O diálogo web (item 6) já separa os gêneros iniciais por `;`, `/` e `,` (`trackMetadataFormValues.ts`) — alinhar com a regra final do backend. A faixa tocando agora no player não atualiza título/artista após a edição (fila em memória do `GlobalMusicProvider`) — fica para depois.

## Critérios de aceite

- [ ] Alterar o gênero de uma faixa sem tag a move de "Desconhecido" para o gênero novo em Faixas, Gêneros e busca, sem recarregar a página.
- [ ] `05/CA-11` — alterar o gênero de 30 faixas em lote as move para o gênero novo; reindexar a biblioteca não desfaz.
- [ ] `05/CA-12` — mesclar dois gêneros deixa um só, com a soma das faixas.
- [ ] Limpar um campo corrigido faz a faixa voltar a mostrar o valor da tag.
- [ ] Editar um álbum inteiro por `album_key` altera faixas que não estavam carregadas na tela.
- [ ] O arquivo de áudio no disco não muda (mesmo checksum antes e depois).
- [ ] Testes: decodificação da resposta real de cada handler; seam test de cada mutação no frontend; render sem mock dos diálogos.
- [ ] `make ci` verde.

## Fora de escopo

- Gravar a correção na tag do arquivo (`05/D-01`).
- Buscar metadados, capas ou letras na internet.
- Edição de metadados no Android (o app passa a **exibir** os valores corrigidos, porque vêm da mesma API; a tela de edição fica para depois).
