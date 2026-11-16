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

## Critérios de aceite

- [ ] `05/CA-10` — playlist por regra "gênero = Rock ou Metal" lista todas as faixas desses gêneros; uma faixa nova de Rock entra sem ação manual.
- [ ] Uma faixa removida da playlist por regra não volta depois de reindexar nem depois de adicionar outras faixas.
- [ ] A prévia mostra o mesmo número que a playlist terá depois de criada.
- [ ] Tocar e Embaralhar a playlist por regra seguem `05/RN-08` a `RN-10` (só faixas dela, todas, sem repetir).
- [ ] Tentar adicionar faixa a uma playlist por regra mostra a mensagem do servidor.
- [ ] Playlists manuais existentes continuam iguais (aparecem como `manual`).
- [ ] Testes: resolução da regra contra Postgres real (`any` e `all`, exclusão, pasta com subpasta); seam test das mutações; render sem mock do formulário.
- [ ] `make ci` verde.

## Fora de escopo

- Playlists sugeridas por IA (`05/D-04`).
- Regras com negação ("tudo menos…") ou limite de quantidade.
- Criar/editar playlist por regra no Android (o app lista e toca, pela mesma API).

## Depende de

Task 40 (a regra precisa enxergar os gêneros corrigidos).
