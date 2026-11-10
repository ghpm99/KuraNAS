# 15 — Notificações e jobs

**Prioridade:** P1 · **Plataformas:** web, Android, kiosk · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Avisar o dono do que aconteceu no servidor sem ele precisar procurar, e mostrar o que o servidor está fazendo em segundo plano — com progresso, resultado e a possibilidade de cancelar.

## 2. Cenários de uso

- Um download por URL terminou: o sino mostra o aviso e o clique leva ao arquivo.
- O backup falhou de madrugada: de manhã há uma notificação de erro com o motivo.
- O dono adicionou uma raiz grande e quer saber quanto falta para terminar de indexar.
- Um job travou: ele cancela.

## 3. Escopo

Notificações dentro do aplicativo (sino, tela, tipos, lidas/não lidas, destino do clique, retenção), tarefas em segundo plano (lista, detalhe, progresso, cancelar, repetir).

## 4. Fora de escopo

- Notificação push do sistema operacional, e-mail, SMS, mensageiros. *(ver D-01)*
- Preferências por tipo de notificação (silenciar categorias). *(ver D-02)*
- Agendar jobs pela interface; alterar prioridade.
- Log técnico completo na interface (fica em arquivo de log no servidor).

## 5. Conceitos

- **Notificação**: aviso com tipo (informação, sucesso, atenção, erro), título, mensagem, data, lida/não lida e, quando faz sentido, um destino.
- **Job**: tarefa em segundo plano, com tipo, estado e passos. Estados: na fila, executando, concluído, concluído com falhas, falhou, cancelado.

## 6. Regras de negócio

### Notificações

- **RN-01** — Geram notificação: conclusão e falha de download por URL; envio, promoção e falha de captura; início, conclusão e falha de importação Takeout; conclusão com falhas e falha de backup; falha de migração de tier; falha de indexação; importação por pasta monitorada; aviso de desligamento automático; atualização disponível (sistema e ferramenta de download); conta de e-mail que precisa ser revinculada; e-mail suspeito ou importante; falha da limpeza automática da lixeira.
- **RN-02** — Operação que o dono fez e viu terminar na tela (renomear, favoritar, enviar um arquivo com o painel aberto) não gera notificação.
- **RN-03** — Toda notificação tem título e mensagem traduzidos, compreensíveis sem contexto, sem erro interno nem caminho do servidor.
- **RN-04** — Notificação com destino leva, ao clique, exatamente ao item ou à tela que resolve: o arquivo baixado, o vídeo capturado, a tela de backup, a conta de e-mail. Não existe notificação clicável que não leva a lugar nenhum.
- **RN-05** — Avisos repetidos do mesmo tipo em sequência (vários arquivos importados) são agrupados em um só, com contagem, preservando o texto do aviso. Erros nunca são agrupados.
- **RN-06** — Clicar em uma notificação a marca como lida. "Marcar todas como lidas" só fica ativo quando há não lidas.
- **RN-07** — Notificações são removidas automaticamente após 30 dias. O dono pode excluir uma notificação ou limpar as lidas.
- **RN-08** — O contador do sino reflete o servidor em até 30 s e imediatamente após uma ação do próprio dispositivo.
- **RN-09** — Ordenação: mais recente primeiro.

### Jobs

- **RN-10** — Toda operação demorada é um job visível: indexação, extração de metadados, geração de miniaturas, marcação por IA, backup, migração de tier, download por URL, promoção de captura, importação Takeout, sincronização e análise de e-mail, download de modelo de IA, indexação de texto de documentos.
- **RN-11** — Cada job mostra nome traduzido do tipo, estado, progresso (porcentagem e passos concluídos/total), início, fim, duração e, quando falha, o motivo traduzido.
- **RN-12** — Jobs na fila ou executando podem ser cancelados. Cancelar interrompe de fato o trabalho e não deixa resultado parcial. Cancelar um job já terminado não é erro.
- **RN-13** — Jobs que falharam e são repetíveis oferecem "Tentar de novo".
- **RN-14** — Ordenação padrão: em andamento primeiro, depois os demais do mais recente para o mais antigo.
- **RN-15** — Jobs terminados são mantidos por 30 dias e depois removidos.
- **RN-16** — Um job interrompido por reinício do servidor volta à fila ou é marcado como falho com motivo — nunca fica "executando" para sempre.

## 7. Requisitos por tela

### Sino (cabeçalho)

- **RF-01** — Ícone com contador de não lidas (até "99+"). Clique abre painel com as 5 mais recentes, "Marcar todas como lidas" e "Ver todas".
- **RF-02** — Abaixo das notificações, as tarefas em andamento com barra de progresso e link para a tela de tarefas.

### Notificações (`/notifications`)

