# 05 — Música

**Prioridade:** P0 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Ouvir a coleção própria de músicas com a comodidade de um serviço de streaming — por gênero, artista, álbum e playlist — com uma diferença essencial: **só toca o que o dono escolheu**.

## 2. Cenários de uso

- O dono tem muitas pastas de música soltas. Quer abrir "Gêneros → Rock" e ouvir só rock, embaralhado, pelo tempo que quiser.
- Monta playlists à mão e por regra ("tudo de Metal e Hard Rock") e nenhuma faixa de outro estilo entra nelas.
- Ouve no PC enquanto trabalha e continua no celular, com a tela bloqueada, controlando pela notificação.
- Encontra uma pasta de músicas sem gênero nas tags e corrige o gênero de todas de uma vez.

## 3. Escopo

Biblioteca (faixas, artistas, álbuns, gêneros, pastas), playlists manuais e por regra, fila, player, reprodução contínua, estado por dispositivo, capas, letras, histórico e mais tocadas, edição de metadados, busca dentro da música.

## 4. Fora de escopo

- Recomendação automática, "rádio" que insere faixas não pedidas, autoplay de conteúdo fora do contexto escolhido.
- Fontes externas de streaming; busca de metadados, capas ou letras em serviços da internet.
- Download de músicas da internet (isso é [11](11-download-por-url.md)).
- Equalizador, normalização de volume, crossfade.
- Podcasts e audiolivros como categorias próprias.
- Sincronizar a reprodução ao vivo entre dispositivos (tocar no PC e controlar pelo celular).
- Download de faixas para ouvir offline no Android. *(ver D-03)*

## 5. Conceitos

- **Faixa**: um arquivo de áudio com seus metadados (título, artista, álbum, artista do álbum, gênero, ano, disco, número).
- **Contexto de reprodução**: de onde a fila veio — um álbum, um artista, um gênero, uma pasta, uma playlist, a lista de todas as faixas ou um resultado de busca.
- **Fila**: sequência de faixas em reprodução no dispositivo.
- **Playlist manual**: lista ordenada de faixas escolhidas uma a uma.
- **Playlist por regra**: lista calculada por critérios (gêneros, artistas, anos, pastas, favoritas); atualiza-se quando a biblioteca muda.
- **Playlists do sistema**: Favoritas, Mais tocadas, Tocadas recentemente, Adicionadas recentemente.

## 6. Regras de negócio

### Biblioteca

- **RN-01** — É faixa todo arquivo de áudio reconhecido: MP3, FLAC, M4A/AAC/ALAC, OGG/OGA, Opus, WAV, WMA. Toda faixa indexada aparece em "Faixas", com ou sem tags.
- **RN-02** — Metadados vêm das tags do arquivo, em qualquer formato de tag. Sem tag: título = nome do arquivo sem extensão; artista, álbum e gênero = "Desconhecido" (texto traduzido). Faixas sem tag continuam alcançáveis em "Faixas" e em "Pastas".
- **RN-03** — Metadado corrigido pelo dono no aplicativo prevalece sobre a tag do arquivo e não é desfeito por reindexação.
- **RN-04** — Uma faixa pode ter vários gêneros (tag com separadores `;`, `/`, `,`). Gêneros que só diferem em maiúsculas, acentos ou espaços são o mesmo gênero.
- **RN-05** — Álbum é identificado por nome do álbum + artista do álbum. Álbum com faixas de artistas diferentes e sem artista do álbum é "Vários artistas". Álbuns de mesmo nome e artistas diferentes são álbuns distintos.
- **RN-06** — Dentro de um álbum as faixas seguem disco e número da faixa (numéricos); nas demais listas, o padrão é ordem alfabética de título. "Adicionadas recentemente" é por data de adição, mais recente primeiro.
- **RN-07** — Capa: imagem embutida no arquivo; senão, `cover`/`folder`/`front` (JPG ou PNG) na pasta; senão, capa genérica.

### Reprodução

