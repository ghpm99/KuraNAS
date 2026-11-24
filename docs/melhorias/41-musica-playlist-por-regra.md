# 41 — Música: playlist por regra

**Tipo:** feature · **Prioridade:** P0 (N5) · **Spec:** [05](../especificacao/05-musica.md) `RN-18`, `RN-20`, `RN-21`, `RF-07`, `RF-11`, `RF-12`, `CA-10`

## Contexto

O dono quer playlists "que só têm o que ele escolheu" (N5), inclusive por critério: "tudo de Metal e Hard Rock". Hoje só existem playlists manuais (`playlist`, `playlist_track`) e as do sistema. Não há regra, nem prévia, nem "Criar playlist deste gênero".

## Objetivo

Criar uma playlist definida por critérios, cujo conteúdo é calculado a cada leitura e acompanha a biblioteca sozinho.

## O que fazer

1. **Migração**: `playlist` ganha `kind` (`manual` — padrão para as existentes —, `rule`, `system`) e `rules` (JSON, nulo fora de `rule`); tabela `playlist_rule_exclusion` (`playlist_id`, `file_id`) para as faixas tiradas à mão (`RN-20`).
2. **Resolução da regra**: um `.sql` que devolve as faixas de uma regra, aplicando `match` (`any`/`all`) sobre gêneros, artistas, álbuns, pastas (com subpastas), intervalo de anos e só favoritas, já descontando as exclusões e usando os metadados corrigidos da task 40. Ordenação: título, desempate por `id`.
3. **`POST /music/playlists/rules/preview`** → `{"track_count"}`.
4. **`POST /music/playlists` e `PUT /music/playlists/:id`** aceitam `rules`; o item de playlist devolve `kind` e `rules`. Regra sem critério → `400`.
5. **Leitura**: `GET /music/playlists/:id/tracks` e `/:id/queue` de uma playlist `rule` devolvem o resultado da regra, paginado. A contagem e a duração da lista de playlists usam o mesmo cálculo.
6. **Escrita bloqueada**: em playlist `rule`, adicionar faixa e reordenar respondem `409` com mensagem própria; `DELETE …/tracks/:fileId` grava a exclusão.
7. **Web — criar/editar por regra** (`RF-11`): formulário de critérios com seleção de gêneros, artistas, álbuns e pastas vindos do servidor, "qualquer"/"todos", e a prévia da quantidade atualizada a cada mudança (espera de 300 ms).
8. **Web — tela da playlist `rule`** (`RF-12`): mostra os critérios, não oferece reordenar nem adicionar; "Remover" continua disponível.
9. **Web — "Criar playlist deste gênero"** na tela do gênero (`RF-07`): abre o formulário de regra com o gênero preenchido.

## Decisões de desenho (2026-11-24)

- **`kind`** vira coluna (`manual`, `rule`, `system`, `ai`), com backfill pelas flags `is_system`/`is_ai_generated`, que continuam existindo; as automáticas (id negativo) seguem saindo como `automatic`. O DTO passa a ler `kind` da coluna e ganha `rules` (nulo fora de `rule`) e `total_length_seconds`.
- **Formato da regra**: o do contrato da spec 05 §8. Gêneros, artistas e álbuns chegam como nome ou chave e o servidor os normaliza com a mesma função das chaves de catálogo, comparando com `catalog_genre_keys`, `catalog_artist_key` e `catalog_album_key` de `audio_metadata_effective` (metadados corrigidos da task 40). Dentro de um critério a lista é "qualquer um"; `match` combina os critérios (gêneros, artistas, álbuns, pastas, intervalo de anos, só favoritas).
- **Resolução**: um `.sql` parametrizado que devolve as faixas da regra (`hf.format` de áudio, `deleted_at IS NULL`, menos `playlist_rule_exclusion`), ordenado por título e `id`; contagem e duração pelo mesmo predicado. `tracks` pagina sobre ele; `queue` usa o mesmo limite das outras filas.
- **Escrita**: adicionar e reordenar em `rule` → `409` com chave i18n própria; remover grava a exclusão.

## Critérios de aceite

- [x] `05/CA-10` — playlist por regra "gênero = Rock ou Metal" lista todas as faixas desses gêneros; uma faixa nova de Rock entra sem ação manual. *Backend local: regra `genres=["Rock","Blues"]` → 8 faixas (6 Rock + 2 Blues, como `/music/library/genres`); corrigir o gênero da faixa 872 (sem gênero) para Rock pela edição de metadados fez a contagem ir de 7 para 8 e a faixa aparecer em `tracks`, sem tocar na playlist.*
- [x] Uma faixa removida da playlist por regra não volta depois de reindexar nem depois de adicionar outras faixas. *Backend local: `DELETE …/tracks/873` → 200, contagem 8 → 7, fila com 7; depois da faixa 872 entrar, a 873 continua fora. Reindexação (novo upsert de `audio_metadata`) coberta por teste em Postgres.*
- [x] A prévia mostra o mesmo número que a playlist terá depois de criada. *Backend local: `rules/preview` → `{"track_count":8}` e o `POST` com a mesma regra → `track_count: 8`.*
- [x] Tocar e Embaralhar a playlist por regra seguem `05/RN-08` a `RN-10` (só faixas dela, todas, sem repetir). *Backend local: `/queue` devolve 7 itens, 7 distintos, `truncated: false`, iguais às faixas da regra; Embaralhar é a mesma fila embaralhada no cliente.*
- [x] Tentar adicionar faixa a uma playlist por regra mostra a mensagem do servidor. *Backend local: adicionar e reordenar → `409 {"error":"Playlist por regra não aceita adicionar nem reordenar faixas"}`; a web não oferece playlist por regra no menu e mostra a mensagem do servidor quando recebe o 409 (testes).*
- [x] Playlists manuais existentes continuam iguais (aparecem como `manual`). *Backend local: playlist criada sem regra sai `kind: manual`, `rules: null`, e aceita adicionar faixa (201); migração com backfill testada (manual, system, ai).*
- [x] Testes: resolução da regra contra Postgres real (`any` e `all`, exclusão, pasta com subpasta); seam test das mutações; render sem mock do formulário.
- [x] `make ci` verde.

## Fora de escopo

- Playlists sugeridas por IA (`05/D-04`).
- Regras com negação ("tudo menos…") ou limite de quantidade.
- Criar/editar playlist por regra no Android (o app lista e toca, pela mesma API).

## Depende de

Task 40 (a regra precisa enxergar os gêneros corrigidos).
