# 14 — Home

**Prioridade:** P1 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Ser a tela inicial que leva o dono, em um clique, de volta ao que ele estava fazendo e ao que acabou de chegar ao servidor.

## 2. Cenários de uso

- Abre o sistema à noite: "Continuar assistindo" mostra o episódio de ontem; um clique e está assistindo.
- Acabou de enviar fotos do celular: elas estão no bloco "Fotos recentes".
- Quer saber se está tudo bem com o servidor: o bloco de situação mostra espaço livre e se há problema.

## 3. Escopo

Blocos de retomada (música, vídeo), conteúdo recente (arquivos, fotos), favoritos, atalhos para as seções, situação do sistema.

## 4. Fora de escopo

- Personalização dos blocos (ordem, esconder). *(ver D-01)*
- Recomendações e sugestões automáticas.
- Clima, agenda, notícias ou qualquer conteúdo de fora do servidor.
- Feed de atividade detalhado (isso é de [15](15-notificacoes-e-jobs.md)).

## 5. Regras de negócio

- **RN-01** — Cada bloco é independente: carrega, fica vazio e falha por conta própria. A falha de um bloco não afeta os outros nem impede a tela de abrir.
- **RN-02** — Bloco com erro mostra erro com "Tentar de novo"; bloco sem conteúdo mostra o estado vazio com a ação para começar. Os dois são visualmente diferentes ([00/RN-30](00-padroes-transversais.md)).
- **RN-03** — A Home reflete mutações feitas em outras telas sem recarregar (favoritar, excluir, enviar, assistir).
- **RN-04** — Os blocos mostram poucos itens (uma fileira) e têm "Ver todos", que leva à lista completa na tela da feature.
- **RN-05** — "Arquivos recentes" são os adicionados ou modificados mais recentemente; "abertos recentemente" são os que o dono abriu. Cada bloco diz qual dos dois mostra.
- **RN-06** — A Home abre rápido: o que demora (contagens, situação do sistema) não segura os blocos de retomada.

## 6. Requisitos por tela

### Home (`/home`) — web

Blocos, nesta ordem:

- **RF-01** — **Continuar assistindo**: até 4 vídeos em andamento com miniatura, título (série e episódio), barra de progresso; clique retoma ([06/RN-14](06-videos.md)).
- **RF-02** — **Continuar ouvindo**: a faixa e a fila do dispositivo, com tocar; se nada está na fila, as tocadas recentemente ([05/RF-01](05-musica.md)).
- **RF-03** — **Fotos recentes**: as 6–12 fotos adicionadas mais recentemente; clique abre no visualizador da galeria.
- **RF-04** — **Arquivos recentes**: os 6 adicionados ou modificados mais recentemente, com tipo, pasta e data; clique abre o arquivo.
- **RF-05** — **Favoritos**: os 6 favoritados mais recentemente.
- **RF-06** — **Situação do sistema**: espaço usado e livre (somando todas as raízes), total de arquivos, situação da indexação (em dia, indexando, com erro), situação do backup, tarefas em andamento. Itens com problema ficam destacados e levam à tela que resolve.
- **RF-07** — **Atalhos**: Enviar arquivos, Arquivos, Imagens, Música, Vídeos, Documentos.
- **RF-08** — Cabeçalho com Recarregar (recarrega todos os blocos).

### Home — Android

- **RF-09** — Os mesmos blocos de retomada (vídeo e música), fotos recentes, arquivos recentes, favoritos, atalhos e situação resumida do servidor. Cada bloco carrega e falha de forma independente.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/video/continue` | continuar assistindo |
| GET | `/music/player-state` · `/music/library/recent-plays` | continuar ouvindo |
| GET | `/files/image/library` (ordenado por adição) | fotos recentes |
| GET | `/analytics/recent-files` | arquivos recentes |
| GET | `/files/starred` | favoritos |
| GET | `/analytics/storage` · `/analytics/health` · `/backup/status` · `/jobs` | situação do sistema |

## 8. Situação em 2026-11-10

- Na web, erro e vazio mostram a mesma mensagem, sem "Tentar de novo" (contraria RN-02).
- No Android, a Home faz cinco chamadas como uma só: a falha de qualquer uma derruba a tela inteira (contraria RN-01); não mostra favoritos, continuar assistindo nem música; e chama de "acessados recentemente" uma lista ordenada por criação (contraria RN-05).
- O espaço total considera só a raiz principal.

## 9. Critérios de aceite

- [ ] **CA-01** — Com vídeo em andamento, música na fila, fotos e favoritos, todos os blocos aparecem preenchidos.
- [ ] **CA-02** — Em um servidor recém-instalado e vazio, cada bloco mostra seu estado vazio com a ação para começar; nenhum mostra erro.
- [ ] **CA-03** — Com um dos endpoints falhando, só o bloco dele mostra erro com "Tentar de novo"; os demais funcionam.
- [ ] **CA-04** — Clicar em um item de "Continuar assistindo" abre o player no ponto em que parou.
- [ ] **CA-05** — Enviar uma foto e voltar à Home: ela está em "Fotos recentes" sem recarregar a página.
- [ ] **CA-06** — Favoritar um arquivo em Arquivos o mostra no bloco de Favoritos.
- [ ] **CA-07** — "Ver todos" de cada bloco abre a lista completa correspondente.
- [ ] **CA-08** — O espaço mostrado é a soma de todas as raízes e confere com o disco.
- [ ] **CA-09** — Com a indexação ou o backup com erro, o bloco de situação destaca o problema e o clique leva à tela correspondente.
- [ ] **CA-10** — No Android, com um endpoint falhando, os demais blocos aparecem.

## 10. Decisões em aberto

- **D-01** — Blocos configuráveis. *Provisório: ordem fixa (§6).*

## 11. Rastreabilidade

`docs/revisao-produto` item 18 (pendente).