- **RN-08** — **O sistema só toca faixas do contexto que o dono escolheu.** Ao acabar a fila, a reprodução para (ou repete, se repetição estiver ligada). Nada é adicionado à fila automaticamente.
- **RN-09** — "Tocar" em um contexto monta a fila com **todas** as faixas dele, na ordem exibida, começando pela faixa clicada. "Embaralhar" monta a fila com todas, em ordem aleatória. Não há limite de tamanho que corte a fila em silêncio.
- **RN-10** — Embaralhar é uma permutação: nenhuma faixa repete antes de todas terem tocado; "anterior" volta de verdade para a faixa que tocou antes.
- **RN-11** — Repetição: desligada, repetir a fila, repetir a faixa.
- **RN-12** — "Anterior" depois de 3 s de faixa volta ao início dela; antes disso, vai para a faixa anterior.
- **RN-13** — Faixa que não toca (arquivo ausente, formato inválido) é pulada com aviso; após 3 falhas seguidas a reprodução para e explica.
- **RN-14** — Formato que o dispositivo não reproduz é convertido pelo servidor na hora, sem ação do usuário.
- **RN-15** — Uma reprodução conta para histórico e "mais tocadas" depois de 30 s ou metade da faixa, o que vier primeiro.
- **RN-16** — O estado do player — fila, faixa atual, posição, volume, embaralhar, repetição — é guardado **por dispositivo** e restaurado ao reabrir o aplicativo, sem começar a tocar sozinho.
- **RN-17** — A música continua tocando ao navegar entre telas e, no Android, com o app em segundo plano e a tela bloqueada.

### Playlists

- **RN-18** — Nome de playlist obrigatório e único. Playlists são globais (valem em todos os dispositivos).
- **RN-19** — Playlist manual mantém a ordem definida pelo dono. A mesma faixa pode ser adicionada duas vezes só com confirmação.
- **RN-20** — Playlist por regra não aceita adição manual; mostra os critérios e o resultado atual. Faixas excluídas explicitamente dela pelo dono não voltam.
- **RN-21** — Playlists do sistema não podem ser renomeadas nem excluídas.
- **RN-22** — Faixa movida ou renomeada continua nas playlists. Faixa excluída sai das playlists; restaurada da lixeira, volta.
- **RN-23** — Excluir uma playlist não exclui faixas.

## 7. Requisitos por tela

Todas as listas abaixo seguem [00](00-padroes-transversais.md): paginadas, scroll infinito, recarregar, ordenação no servidor, estados.

### Início da música (`/music`)

- **RF-01** — Blocos: Continuar ouvindo (a fila do dispositivo), Tocadas recentemente, Mais tocadas, Adicionadas recentemente, Playlists, atalho para Gêneros. Cada bloco carrega e falha de forma independente.

### Faixas (`/music/tracks`)

- **RF-02** — Todas as faixas, em lista: capa, título, artista, álbum, gênero, duração, estrela. Ordenação por título, artista, álbum, gênero, duração, data de adição, mais tocadas.
- **RF-03** — Clique em uma faixa toca a partir dela, com a lista como contexto. Menu da faixa: Tocar a seguir, Adicionar ao fim da fila, Adicionar a playlist, Ir para o álbum, Ir para o artista, Favoritar, Editar informações, Detalhes, Baixar, Excluir.
- **RF-04** — Seleção múltipla com: Tocar, Adicionar à fila, Adicionar a playlist, Favoritar, Editar informações em lote, Excluir.

### Artistas, álbuns, gêneros, pastas

- **RF-05** — **Artistas**: grade com nome e contagens. A tela do artista mostra os álbuns e todas as faixas, com Tocar e Embaralhar.
- **RF-06** — **Álbuns**: grade com capa, nome, artista, ano. A tela do álbum mostra capa, ano, duração total e as faixas por disco/número, com Tocar e Embaralhar.
- **RF-07** — **Gêneros**: grade com nome e contagem de faixas. A tela do gênero mostra todas as faixas do gênero, com Tocar, Embaralhar e "Criar playlist deste gênero".
- **RF-08** — **Pastas**: navegação pela estrutura de pastas que contêm áudio, com Tocar e Embaralhar a pasta (incluindo subpastas).
- **RF-09** — Nomes com `/`, `&`, acentos ou caracteres especiais ("AC/DC") abrem normalmente.

