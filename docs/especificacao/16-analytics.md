# 16 — Analytics

**Prioridade:** P2 · **Plataformas:** web · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Mostrar como o espaço do servidor está sendo usado e onde é possível liberá-lo, e se a biblioteca está indexada e saudável.

## 2. Cenários de uso

- O disco está enchendo: o dono vê quais pastas e quais arquivos ocupam mais e quanto as duplicatas desperdiçam.
- Quer saber se todas as músicas e fotos já foram processadas depois de uma importação grande.
- Quer acompanhar o crescimento do uso ao longo dos meses.

## 3. Escopo

Uso de armazenamento (total, por tipo, por extensão, por pasta, ao longo do tempo), maiores arquivos, duplicados, arquivos recentes, situação da biblioteca (metadados, processamento, falhas), saúde da indexação.

## 4. Fora de escopo

- Monitoramento de hardware (temperatura, SMART, CPU, memória).
- Estatísticas de uso por pessoa ou por dispositivo.
- Exportar relatórios.
- Remoção automática de duplicados.
- Sugestões geradas por IA. *(ver D-01)*

## 5. Regras de negócio

- **RN-01** — Os números consideram todas as raízes habilitadas e os dois tiers; não contam lixeira, área de backup nem temporários. O espaço da lixeira é mostrado à parte.
- **RN-02** — Espaço total e livre são os dos volumes das raízes, somados sem contar duas vezes o mesmo volume.
- **RN-03** — Cada bloco é um pedido independente, pequeno, e carrega e falha por conta própria. A falha de um bloco nunca aparece como zero: aparece como erro com "Tentar de novo".
- **RN-04** — O período escolhido (24 h, 7, 30, 90 dias) vale para os blocos de variação e fica na URL.
- **RN-05** — Duplicados são arquivos de conteúdo idêntico (mesmo checksum). Um grupo mostra todas as cópias com caminho; o espaço recuperável é o tamanho das cópias além da primeira.
- **RN-06** — Todo número que representa uma lista é um link para essa lista, já filtrada: "maiores arquivos" abre os arquivos ordenados por tamanho; uma pasta abre a pasta; um grupo de duplicados abre as cópias; "falhas" abre os itens que falharam.
- **RN-07** — Toda ação tem o efeito que o rótulo diz. "Atualizar" recarrega os números; "Reindexar" inicia a indexação e mostra o job ([21](21-indexacao.md)).
- **RN-08** — Os números podem ter até 5 minutos de atraso; a tela mostra quando foram calculados.

## 6. Requisitos por tela

### Visão geral (`/analytics`)

- **RF-01** — Cabeçalho: período, "atualizado há…", Atualizar.
- **RF-02** — Indicadores: espaço usado, livre e total; crescimento no período; total de arquivos e de pastas; arquivos adicionados no período; espaço por tier; espaço na lixeira.
- **RF-03** — Uso ao longo do tempo (gráfico do período).
- **RF-04** — Uso por tipo (imagens, músicas, vídeos, documentos, outros): quantidade e espaço.
- **RF-05** — Extensões que mais ocupam.
- **RF-06** — Pastas que mais ocupam e pastas que mais cresceram no período; cada linha abre a pasta em Arquivos.
- **RF-07** — Maiores arquivos (os 10 primeiros), com "ver todos", que abre `/analytics/largest`: lista paginada de todos os arquivos, do maior para o menor, com nome, pasta, tamanho e tipo; cada linha abre o arquivo.
- **RF-08** — Duplicados: quantidade de grupos, de arquivos e espaço recuperável; "ver todos" abre a lista paginada de grupos.

### Duplicados (`/analytics/duplicates`)

- **RF-09** — Lista paginada de grupos, do maior desperdício para o menor, com filtro por tipo. Cada grupo expande mostrando as cópias (caminho, data, tier) com as ações: abrir, abrir pasta, excluir (para a lixeira). Selecionar várias cópias e excluir em lote; a tela impede excluir todas as cópias de um grupo sem confirmação extra.

### Biblioteca (`/analytics/library`)

- **RF-10** — Por tipo de mídia: total, com metadados, pendentes, com falha. Fila de processamento (por tipo de tarefa). Falhas recentes com o motivo e link para o arquivo. "Reprocessar falhas".
- **RF-11** — Saúde da indexação: última indexação completa, situação (em dia, indexando, com erro), erros nas últimas 24 h, "Reindexar".
- **RF-12** — Uso de IA: chamadas por provedor e por função no período.

## 7. Contrato HTTP

Um endpoint por informação (regra do projeto — não existe endpoint agregado).

| Método | Rota | Uso |
|---|---|---|
| GET | `/analytics/storage` | espaço, contagens e espaço por tier |
| GET | `/analytics/timeseries` | uso ao longo do tempo |
| GET | `/analytics/types` · `/extensions` | por tipo, por extensão |
| GET | `/analytics/top-folders` · `/hot-folders` | pastas |
| GET | `/analytics/recent-files` | recentes |
| GET | `/analytics/duplicates` · `/duplicates/groups` | duplicados |
| GET | `/analytics/library` · `/processing` · `/health` | biblioteca e saúde |
| GET | `/analytics/ai-usage` | uso de IA |
| GET | `/files/top-files-by-size` | maiores arquivos |

