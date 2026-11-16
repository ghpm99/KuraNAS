# 40 — Música: corrigir metadados das faixas

**Tipo:** feature · **Prioridade:** P0 (N5) · **Spec:** [05](../especificacao/05-musica.md) `RN-02`, `RN-03`, `RN-04`, `RF-14`, `RF-15`, `RF-16`, `CA-11`, `CA-12`

## Contexto

O dono tem pastas de música sem gênero nas tags. Hoje o gênero, o artista e o álbum vêm só da tag do arquivo (`audio_metadata`) e não há como corrigir pelo aplicativo: a faixa fica em "Desconhecido" e não entra em "Gêneros → Rock". Não existe endpoint de edição (spec 05 §8, "A criar").

## Objetivo

Corrigir título, artista, álbum, artista do álbum, gênero(s), ano, disco e número de uma faixa — ou de várias de uma vez — pelo aplicativo, com a correção valendo em todas as listas e sobrevivendo à reindexação.

## O que fazer

Uma alteração por item, cada uma em commit próprio, nesta ordem.

1. **Migração** `audio_metadata_override`: uma linha por `file_id` com as colunas dos oito campos, todas anuláveis (nulo = usa a tag). Nada é gravado no arquivo de áudio (decisão provisória `05/D-01`).
2. **Leitura**: toda consulta de música que hoje lê título/artista/álbum/artista do álbum/gênero/ano/disco/número de `audio_metadata` passa a usar `COALESCE(override, tag)`. Vale para faixas, artistas, álbuns, gêneros, busca e filas. Gêneros múltiplos seguem `RN-04` (separadores `;`, `/`, `,`; comparação sem maiúsculas, acentos e espaços).
3. **`PUT /music/tracks/:file_id/metadata`** conforme o contrato de 05 §8 (campo ausente não muda; vazio ou `null` volta à tag; validações `400`; `404`).
4. **`PUT /music/tracks/metadata`** (lote) com `file_ids` | `album_key` | `folder_key`, `fields` e `?dry_run=true`, respondendo `matched`/`updated`/`failed`.
5. **`PUT /music/library/genres/:key`** — renomear e mesclar gênero.
6. **Web — "Editar informações"** no menu da faixa (`RF-14`): diálogo com os oito campos, sugestão de gêneros e artistas existentes, erro do servidor no campo. Ao salvar, as listas de música abertas refletem a mudança sem recarregar (`00/RN-35`).
7. **Web — edição em lote** (`RF-15`) a partir de seleção, álbum e pasta: só campos preenchidos são enviados; antes de confirmar, mostra a quantidade vinda de `dry_run`; ao fim, mostra o resultado por item (`00/RN-27`).
8. **Web — renomear/mesclar gênero** na tela do gênero (`RF-16`), com confirmação que diz quantas faixas mudam.

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