### Playlists (`/music/playlists`)

- **RF-10** — Lista: playlists do sistema primeiro, depois as do dono, com capa (mosaico das capas das faixas), nome, quantidade e duração total.
- **RF-11** — Criar playlist manual (nome) ou por regra (critérios: gêneros, artistas, álbuns, pastas, anos, só favoritas; combinados com "qualquer" ou "todos"; com prévia da quantidade de faixas).
- **RF-12** — Tela da playlist: Tocar, Embaralhar, renomear, excluir (com confirmação), lista de faixas. Em playlist manual: reordenar arrastando, remover faixa, adicionar faixas por busca.
- **RF-13** — "Adicionar a playlist" a partir de qualquer faixa, álbum, artista, gênero, pasta ou seleção: menu com busca das playlists e "Nova playlist".

### Edição de metadados

- **RF-14** — "Editar informações" de uma faixa: título, artista, álbum, artista do álbum, gênero(s), ano, disco, número. Gênero e artista têm sugestão dos valores já existentes.
- **RF-15** — Edição em lote (seleção, álbum inteiro, pasta inteira): só os campos preenchidos são aplicados; os demais ficam como estão. Mostra quantas faixas serão alteradas antes de confirmar.
- **RF-16** — Gêneros podem ser renomeados e mesclados (todas as faixas de "Rock n Roll" passam a "Rock").

### Busca (`/music/search`)

- **RF-17** — Busca em faixas, artistas, álbuns, gêneros e playlists, agrupada por tipo, cada grupo com "ver todos" paginado. Ignora maiúsculas e acentos.

### Player

- **RF-18** — Barra persistente: capa, título, artista, anterior, tocar/pausar, próxima, barra de progresso com tempo decorrido e total, volume, embaralhar, repetição, estrela, abrir fila.
- **RF-19** — Arrastar a barra de progresso só reposiciona ao soltar.
- **RF-20** — Player expandido: capa grande, controles completos, letra (quando o arquivo tem), fila.
- **RF-21** — Fila: lista das próximas, com a atual destacada; reordenar arrastando, remover, limpar, "Salvar fila como playlist". Clicar em uma faixa da fila pula para ela.
- **RF-22** — Controles do sistema operacional: teclas de mídia do teclado, notificação/tela de bloqueio (título, artista, capa, anterior, tocar/pausar, próxima).
- **RF-23** — Atalhos na web: espaço toca/pausa, setas ±10 s, Shift+setas anterior/próxima, `m` mudo. Não disparam com o foco em campo de texto.

### Android

- **RF-24** — As mesmas listas, paginadas; player com reprodução em segundo plano, notificação com controles, controles de fone/Bluetooth, pausa ao desconectar o fone e ao receber chamada; fila, playlists (criar, adicionar, reordenar), favoritar.

## 8. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/music/library` · `/music/library/tracks/:file_id` | faixas |
| GET | `/music/library/home` · `/most-played` · `/recent-plays` | início |
| GET | `/music/library/artists[/:key[/albums\|tracks\|queue]]` | artistas |
| GET | `/music/library/albums[/:key[/tracks\|queue\|cover]]` | álbuns |
| GET | `/music/library/genres[/:key[/tracks\|queue]]` | gêneros |
| GET | `/music/library/folders[/:key[/tracks\|queue]]` | pastas |
| GET | `/music/search` | busca |
| GET | `/music/tracks/:file_id/stream` · `/cover` | áudio e capa |
| POST | `/music/plays` | registrar reprodução |
| GET, POST | `/music/playlists` | listar, criar |
| GET | `/music/playlists/system` · `/now-playing` | playlists do sistema |
| GET, PUT, DELETE | `/music/playlists/:id` | detalhe, editar, excluir |
| GET, POST | `/music/playlists/:id/tracks` · GET `/:id/queue` | faixas da playlist |
| DELETE | `/music/playlists/:id/tracks/:fileId` | remover faixa |
| PUT | `/music/playlists/:id/tracks/reorder` | reordenar |
| GET, PUT | `/music/player-state` · `/music/player-state/queue` | estado por dispositivo |

