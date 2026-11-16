# 17 — Diário de atividades

**Prioridade:** P2 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Registrar o que o dono está fazendo ao longo do dia e quanto tempo dedicou a cada atividade, com o mínimo de cliques.

## 2. Cenários de uso

- Ao começar uma tarefa, digita o nome e adiciona; a anterior é encerrada sozinha.
- Volta a uma tarefa de mais cedo: "repetir" nela.
- No fim do dia, vê quanto tempo gastou em cada coisa.
- Errou o nome ou esqueceu de registrar a troca: corrige o nome e os horários.

## 3. Escopo

Registrar atividade, encerrar, repetir, editar, excluir, lista por dia, resumo do dia.

## 4. Fora de escopo

- Registro automático do que o servidor ou o dono fez nos arquivos (isso é log e [15](15-notificacoes-e-jobs.md)).
- Projetos, clientes, etiquetas, valor por hora, faturamento.
- Relatórios semanais/mensais e exportação. *(ver D-01)*
- Lembretes e metas.
- Várias atividades simultâneas.

## 5. Regras de negócio

- **RN-01** — Uma atividade tem nome (obrigatório, até 200 caracteres), descrição (opcional), início e fim.
- **RN-02** — Só uma atividade fica "em andamento". Registrar uma nova encerra a anterior no mesmo instante.
- **RN-03** — "Encerrar" termina a atividade em andamento sem iniciar outra.
- **RN-04** — "Repetir" cria uma atividade nova com o mesmo nome e descrição, começando agora (e encerra a que estiver em andamento).
- **RN-05** — Editar permite alterar nome, descrição, início e fim. O fim não pode ser anterior ao início; uma atividade não pode sobrepor outra — o conflito é apontado.
- **RN-06** — Excluir pede confirmação e é definitivo.
- **RN-07** — O **resumo do dia** considera o dia inteiro, da meia-noite até agora, no fuso do servidor: total de atividades, tempo total registrado, tempo por atividade (somando repetições do mesmo nome) e a atividade mais longa. A atividade em andamento conta até o momento atual.
- **RN-08** — A lista é ordenada do início mais recente para o mais antigo e agrupada por dia.
- **RN-09** — O diário é global: o mesmo em todos os dispositivos.
- **RN-10** — O contrato de edição é o mesmo para web e Android e aceita todos os campos de RN-05.

## 6. Requisitos por tela

### Diário (`/activity-diary`) — web

- **RF-01** — Alcançável pelo menu ([01/RF-01](01-shell-e-navegacao.md)).
- **RF-02** — No topo, a atividade em andamento com cronômetro e "Encerrar"; ao lado, o formulário de nova atividade (nome com sugestão dos nomes já usados, descrição, "Adicionar"; Enter adiciona).
- **RF-03** — Resumo do dia selecionado (RN-07), com seletor de data (hoje por padrão; dia anterior/seguinte).
- **RF-04** — Lista paginada, agrupada por dia: nome, descrição, início, fim ou "em andamento", duração. Ações por item: Repetir, Editar, Excluir.
- **RF-05** — Recarregar, estados e erros conforme [00](00-padroes-transversais.md); falha em uma página seguinte não esconde a lista nem o formulário.

### Android

- **RF-06** — As mesmas funções: atividade em andamento, adicionar, encerrar, repetir, editar (todos os campos), excluir, resumo do dia e lista paginada. Falhas de criar e editar são mostradas.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/diary/` | lista paginada |
| GET | `/diary/summary` | resumo |
| POST | `/diary/` | registrar (`name`, `description`) |
| POST | `/diary/copy` | repetir |
| PUT | `/diary/:id` | editar |

**A criar** (não existe hoje):

| Método | Rota | Uso |
|---|---|---|
| POST | `/diary/stop` | encerra a atividade em andamento (RN-03) |
| DELETE | `/diary/:id` | exclui (RN-06) → `204` |

- **`POST /diary/stop`**: sem corpo. `200` com a atividade encerrada; `409` quando não há atividade em andamento.
- **`PUT /diary/:id`** passa a aceitar JSON com `name`, `description`, `start_time` e `ended_at` (data ISO ou `null` para "em andamento"), além do formulário só com `name` que já aceita. `400` quando o fim é anterior ao início ou o nome passa de 200 caracteres; `409` quando o intervalo sobrepõe outra atividade, com a atividade em conflito na mensagem (RN-05).
- **Item de atividade** ganha `ended_at` (data ISO ou `null`). O campo `end_time`, que hoje sai como `{"Value", "HasValue"}`, continua igual por compatibilidade; clientes novos leem só `ended_at`.
- **`GET /diary/summary`** aceita `date=AAAA-MM-DD` (padrão: hoje, no fuso do servidor) e considera o dia inteiro, sem limite de quantidade (RN-07). A resposta ganha `activities`: `[{"name", "duration_seconds", "occurrences"}]`, um item por nome.
- **`GET /diary/`** aceita `date_from` e `date_to` (`AAAA-MM-DD`, inclusivos) além de `page` e `page_size`.

## 8. Situação em 2026-11-10

- A tela não está no menu; só é alcançada pela busca global ou digitando o endereço (`/internal/activity-diary`).
- **O "resumo do dia" considera apenas a última hora**, e para em 200 itens sem avisar (contraria RN-07).
- Não há edição nem exclusão na web. A edição existente aceita só o nome, em formato de formulário; no Android o campo de conteúdo aparece no diálogo e não é salvo; falhas de criar e editar no Android são ignoradas.
- O fim da atividade é serializado como `{Value, HasValue}`.
- Android lê só a primeira página e não mostra resumo nem repetir.

## 9. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) na lista:

- [ ] **CA-01** — Adicionar uma atividade a mostra em andamento com cronômetro; adicionar outra encerra a primeira no mesmo instante.
- [ ] **CA-02** — "Encerrar" termina a atividade em andamento e nenhuma outra começa.
- [ ] **CA-03** — "Repetir" em uma atividade antiga cria uma nova com o mesmo nome, em andamento.
- [ ] **CA-04** — Com atividades às 9h, 11h e 15h, o resumo do dia às 18h considera as três e o tempo total confere com a soma das durações.
- [ ] **CA-05** — Duas atividades com o mesmo nome no dia aparecem somadas no tempo por atividade.
- [ ] **CA-06** — Editar o início e o fim de uma atividade atualiza a duração e o resumo; fim antes do início é recusado.
- [ ] **CA-07** — Excluir uma atividade a remove da lista e do resumo.
- [ ] **CA-08** — Selecionar o dia anterior mostra o resumo e as atividades daquele dia.
- [ ] **CA-09** — Uma atividade criada no Android aparece na web e vice-versa.
- [ ] **CA-10** — No Android, editar nome e descrição salva os dois; uma falha é mostrada.

## 10. Decisões em aberto

- **D-01** — Resumo por semana/mês e exportação. *Provisório: fora de escopo.*
- **D-02** — Esta feature não atende a nenhuma das necessidades N1–N7. Manter no menu principal ou agrupar em "Ferramentas". *Provisório: "Ferramentas" ([01/RF-01](01-shell-e-navegacao.md)).*

## 11. Rastreabilidade

`docs/revisao-produto` item 14 (pendente); `02-shell-layout.md` lacuna 14.