- **RF-03** — Lista paginada com scroll infinito. Cada item: ícone do tipo, título, mensagem, data relativa, indicador de não lida.
- **RF-04** — Filtros: Todas, Não lidas, e por tipo. Recarregar; "Marcar todas como lidas"; "Limpar lidas".
- **RF-05** — Clique marca como lida e leva ao destino (RN-04). Menu do item: marcar como lida/não lida, excluir.
- **RF-06** — Marcar como lida na aba "Não lidas" tira o item da lista sem repetir nem pular outros.

### Tarefas (`/jobs`)

- **RF-07** — Lista paginada com filtros por estado e por tipo; em andamento no topo, atualizando sozinhas.
- **RF-08** — Item: tipo, descrição do alvo (qual arquivo, qual raiz, qual endereço), estado, barra de progresso, tempo decorrido; ações Cancelar / Tentar de novo.
- **RF-09** — Detalhe: passos com estado de cada um, horários, motivo da falha, e o destino do resultado quando houver.
- **RF-10** — Telas que iniciam um job (Download por URL, Takeout, Armazenamento, Configurações de IA) mostram o progresso do próprio job e um link para o detalhe.

### Android

- **RF-11** — Tela de notificações paginada, com filtros, marcar como lida (com volta ao estado anterior em caso de falha), marcar todas e contador no ícone do app. Tela de tarefas alcançável pelo menu, com progresso, cancelar e detalhe.

### Kiosk

- **RF-12** — O painel do app legado mostra as últimas notificações, somente leitura ([25](25-app-legado-kiosk.md)).

## 8. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/notifications` | lista paginada (`type`, `is_read`) |
| GET | `/notifications/unread-count` | contador |
| GET | `/notifications/:id` | detalhe |
| PUT | `/notifications/:id/read` · `/notifications/read-all` | marcar lida(s) |
| GET | `/jobs` | lista paginada (`status`, `type`, `priority`) |
| GET | `/jobs/:id` · `/jobs/:id/steps` | detalhe, passos |
| POST | `/jobs/:id/cancel` | cancelar |

**A criar** (não existe hoje): excluir notificação e limpar lidas; marcar como não lida; destino explícito na notificação (tipo e identificador do alvo); ordenação de jobs por mais recente; repetir job.

## 9. Situação em 2026-11-10

- **Não existe tela de tarefas na web**; no Android a tela existe mas não é alcançável, lê só a primeira página e não reconhece o estado "na fila" (não atualiza nem cancela job enfileirado).
- `GET /jobs` ordena do mais antigo para o mais novo, sem desempate.
- Só parte das notificações tem destino no clique; o destino é deduzido de prefixos de texto.
- O agrupamento substitui a mensagem original por "N título".
- "Marcar todas como lidas" está sempre ativo; não há exclusão.
- A tela de notificações troca a lista pela tela de erro quando uma página seguinte falha ([00 anexo A4](00-padroes-transversais.md)).
- No Android, notificações: só a primeira página, sem filtro e sem contador; falha ao marcar como lida é ignorada.

## 10. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) nas duas listas:

- [ ] **CA-01** — Um download por URL concluído gera notificação; o clique abre o arquivo.
- [ ] **CA-02** — Um backup que falha gera notificação de erro com motivo compreensível; o clique abre a tela de backup.
- [ ] **CA-03** — Dez arquivos importados em sequência por pasta monitorada geram um aviso agrupado com a contagem, não dez.
- [ ] **CA-04** — O contador do sino aumenta em até 30 s após um evento e zera após "Marcar todas como lidas".
- [ ] **CA-05** — Na aba "Não lidas" com três páginas, marcar itens como lidos não gera duplicados nem esconde itens não lidos.
- [ ] **CA-06** — Nenhuma notificação mostra texto em inglês cru, erro interno ou caminho do servidor.
- [ ] **CA-07** — Uma notificação com mais de 30 dias não aparece mais.
- [ ] **CA-08** — Adicionar uma raiz mostra o job de indexação na tela de tarefas, com progresso que avança.
- [ ] **CA-09** — Cancelar um download por URL em andamento o interrompe e o job fica "cancelado".
- [ ] **CA-10** — A lista de tarefas mostra as em andamento no topo e as terminadas da mais recente para a mais antiga.
- [ ] **CA-11** — Reiniciar o servidor com um job em execução: ao voltar, o job não fica preso em "executando".
- [ ] **CA-12** — O detalhe de um job que falhou mostra em qual passo e por quê.
- [ ] **CA-13** — No Android, as notificações rolam além da primeira página e a tela de tarefas é alcançável e cancela um job na fila.

## 11. Decisões em aberto

- **D-01** — Notificação push no Android (aviso mesmo com o app fechado). Sem serviço externo, exige o app manter verificação periódica em segundo plano. *Provisório: fora de escopo; os avisos aparecem ao abrir o app.*
- **D-02** — Silenciar tipos de notificação. *Provisório: fora de escopo.*

## 12. Rastreabilidade

`docs/revisao-produto` item 13 (pendente); `02-shell-layout.md` lacunas 12 e 13.
