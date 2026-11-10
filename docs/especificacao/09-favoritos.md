# 09 — Favoritos

**Prioridade:** P1 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Marcar com uma estrela o que o dono usa sempre — uma pasta, um documento, uma foto, uma música, um vídeo — e chegar a esses itens em um lugar só.

## 2. Cenários de uso

- O dono favorita a pasta "Documentos/Casa" e os PDFs que consulta todo mês; abre Favoritos e está tudo lá.
- Favorita músicas enquanto ouve; elas formam a playlist "Favoritas".
- Favorita as melhores fotos de uma viagem e filtra a galeria por favoritas.

## 3. Escopo

Marcar e desmarcar (um item e em lote), tela de Favoritos, filtro de favoritos nas telas de mídia, estrela em toda representação do item.

## 4. Fora de escopo

- Favoritos por dispositivo ou por pessoa (o sistema é de usuário único; a estrela é global).
- Coleções ou pastas de favoritos, etiquetas, notas.
- Nível de avaliação (1 a 5 estrelas).

## 5. Regras de negócio

- **RN-01** — Qualquer item — arquivo de qualquer tipo ou pasta — pode ser favoritado. A estrela é uma só por item e vale em todos os dispositivos.
- **RN-02** — Favoritar uma pasta não favorita o conteúdo dela.
- **RN-03** — Marcar e desmarcar são operações idempotentes: marcar o que já está marcado não muda nada e não é erro.
- **RN-04** — A estrela acompanha o item ao renomear, mover (inclusive entre raízes), enviar para a lixeira e restaurar. Reindexação e tarefas de segundo plano nunca apagam uma estrela.
- **RN-05** — Item na lixeira não aparece em Favoritos e não pode ser favoritado (`404`).
- **RN-06** — A estrela é o mesmo dado em todo lugar: favoritar uma música no player a mostra favoritada em Arquivos, na busca e em Favoritos.
- **RN-07** — A data em que o item foi favoritado é guardada. Ordenação padrão de Favoritos: favoritado mais recentemente primeiro.
- **RN-08** — Em lote: se todos os selecionados já são favoritos, a ação desmarca todos; caso contrário, marca todos.

## 6. Requisitos por tela

### Estrela (em todo lugar)

- **RF-01** — A estrela aparece e pode ser alternada em: cards e linhas de Arquivos, visualizador e painel de detalhes, galeria e visualizador de imagens, listas e player de música, cards e player de vídeo, resultados da busca global, Home.
- **RF-02** — Alternar é imediato na tela; se o servidor recusar, volta ao estado anterior com a mensagem de erro.

### Favoritos (`/favorites`)

- **RF-03** — Lista paginada de todos os favoritos, com as mesmas visões (grade/lista) e a mesma apresentação de item de Arquivos.
- **RF-04** — Filtro por tipo: Todos, Pastas, Documentos, Imagens, Músicas, Vídeos, Outros — cada um com a contagem real, vinda do servidor.
- **RF-05** — Ordenação: data de favoritado (padrão, decrescente), nome, tamanho, data de modificação.
- **RF-06** — Abrir uma pasta favorita navega por ela dentro de Favoritos (`/favorites/<caminho>`), com breadcrumb que começa em "Favoritos"; abrir um arquivo usa o visualizador de Arquivos.
- **RF-07** — Seleção, ações em lote, menu de contexto e atalhos iguais aos de Arquivos, mais "Remover dos favoritos".
- **RF-08** — Remover a estrela de um item o tira da lista na hora.

### Android

- **RF-09** — Tela de Favoritos paginada, com filtro por tipo e navegação por pastas; estrela no explorer, nas imagens, nas músicas e nos vídeos.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/files/starred` | lista paginada (tipo, ordenação) |
| GET | `/files/starred/counts` | contagem por tipo |
| PUT | `/files/starred/:id` | define o estado (`{"starred": bool}`), idempotente |
| POST | `/files/starred` | lote (`{"ids": [], "starred": bool}`) |
| POST | `/files/starred/:id` | alterna (mantido por compatibilidade) |

## 8. Estados e erros

- Sem favoritos: estado vazio explicando como favoritar.
- Favoritar item inexistente ou na lixeira: `404`, com mensagem.
- Lote com falha parcial: informa quantos foram marcados e quais falharam.

## 9. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) na tela de Favoritos e em cada filtro:

- [ ] **CA-01** — Favoritar um arquivo em Arquivos o faz aparecer em Favoritos e na Home sem recarregar.
- [ ] **CA-02** — Favoritar uma música no player mostra a estrela na lista de faixas, em Arquivos e na playlist "Favoritas".
- [ ] **CA-03** — Um item favoritado em uma subpasta profunda aparece em Favoritos.
- [ ] **CA-04** — Favoritar, mover para outra pasta, renomear, excluir e restaurar: a estrela continua.
- [ ] **CA-05** — Reindexar a biblioteca não remove nenhuma estrela.
- [ ] **CA-06** — As contagens dos filtros conferem com a quantidade de itens de cada filtro.
- [ ] **CA-07** — Abrir uma pasta favorita mostra o conteúdo dela dentro de Favoritos; o breadcrumb volta para a lista de favoritos.
- [ ] **CA-08** — Selecionar 30 itens e favoritar marca os 30 com uma única ação; repetir a ação desmarca os 30.
- [ ] **CA-09** — Com o servidor recusando a operação, a estrela volta ao estado anterior e o erro aparece.
- [ ] **CA-10** — Favoritar em um dispositivo aparece no outro após recarregar.
- [ ] **CA-11** — No Android: favoritar uma foto na galeria a mostra na tela de Favoritos e na web.

## 10. Decisões em aberto

- **D-01** — Arquivo movido por fora do aplicativo (direto no disco) perde a estrela, porque é reindexado como arquivo novo. Aceitar, ou herdar a estrela por checksum ao reindexar. *Provisório: aceitar; mover pelo aplicativo preserva.*

## 11. Rastreabilidade

`docs/revisao-produto/07-favoritos.md` (31 lacunas; 24–26 aguardam decisão — 24 e 25 estão em [06/D-05](06-videos.md), 26 é a D-01 acima).