**A criar** (não existem hoje):

| Método | Rota | Uso |
|---|---|---|
| PUT | `/music/tracks/:file_id/metadata` | corrige os metadados de uma faixa (RN-03, RF-14) |
| PUT | `/music/tracks/metadata` | corrige em lote (RF-15) |
| PUT | `/music/library/genres/:key` | renomeia ou mescla um gênero (RF-16) |
| POST | `/music/playlists/rules/preview` | conta as faixas que uma regra seleciona (RF-11) |

- **Corpo de `PUT /music/tracks/:file_id/metadata`**: qualquer subconjunto de `title`, `artist`, `album`, `album_artist`, `genres` (array de textos), `year`, `disc_number`, `track_number`. Campo ausente não muda; texto vazio ou `null` volta o campo ao valor da tag do arquivo. Resposta `200` com a faixa no mesmo formato de `/music/library/tracks/:file_id`. `year`, `disc_number` e `track_number` são inteiros. `400` para ano fora de 1000–9999, número de disco/faixa menor que 1 ou texto acima de 255 caracteres, com corpo `{"error": "...", "field": "year"}` (`field` = nome do campo recusado, para a tela marcar o campo); `404` para faixa inexistente.
- **Corpo de `PUT /music/tracks/metadata`**: `fields` (os mesmos campos acima) e exatamente um seletor — `file_ids` (até 500), `album_key` ou `folder_key` (inclui subpastas). Com `?dry_run=true` nada é gravado. Resposta `200`: `{"matched": 30, "updated": 30, "failed": [{"file_id": 7, "error": "..."}]}` (em `dry_run`, `updated` é 0). Nenhum ou mais de um seletor → `400`.
- **Corpo de `PUT /music/library/genres/:key`**: `{"name": "Rock"}`. Se o nome novo (ignorando maiúsculas, acentos e espaços — RN-04) é um gênero existente, as faixas são mescladas nele. Resposta `200`: `{"key": "...", "name": "Rock", "track_count": 412, "merged": true}`.
- **Playlist por regra**: `POST /music/playlists` e `PUT /music/playlists/:id` passam a aceitar `rules`; o item de playlist ganha `kind` (`manual`, `rule`, `system`) e `rules` (`null` nas demais). Formato:

```json
{
  "match": "any",
  "genres": ["Rock", "Metal"],
  "artists": [],
  "albums": [],
  "folders": [],
  "year_from": null,
  "year_to": null,
  "favorites_only": false
}
```

  `match` é `any` (basta um critério) ou `all`. Regra sem nenhum critério → `400`. `POST /music/playlists/rules/preview` recebe o mesmo objeto e responde `{"track_count": 412}`.
- Em playlist `rule`: `POST /music/playlists/:id/tracks` e `PUT …/tracks/reorder` respondem `409`; `DELETE …/tracks/:fileId` registra a exclusão explícita (RN-20) e a faixa deixa de aparecer no resultado.

## 9. Estados e erros

- Biblioteca sem música: estado vazio explicando como adicionar (enviar arquivos ou configurar a pasta de música).
- Fila vazia: a barra do player não aparece.
- Faixa indisponível: aviso curto e pula (RN-13).
- Estado do player não pôde ser salvo: a reprodução continua; não bloqueia nem mostra erro a cada tentativa.

## 10. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) em Faixas, Artistas, Álbuns, Gêneros, Pastas, Playlists, faixas de cada um deles e resultados de busca:

