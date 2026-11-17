# 00 — Padrões transversais

**Prioridade:** P0 · **Plataformas:** web, Android

## 1. Objetivo

Definir uma única vez o comportamento que se repete em todas as features — listas paginadas, estados de tela, mutações, erros, seleção, textos — para que toda tela se comporte igual e para que uma mudança em uma tela não quebre outra.

Toda spec de feature herda este documento. Uma spec só repete uma regra daqui quando precisa de um valor diferente (por exemplo, outra ordenação padrão).

## 2. Escopo

Listas e paginação, estados de carregamento/vazio/erro, recarregar, ordenação, filtros e busca em lista, abrir item, seleção e ações em lote, mutações, mensagens de erro, textos, datas e tamanhos, comportamento responsivo de listas, contrato de paginação da API.

## 3. Fora de escopo

- Conteúdo específico de cada lista (colunas, cards, ações próprias).
- Moldura da aplicação ([01](01-shell-e-navegacao.md)).

## 4. Listas

Vale para **toda** lista de itens vindos do servidor: arquivos, fotos, faixas, álbuns, vídeos, playlists, favoritos, lixeira, notificações, jobs, capturas, resultados de busca, conversas, mensagens.

### 4.1 Paginação e scroll infinito

- **RN-01** — Toda lista é paginada no servidor. Nenhuma tela pede "tudo" e nenhuma tela filtra, ordena ou conta no cliente o que só foi carregado em parte.
- **RN-02** — A primeira página é pedida ao abrir a tela. As seguintes são pedidas automaticamente quando o usuário rola até perto do fim (scroll infinito). Não há botão "carregar mais" como único meio; ele existe apenas como reserva quando o carregamento automático falha (RN-09).
- **RN-03** — **A página seguinte é anexada ao fim da lista já exibida.** Os itens anteriores permanecem na tela, na mesma ordem e na mesma posição de rolagem. Carregar a página N nunca substitui, reordena ou remove itens das páginas 1…N-1.
- **RN-04** — Enquanto uma página seguinte carrega, aparece um indicador no fim da lista; os itens já exibidos continuam visíveis e utilizáveis.
- **RN-05** — Só há um pedido de próxima página em andamento por lista. Rolar rápido não dispara pedidos duplicados nem pula páginas.
- **RN-06** — A lista termina quando o servidor diz que não há próxima página. No fim aparece o total ("N itens") em vez do indicador de carregamento. Página seguinte vazia com indicação de que não há mais é fim de lista, não erro.
- **RN-07** — Um item nunca aparece duas vezes. Se uma página traz um item já exibido (porque algo foi inserido ou removido no servidor entre os pedidos), ele é ignorado pela identidade (`id`).
- **RN-08** — Tamanho de página padrão: 50 itens em listas de linha, 60 em grades de miniaturas. Tela que alterna entre grade e lista (Arquivos, Favoritos, Recentes, busca na pasta) usa 60 nos dois modos, para que trocar o modo não recarregue a lista. Todo cliente envia `page` e `page_size` explicitamente — nenhum depende do padrão do servidor. O servidor impõe teto de 500 e limita ao teto o pedido que passar dele.

### 4.2 Falha ao carregar

- **RN-09** — Falha ao carregar uma **página seguinte** não altera a lista: os itens já exibidos ficam na tela e, no fim da lista, aparece a mensagem de erro com "Tentar de novo", que repete apenas a página que falhou. **Uma falha nunca esvazia a lista.**
- **RN-10** — Falha ao carregar a **primeira página** mostra o estado de erro da tela (§5), nunca o estado vazio.
- **RN-11** — Resposta fora do formato esperado (campo ausente, tipo errado) é tratada como falha (RN-09/RN-10), não como lista vazia.

### 4.3 Recarregar

- **RN-12** — Toda lista tem a ação "Recarregar" visível no cabeçalho da tela. Ela busca de novo a partir da primeira página e volta a rolagem ao topo.
- **RN-13** — Durante o recarregamento, os itens atuais continuam na tela até a resposta chegar; se a resposta falhar, os itens atuais são mantidos e o erro é avisado.
- **RN-14** — Ao voltar o foco para a aba depois de mais de um minuto fora, a lista revalida em segundo plano sem mover a rolagem nem piscar.

### 4.4 Ordenação

