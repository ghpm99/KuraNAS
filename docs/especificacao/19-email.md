# 19 — E-mail

**Prioridade:** P2 · **Plataformas:** web (contas e mensagens), kiosk (painel) · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Mostrar, em um painel dentro de casa, os e-mails recentes das contas do dono já triados: os importantes resumidos, os suspeitos sinalizados — sem que o sistema possa enviar, apagar ou alterar qualquer coisa na caixa de entrada.

## 2. Cenários de uso

- O tablet na parede mostra os últimos e-mails; o dono vê de relance se chegou algo importante.
- Um e-mail de phishing chega: aparece marcado como suspeito, sem resumo.
- A conta Google pede novo login: uma notificação avisa e ele revincula.

## 3. Escopo

Vincular e desvincular contas (Google, Microsoft), sincronização periódica somente leitura, triagem determinística, classificação e resumo por IA, lista de mensagens, escolha do provedor de IA dessa análise.

## 4. Fora de escopo

Regras duras do projeto — não entram em nenhuma evolução:

- **Enviar, responder, encaminhar, apagar, mover ou marcar e-mail.** O acesso é somente leitura.
- **Baixar, guardar ou abrir anexos; seguir links do e-mail.**
- **Dar ferramentas ao modelo de IA** que analisa e-mail.

Também fora de escopo:

- Cliente de e-mail completo (ler o corpo inteiro, pastas, busca na caixa).
- Outros provedores (IMAP genérico, Yahoo).
- Consulta de reputação de remetente ou de links em serviços externos.
- Reanálise em massa ao trocar de provedor de IA.

## 5. Regras de negócio

### Contas

- **RN-01** — A feature só funciona com a chave de cifragem configurada no servidor; sem ela, a tela explica o que falta e nada é gravado.
- **RN-02** — Vincular pede ao provedor apenas permissão de leitura (mais a identificação do endereço). Nenhuma permissão de envio ou alteração é pedida, nunca.
- **RN-03** — Credenciais das contas são guardadas cifradas e não aparecem em nenhuma resposta, tela ou log.
- **RN-04** — Cada conta pode ter a sincronização pausada. Remover uma conta apaga suas credenciais e suas mensagens guardadas.
- **RN-05** — Conta cujo acesso expirou fica marcada como "precisa revincular" e gera notificação; as demais continuam sincronizando.

### Sincronização e análise

- **RN-06** — A sincronização roda periodicamente (padrão 10 min) e pode ser pedida na hora. Guarda de cada mensagem: remetente, assunto, data, um trecho curto e o resultado da análise — nunca o corpo completo depois de analisado, nunca anexos.
- **RN-07** — O conteúdo é reduzido a texto puro e limitado em tamanho antes de qualquer análise.
- **RN-08** — Spam evidente é barrado por regras determinísticas e não chega à IA.
- **RN-09** — A IA classifica cada mensagem (legítima, suspeita, maliciosa), indica a importância e resume as legítimas em 2–3 frases. Resposta da IA fora do formato esperado classifica a mensagem como suspeita.
- **RN-10** — Mensagem suspeita ou maliciosa nunca tem resumo exibido; gera notificação de atenção. Mensagem legítima importante gera notificação informativa.
- **RN-11** — Com a IA indisponível, as mensagens ficam pendentes e são analisadas no ciclo seguinte; nada falha.
- **RN-12** — O provedor de IA desta análise é escolhido à parte, local por padrão; escolher nuvem exibe aviso de privacidade. A troca vale na análise seguinte.
- **RN-13** — Mensagens guardadas são removidas após 30 dias.

## 6. Requisitos por tela

### Configurações → E-mail

- **RF-01** — Situação da feature (ativa / o que falta configurar).
- **RF-02** — Vincular conta Google e conta Microsoft, com o passo a passo de cada fluxo e o resultado.
- **RF-03** — Lista das contas: endereço, provedor, situação (ativa, pausada, precisa revincular), última sincronização. Ações: pausar/retomar, "Sincronizar agora", revincular, remover (com confirmação).
- **RF-04** — Provedor de IA da análise, com o aviso de privacidade.