**A criar** (não existe hoje):

| Método | Rota | Uso |
|---|---|---|
| GET | `/analytics/trash` | espaço da lixeira (RF-02) → `{"item_count": 31, "size_bytes": 5368709120}` |
| GET | `/analytics/processing/failures` | falhas de processamento, paginadas, com filtro `stage` (RF-10) |
| POST | `/analytics/processing/retry-failed` | reprocessar as falhas (RF-10) → `202` com `{"job_id": 6291}`, ou `200` com `{"job_id": null}` |

- **`/analytics/storage`** ganha o campo opcional `tiers`: `{"hot_bytes": 107374182400, "cold_bytes": 536870912000}` (RF-02). Quente é o que não tem `physical_path` no tier frio; frio é o que tem. Os campos atuais não mudam — o contrato só cresce.

- **`/analytics/duplicates/groups`** e **`/files/top-files-by-size`** passam a aceitar `page` e `page_size` e a responder no envelope de [00 §10](00-padroes-transversais.md) quando `page` é enviado; sem `page`, respondem como hoje (compatibilidade). `/analytics/duplicates/groups` aceita ainda `type` (`image`, `audio`, `video`, `document`, `other`) e ordena por espaço recuperável decrescente, com desempate pelo checksum.
- **Item de `/analytics/processing/failures`**: `file_id`, `name`, `stage` (`checksum`, `metadata`, `thumbnail`, `document_text`, `ai`), `error` (traduzido), `failed_at`. Filtro `stage`.
- **`/analytics/processing/retry-failed`**: corpo opcional `{"stage": "thumbnail"}`; sem corpo, todas as etapas. Sem nenhuma falha → `200` com `{"job_id": null}`.
- **"Reindexar" (RF-11)** usa `POST /indexing/reindex` ([21 §8](21-indexacao.md)).
- **Links de RN-06**: maiores arquivos → `/analytics/largest` (tela desta spec, RF-07: lista paginada de `/files/top-files-by-size`); recentes → a tela Recentes de [02/RF-22](02-arquivos.md); pasta → a pasta em Arquivos ([02/RF-01](02-arquivos.md)); duplicados → `/analytics/duplicates` (RF-09); falhas → `/analytics/library` (RF-10), no bloco de falhas.

## 8. Situação em 2026-11-25

- Atendido: "Reindexar" chama `POST /indexing/reindex` com `scope: all` e "Atualizar" só recarrega os números (RN-07); os links levam à lista certa — maiores arquivos, duplicados, recentes e pastas (RN-06); as linhas de pasta existem e abrem a pasta em Arquivos; as telas `/analytics/largest` e `/analytics/duplicates` existem (RF-07, RF-09).
- Em aberto: o espaço total considera só a raiz principal (contraria RN-02).
- Em aberto: só a falha do bloco de armazenamento vira erro; as demais viram zeros e "saudável" (contraria RN-03).
- Em aberto: o uso de IA zera a cada reinício do servidor; existe um endpoint de sugestões por IA sem tela.
- Em aberto: espaço por tier (`tiers`), espaço da lixeira e falhas de processamento com reprocessamento ainda não têm endpoint (ver "A criar").

## 9. Critérios de aceite

- [ ] **CA-01** — O espaço usado, livre e total confere com o disco, somando todas as raízes.
- [ ] **CA-02** — A soma do espaço por tipo confere com o espaço usado pelos arquivos indexados.
- [ ] **CA-03** — Trocar o período altera os blocos de variação e o endereço; F5 mantém o período.
- [ ] **CA-04** — Com um endpoint falhando, só o bloco dele mostra erro; nenhum bloco mostra zero no lugar de erro.
- [ ] **CA-05** — Clicar em uma pasta da lista de maiores abre essa pasta em Arquivos.
- [ ] **CA-06** — "Ver todos" em maiores arquivos abre a lista ordenada por tamanho, do maior para o menor, com scroll infinito.
- [ ] **CA-07** — Copiar um arquivo grande para outra pasta o faz aparecer como grupo de duplicados; excluir uma cópia tira o grupo da lista e atualiza o espaço recuperável.
- [ ] **CA-08** — A lista de duplicados com mais de 100 grupos rola até o fim.
- [ ] **CA-09** — "Reindexar" inicia um job de indexação visível em Tarefas.
- [ ] **CA-10** — Durante uma importação grande, os pendentes da biblioteca diminuem até zerar.
- [ ] **CA-11** — Uma falha de processamento aparece com o motivo e o link abre o arquivo.

## 10. Decisões em aberto

- **D-01** — Sugestões por IA ("você pode liberar X GB removendo…"): o endpoint existe, sem tela. *Provisório: fora de escopo; o endpoint fica sem uso até decisão.*

## 11. Rastreabilidade

`docs/revisao-produto` item 12 (pendente); `03-imagens.md` lacuna 18; `08-lixeira.md` decisão 34.