- **RN-15** — A ordenação é feita no servidor e vale para a lista inteira, não para o que está carregado.
- **RN-16** — Ordenação padrão quando a spec da feature não diz outra: **data, mais recente primeiro**. Pastas vêm antes de arquivos nas listas que misturam os dois.
- **RN-17** — Toda ordenação tem desempate determinístico por `id`, para que a paginação não repita nem pule itens.
- **RN-18** — Trocar a ordenação recomeça a lista da primeira página e volta ao topo. A escolha fica na URL e é lembrada por tela no dispositivo.

### 4.5 Filtro e busca dentro da lista

- **RN-19** — Filtro e busca são aplicados no servidor. O total mostrado é o total do filtro, não o do que foi carregado.
- **RN-20** — Mudar filtro ou texto de busca recomeça a lista da primeira página. A digitação tem espera de 300 ms e o pedido anterior é cancelado; resposta de um pedido antigo nunca substitui a de um mais novo.
- **RN-21** — Filtro e busca ficam na URL: F5 e "voltar" reproduzem a mesma lista.
- **RN-22** — Filtro sem resultado mostra o estado vazio de filtro ("nada encontrado para …", com a ação "Limpar filtros"), diferente do estado vazio da lista.

### 4.6 Abrir um item

- **RN-23** — Clicar (ou tocar) em um item abre a visualização detalhada dele: o arquivo no visualizador, a pasta na listagem dela, o álbum na tela do álbum. Onde a tela tem seleção, um clique seleciona e duplo clique (ou Enter) abre; no toque, um toque abre e toque longo seleciona.
- **RN-24** — Todo item é um link real: abre em nova aba com Ctrl+clique e tem endereço próprio.
- **RN-25** — Ao voltar da visualização detalhada, a lista está como foi deixada: mesmos itens carregados, mesma rolagem, mesma ordenação e filtro.

### 4.7 Seleção e ações em lote

- **RN-26** — Onde há seleção: clique seleciona um; Ctrl+clique alterna; Shift+clique seleciona o intervalo; Ctrl+A seleciona todos os carregados; Esc limpa.
- **RN-27** — Com itens selecionados aparece a barra de ações em lote com a contagem e as ações da feature. Uma ação em lote informa o resultado por item: quantos deram certo, quais falharam e por quê.

### 4.8 Desempenho

- **RN-28** — A primeira página aparece em até 1 s na rede local para bibliotecas de até 100 mil itens. Listas longas são virtualizadas: com 5 mil itens carregados, só os itens próximos da área visível existem na página (no máximo 300 elementos de item montados), e rolar não trava a interface por mais de 100 ms.
- **RN-29** — Miniaturas são carregadas sob demanda, só as próximas da área visível, com tamanho reservado para a lista não pular enquanto carregam.

## 5. Estados de tela

Toda tela que depende do servidor tem exatamente estes estados, mutuamente exclusivos:

| Estado | Quando | O que mostra |
|---|---|---|
| Carregando | primeira carga, sem dado anterior | esqueleto no formato do conteúdo |
| Com dados | há pelo menos um item | o conteúdo |
| Vazio | o servidor respondeu com sucesso e zero itens | ilustração, frase do que é a tela e a ação principal para começar |
| Vazio de filtro | sucesso, zero itens, com filtro/busca ativos | "nada encontrado" + "Limpar filtros" |
| Erro | a primeira carga falhou | mensagem do servidor, "Tentar de novo", detalhes técnicos recolhidos |

- **RN-30** — **Vazio só aparece depois de uma resposta bem-sucedida com zero itens.** Carregando, erro e resposta inválida nunca são mostrados como vazio.
- **RN-31** — Dado já exibido não é trocado por esqueleto durante revalidação.
- **RN-32** — Cada parte independente da tela tem seu próprio estado: a falha de um bloco não derruba os outros.

## 6. Mutações

Criar, renomear, mover, copiar, excluir, restaurar, favoritar, enviar, adicionar a playlist, salvar configuração.

- **RN-33** — Toda mutação dá retorno visível: sucesso (aviso curto) ou erro (mensagem do servidor). Nenhuma falha é silenciosa.
- **RN-34** — Durante a mutação, só o controle e o item envolvidos ficam desabilitados; o resto da tela continua utilizável.
- **RN-35** — Depois de uma mutação bem-sucedida, toda tela que mostra o dado afetado reflete a mudança sem recarregar a página — inclusive telas de outras features (excluir um arquivo atualiza Arquivos, Imagens, Favoritos, Lixeira, Home).
- **RN-36** — Atualização otimista só em mutações reversíveis e baratas (favoritar, marcar como lida, reordenar). Se o servidor recusar, a tela volta ao estado anterior e mostra o erro.
- **RN-37** — Ação destrutiva pede confirmação em diálogo da aplicação, dizendo o que será afetado e se é reversível. Excluir é reversível (lixeira) por padrão.
- **RN-38** — Clicar duas vezes não executa a mutação duas vezes.
- **RN-39** — Operação demorada (mais de 2 s) vira job em segundo plano com progresso ([15](15-notificacoes-e-jobs.md)); a tela não fica bloqueada esperando.

