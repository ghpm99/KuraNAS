# 42 — Vídeos: tela da série, com temporadas e correção manual

**Tipo:** feature · **Prioridade:** P0 (N6) · **Spec:** [06](../especificacao/06-videos.md) `RN-02`, `RN-12`, `RN-13`, `RN-17`, `RF-07`, `RF-08`, `CA-07`, `CA-08`, `CA-10`, `CA-15`

## Contexto

O dono abre um anime e quer ver em qual episódio parou, por temporada, e continuar (N6). Hoje a série é uma playlist automática por pasta, sem tela própria com temporadas; quando o sistema erra a classificação ou o número do episódio, não há como corrigir (spec 06 §8, "A criar"; §12: "tela de série com temporadas" sem item no backlog).

## Objetivo

Uma tela por série com seletor de temporada, lista de episódios com progresso, botão "Continuar", e correção manual de classificação e numeração.

## O que fazer

1. **Migração** `video_manual_override` (`file_id`, `classification`, `season`, `episode`, todos anuláveis). A leitura de classificação, temporada e episódio passa a ser `COALESCE(override, automático)`; a indexação nunca escreve nessa tabela (`21/RN-09`).
2. **`PUT /video/library/files/:file_id/classification`** (contrato de 06 §8).
3. **`PUT /video/library/files/:file_id/episode`** (contrato de 06 §8).
4. **`PUT /video/playlists/:id/watched`** — série inteira ou uma temporada; `409` em playlist do dono.
5. **Episódios da série**: `GET /video/playlists/:id/items` passa a aceitar `season` e a devolver, por item, `season`, `episode`, progresso e assistido, ordenados por temporada e episódio em ordem natural (`RN-02`). O item da playlist ganha `seasons` (lista de números com contagem de episódios e de assistidos).
6. **Web — tela da série** (`RF-07`): pôster, nome, total e progresso; botão principal "Continuar T1E5" / "Assistir" / "Rever" conforme `RN-12`; seletor de temporada; lista paginada de episódios com miniatura, número, título, duração, barra de progresso e selo de assistido. O card de série do catálogo abre esta tela (`RF-04`).
7. **Web — ações** (`RF-08`): marcar temporada/série como assistida ou não; favoritar; ocultar; "Corrigir classificação" e "Corrigir episódio" no menu de cada vídeo.
8. **Player**: "próximo" no último episódio de uma temporada abre o primeiro da seguinte; no último da série fica desabilitado (`RN-17`).

## Reconferência no código (2026-11-24)

O backend não guarda temporada nem episódio de arquivos (só de capturas, em `captures`); o classificador só testa padrões por regex, sem extrair números. A web agrupa por temporada no cliente (`parseSeasonEpisode.ts`) sobre os itens já carregados, o que faz "Continuar" errar com mais de uma página. A ordem das playlists automáticas é gravada em `video_playlist_item.order_index` no rebuild, e o player (web e `POST /video/playback/next`) segue essa ordem. Séries é listada por playlist (`video_playlist.classification`), Filmes por vídeo (`video_metadata.classification`). Progresso é por `client_id`.

## Decisões de desenho (2026-11-24)

- **Temporada e episódio automáticos** extraídos em Go no rebuild (função com captura que reaproveita os padrões do classificador: `S02E03`, `2x03`, `Ep 3`/`Episódio 3`, `- 03`; sem temporada → 1; capturas usam a proveniência) e gravados em `video_playlist_item` (`season`, `episode`). Com `video_manual_override` (`file_id`, `classification`, `season`, `episode`), o valor efetivo é `COALESCE(override, automático)`.
- **Uma ordem só**: em playlist automática de série, `order_index` = temporada, episódio, nome natural, id — com os valores efetivos. Corrigir episódio renumera aquela playlist na hora; assim lista, "próximo" do player (web e backend) e `RN-17` usam a mesma ordem e cruzar temporada sai de graça. Playlist do dono mantém a ordem dela.
- **Classificação**: o `PUT` aceita os valores da spec (`series`, `movie`, `personal`, `clip`, `other`) e `null`; leitura de Filmes usa o efetivo na hora; a seção Séries depende da playlist, então o `PUT` dispara a reconstrução das playlists automáticas, que passa a usar a classificação efetiva.
- **Cabeçalho da série** (`GET /video/playlists/:id`) ganha `seasons` (número, episódios, assistidos) e `resume` (vídeo, temporada, episódio, posição e modo `continue`/`start`/`rewatch`, `RN-12`), calculados no servidor para o cliente da requisição. `items` aceita `season` (inválido → `400`).
- **Favoritar/ocultar** (`RF-08`): favoritar é por episódio (estrela do arquivo, já existe) e ocultar é da série (já existe); a tela só os expõe.

## Critérios de aceite

- [ ] `06/CA-07` — série de 24 episódios `Ep 1`…`Ep 24` aparece na ordem 1, 2, 3 … 24.
- [ ] `06/CA-08` — `Serie.S02E03.1080p.x264.mkv` aparece como episódio 3 da temporada 2.
- [ ] `06/CA-10` — "próximo" no último episódio da temporada abre o primeiro da seguinte; no último da série, desabilitado.
- [ ] `06/CA-15` — marcar uma temporada como assistida marca todos os episódios dela e só eles.
- [ ] Corrigir o número de um episódio o reposiciona na lista; reindexar não desfaz.
- [ ] Corrigir a classificação de "filme" para "série" tira o vídeo de Filmes e o coloca em Séries.
- [ ] "Continuar" abre o episódio em andamento no ponto salvo; sem episódio em andamento, abre o próximo não assistido.
- [ ] Os CA de `00` valem na lista de episódios (três páginas, falha de página seguinte mantém a lista).
- [ ] Testes: ordenação e overrides contra Postgres real; seam test das mutações; render sem mock da tela.
- [ ] `make ci` verde.

## Sub-itens

- [ ] **42.1** — o dedup do motor de playlists (`deduplicateCandidates`, sobreposição > 80% com o candidato de maior pontuação) deixa a playlist "mixed" (`random:smart_mix`, até 20 itens) engolir uma série pequena (≤ 12 episódios) num acervo pequeno: a série não vira playlist e não aparece em Séries. Achado no passo B1; afeta `06/CA-07` em bibliotecas pequenas. Playlists de série não devem ser descartadas por sobreposição com playlists de mistura.

## Fora de escopo

- Metadados de serviços externos (sinopse, pôster da internet).
- Faixa de áudio (task 43).
- Tela de série no Android: já existe catálogo; a paridade da tela entra pelo item de Android do board de revisão.