### E-mails (`/email`)

- **RF-05** — Lista paginada, mais recente primeiro: remetente, assunto, data, conta, selo de classificação e de importância, resumo (só das legítimas).
- **RF-06** — Filtros: conta, classificação, importância, só pendentes.
- **RF-07** — Somente leitura: clicar expande o resumo e as evidências da classificação. Não há abrir corpo, responder nem link clicável do e-mail.
- **RF-08** — Suspeitas e maliciosas são visualmente distintas e mostram por que foram sinalizadas.

### Kiosk

- **RF-09** — Painel com os últimos e-mails ([25](25-app-legado-kiosk.md)).

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/email/accounts` | contas |
| DELETE | `/email/accounts/:id` | remover |
| PUT | `/email/accounts/:id/sync-enabled` | pausar/retomar |
| POST | `/email/accounts/:id/sync` | sincronizar agora |
| POST | `/email/accounts/google/auth-url` · GET `/email/oauth/google/callback` | vincular Google |
| POST | `/email/accounts/microsoft/device-code` · GET `…/status` | vincular Microsoft |
| GET | `/email/messages` | lista paginada |
| GET | `/email/messages/:id/summary` | resumo |
| GET, PUT | `/email/settings/provider` | provedor de IA |

**A criar** (não existe hoje): filtros na lista de mensagens.

## 8. Situação em 2026-11-10

- Não há tela de mensagens na web; o único lugar que mostra e-mails é o kiosk. "Sincronizar agora" e o resumo por mensagem existem na API sem tela.
- A vinculação das contas reais nunca foi validada pelo dono (task 14 bloqueada).

## 9. Critérios de aceite

- [ ] **CA-01** — Sem a chave de cifragem, a tela explica o que configurar e nenhuma rota grava dados.
- [ ] **CA-02** — Vincular uma conta Google e uma Microsoft funciona de ponta a ponta; a tela de consentimento pede só leitura.
- [ ] **CA-03** — Depois de vincular, os e-mails recentes aparecem na lista em até um ciclo de sincronização, ou na hora com "Sincronizar agora".
- [ ] **CA-04** — E-mails legítimos têm resumo; um e-mail de phishing de teste aparece como suspeito ou malicioso, sem resumo, e gera notificação.
- [ ] **CA-05** — Um e-mail com instruções para a IA ("ignore as instruções…") não altera o comportamento: é classificado, no pior caso, como suspeito.
- [ ] **CA-06** — Nenhum anexo é baixado e nenhum link do e-mail é aberto pelo servidor (verificável no log de rede).
- [ ] **CA-07** — Pausar uma conta interrompe a sincronização só dela.
- [ ] **CA-08** — Remover uma conta apaga suas mensagens e credenciais.
- [ ] **CA-09** — Com a IA desligada, as mensagens aparecem como pendentes e são analisadas quando a IA volta.
- [ ] **CA-10** — As credenciais não aparecem em nenhuma resposta da API nem nos logs.
- [ ] **CA-11** — A lista de e-mails segue os CA de [00](00-padroes-transversais.md).

## 10. Decisões em aberto

- **D-01** — Tela de e-mails na web (RF-05 a RF-08) ou só o painel do kiosk, que era o consumo previsto. *Provisório: em escopo, com prioridade abaixo de todas as P0 e P1.*
- **D-02** — Esta feature não atende a nenhuma das necessidades N1–N7. *Provisório: manter como está; nenhuma evolução antes das P0.*

## 11. Rastreabilidade

`docs/melhorias/` tasks 14 (bloqueada, aguardando o dono registrar os apps OAuth e vincular as contas), 15 e 16 (concluídas); decisões "Regras duras de e-mail" em `docs/melhorias/README.md`; `docs/revisao-produto` item 17 (pendente).