## 7. Erros

- **RN-40** — O servidor responde erros como `{"error": "<mensagem traduzida>"}` com status HTTP coerente: `400` entrada inválida, `403` dispositivo fora da whitelist, `404` não existe, `409` conflito, `413` grande demais, `507` sem espaço no destino, `500` falha interna. Nenhuma resposta de erro traz erro cru da linguagem, caminho interno do servidor ou stack trace.
- **RN-41** — O cliente mostra a mensagem do servidor como chegou. Quando não há resposta (rede), mostra a mensagem própria de falha de conexão.
- **RN-42** — Todo erro do servidor é registrado em arquivo de log com método, rota, status, o erro original e o stack trace.

## 8. Textos, datas e números

- **RN-43** — Nenhum texto visível é literal no código; tudo vem do catálogo de traduções, nos idiomas pt-BR e en-US, com as mesmas chaves nos dois.
- **RN-44** — Datas e horas no idioma do aplicativo; data relativa ("há 3 horas") até 7 dias, absoluta depois. O servidor envia datas em ISO 8601 com fuso.
- **RN-45** — Tamanhos em unidades binárias legíveis (KB, MB, GB) com uma casa decimal; durações em `h:mm:ss` ou `m:ss`.
- **RN-46** — Nome longo é truncado com reticências e o nome completo aparece em tooltip; nada estoura o layout.

## 9. Responsividade de listas

- **RN-47** — Grades adaptam o número de colunas à largura; listas em tabela viram cards ou escondem colunas secundárias em tela estreita. Nunca há rolagem horizontal da lista.
- **RN-48** — Toda ação disponível com mouse existe no toque (menu de contexto → toque longo ou botão "⋯").

## 10. Contrato de paginação da API

Envelope único para toda listagem paginada:

```json
{
  "items": [],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "has_next": true,
    "has_prev": false
  }
}
```

- **RN-49** — `page` começa em 1. `has_next` é a única fonte de verdade para "existe próxima página"; o cliente não deduz isso do número de itens recebidos.
- **RN-50** — `items` é sempre um array; lista vazia é `[]`, nunca `null` nem campo ausente.
- **RN-51** — O total de itens é pedido em endpoint ou campo próprio quando a tela precisa mostrá-lo; a listagem não paga uma contagem a cada página.
- **RN-52** — Parâmetros: `page`, `page_size`, `sort`, `order` (`asc`|`desc`), mais os filtros da feature. Valor inválido responde `400` com mensagem traduzida — o servidor não ignora parâmetro inválido nem cai em um padrão silencioso.
- **RN-53** — Enquanto houver endpoints com envelope diferente (contrato só cresce), o cliente tem **um** adaptador por formato, testado contra a resposta real do servidor; nenhuma tela lê campos de paginação por conta própria.

> Os envelopes realmente devolvidos hoje e suas divergências estão no anexo A.

## 11. Critérios de aceite

Aplicam-se a cada tela com lista. A verificação é feita no sistema rodando, com uma biblioteca que tenha pelo menos três páginas de itens.