- [ ] **CA-01** — Em uma biblioteca com mais de 2 000 faixas, a lista "Faixas" rola até o fim mostrando todas.
- [ ] **CA-02** — MP3, FLAC, M4A, OGG e Opus aparecem com título, artista, álbum, gênero e duração corretos, e tocam.
- [ ] **CA-03** — Uma faixa sem nenhuma tag aparece em "Faixas" com o nome do arquivo como título e toca.
- [ ] **CA-04** — Abrir um gênero e clicar em "Embaralhar" toca apenas faixas daquele gênero, sem repetir, até todas terem tocado; ao terminar, para.
- [ ] **CA-05** — "Tocar" em um gênero com mais de 500 faixas coloca todas na fila.
- [ ] **CA-06** — Deixar uma playlist de 5 faixas tocar até o fim, com repetição desligada: a reprodução para; nenhuma faixa de fora entra.
- [ ] **CA-07** — Artista "AC/DC" abre a tela do artista com seus álbuns e faixas.
- [ ] **CA-08** — Um álbum de dois discos lista disco 1 (faixas 1…n) e depois disco 2, e a faixa 10 vem depois da 9.
- [ ] **CA-09** — Criar playlist, adicionar 20 faixas, reordenar três, remover duas, renomear: tudo persiste após recarregar e aparece no outro dispositivo.
- [ ] **CA-10** — Criar playlist por regra "gênero = Rock ou Metal" lista todas as faixas desses gêneros; adicionar à biblioteca uma faixa nova de Rock a inclui sem ação manual.
- [ ] **CA-11** — Alterar o gênero de 30 faixas em lote as move para o novo gênero; reindexar a biblioteca não desfaz a alteração.
- [ ] **CA-12** — Mesclar dois gêneros deixa um só, com a soma das faixas.
- [ ] **CA-13** — Trocar de tela dez vezes com música tocando não interrompe nem reinicia a faixa.
- [ ] **CA-14** — Fechar o navegador no meio de uma faixa e reabrir: a mesma fila, a mesma faixa, a mesma posição, pausada.
- [ ] **CA-15** — "Anterior" com embaralhar ligado volta para a faixa que tocou antes.
- [ ] **CA-16** — Capas aparecem na lista, no player e na notificação do sistema; álbum sem capa mostra a capa genérica.
- [ ] **CA-17** — As teclas de mídia do teclado controlam o player.
- [ ] **CA-18** — Mover uma faixa de pasta em Arquivos não a remove das playlists nem zera sua contagem de reproduções.
- [ ] **CA-19** — No Android: tocar um álbum, bloquear a tela, trocar de faixa pela notificação e pelo fone Bluetooth; desconectar o fone pausa.
- [ ] **CA-20** — No Android, as listas de artistas, álbuns e faixas rolam além da primeira página.

## 11. Decisões em aberto

- **D-01** — Editar metadados grava só no banco do KuraNAS ou também na tag do arquivo. Gravar no arquivo deixa a correção visível em outros players e sobrevive a uma reinstalação, mas altera o arquivo (muda checksum e backup). *Provisório: só no banco (RN-03).*
- **D-02** — Ordenação padrão de "Faixas": título (como players de música) ou data de adição decrescente (o padrão geral de [00](00-padroes-transversais.md)). *Provisório: título.*
- **D-03** — Ouvir offline no Android (baixar playlists para o aparelho). *Provisório: fora de escopo.*
- **D-04** — Playlists geradas por IA: manter como ação explícita ("sugerir playlist") ou remover. *Provisório: ação explícita, nunca automática, e a sugestão só vira playlist depois que o dono confirma as faixas.*

## 12. Rastreabilidade

`docs/revisao-produto/05-musica.md` (21 lacunas). `docs/melhorias/`: 40 (edição de metadados, RF-14 a RF-16) e 41 (playlist por regra, RF-11). A regra "só toca o que foi escolhido" (RN-08) não tem tarefa: é verificada por `CA-04` a `CA-06`; falha vira lacuna.