- [ ] **CA-01** — Abrir a tela mostra a primeira página; rolar até o fim carrega a segunda e a terceira sem clique.
- [ ] **CA-02** — Depois de carregar a segunda página, os itens da primeira continuam na tela, na mesma ordem; a quantidade exibida é a soma das páginas.
- [ ] **CA-03** — Rolar muito rápido até o fim de uma lista de várias páginas não gera item duplicado nem página pulada.
- [ ] **CA-04** — Derrubar o servidor entre a primeira e a segunda página: os itens da primeira ficam na tela e aparece erro com "Tentar de novo" no fim da lista; religar e tentar de novo continua do ponto certo.
- [ ] **CA-05** — Com o servidor desligado desde o início, a tela mostra o estado de erro, não o estado vazio.
- [ ] **CA-06** — Em uma pasta/coleção realmente vazia, a tela mostra o estado vazio com a ação principal.
- [ ] **CA-07** — "Recarregar" volta ao topo e busca do início; um item criado por fora aparece.
- [ ] **CA-08** — Trocar a ordenação reordena a lista inteira (o primeiro item passa a ser o primeiro do servidor, não o primeiro do que estava carregado).
- [ ] **CA-09** — Digitar na busca da lista filtra pelo servidor: encontra item que ainda não tinha sido carregado.
- [ ] **CA-10** — F5 com ordenação, filtro e busca ativos reabre a mesma lista.
- [ ] **CA-11** — Rolar até a terceira página, abrir um item e voltar: a lista está na mesma posição, com as três páginas.
- [ ] **CA-12** — No fim da lista aparece o total e nenhum indicador de carregamento permanece.
- [ ] **CA-13** — Excluir um item em uma tela faz ele sumir das outras telas que o mostram, sem recarregar a página.
- [ ] **CA-14** — Uma mutação recusada pelo servidor mostra a mensagem do servidor e a tela volta ao estado anterior.
- [ ] **CA-15** — Em 360 px de largura a lista não tem rolagem horizontal e todas as ações são alcançáveis.
- [ ] **CA-16** — Existe teste automatizado, por tela com lista, que alimenta o cliente com a **resposta real do handler** (não um objeto escrito à mão no teste) e verifica CA-02 e CA-04.

## 12. Anexo A — situação em 2026-11-10

Levantamento por leitura de código (não reproduzido em execução). Cada linha é uma violação das regras acima e explica o defeito relatado pelo dono.

**Envelope real do servidor** (`backend/pkg/utils/dto.go`): `{"items": [], "pagination": {"page", "page_size", "has_next", "has_prev"}}` — igual ao de §10. Exceção: a galeria de imagens devolve um envelope plano com cursor (`{items, next_cursor, has_next, page, page_size}`).

| # | Onde | O que acontece | Regra violada |
|---|---|---|---|
| A1 | Arquivos, Favoritos, Recentes (`features/files/providers/fileProvider`) | O cliente lê `pagination.hasNext`; o servidor envia `pagination.has_next`. O campo é sempre indefinido, então a segunda página **nunca é pedida**: cada pasta mostra no máximo 200 itens. Os testes passam porque o mock foi escrito com `hasNext`. | RN-02, RN-49, RN-53, CA-16 |
| A2 | Busca na pasta (`features/files/search/useFileSearchResults`) | Mesma divergência: só os 100 primeiros resultados. | idem |
| A3 | Seletor de pasta de mover/copiar (`components/folderPicker`) | Mesma divergência; além disso, erro na primeira página esvazia a lista sem estado de erro, e o filtro de pastas é feito no cliente. | RN-01, RN-10, RN-49 |
| A4 | Arquivos, Favoritos, Recentes, busca na pasta, busca de documentos, Vídeos (filmes e pastas), Notificações, Diário | A tela testa "houve erro" antes de desenhar os itens. Quando uma página seguinte falha — ou quando a revalidação automática ao voltar o foco falha em qualquer página — **a lista inteira é trocada pela tela de erro**. É a causa mais provável de "na segunda requisição fica sem itens". | RN-09, RN-13, RN-14, RN-31 |
| A5 | Capturas | Carrega só a primeira página (50) e ignora a paginação. | RN-02 |
| A6 | Lixeira | Paginação por "anterior/próxima" com 15 itens, não scroll infinito. | RN-02, RN-03 |
| A7 | Assistente (conversas), Analytics (listas) | Sem paginação. | RN-01 |
| A8 | Jobs (`GET /jobs`) | Ordena do mais antigo para o mais novo e sem desempate por `id`. | RN-16, RN-17 |
| A9 | Notificações, aba "Não lidas" | Marcar como lida refaz todas as páginas sobre uma lista que encolheu: pode repetir ou pular itens. | RN-07 |
| A10 | Android — Arquivos | Pede a pasta sem `page`/`page_size`; o padrão do servidor é 15, então cada pasta mostra **15 itens** e não carrega mais. | RN-02, RN-08 |
| A11 | Android — Notificações, Diário, Jobs | Só a primeira página. | RN-02 |
| A12 | Home (web) | Erro e vazio mostram a mesma mensagem, sem "Tentar de novo". | RN-30 |

Telas que já cumprem RN-09 (mantêm a lista quando a página seguinte falha): Imagens e Música. Servem de referência para corrigir as demais.

## 13. Rastreabilidade

Regras extraídas das lacunas recorrentes em `docs/revisao-produto/` (01.01, 02.13, 03.03, 05.09, 06.33, 08.16, entre outras) e do defeito relatado pelo dono em 2026-11-10: telas com scroll em que a segunda requisição deixa a lista sem itens.
