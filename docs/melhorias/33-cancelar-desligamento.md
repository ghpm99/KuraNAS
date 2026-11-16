# 33 — Desligamento automático: aviso com contagem e "Cancelar desta vez"

**Tipo:** bug · **Prioridade:** P1 · **Spec:** [23](../especificacao/23-configuracoes-e-sistema.md) `RN-10`, `RN-11`, `RF-06`, `RF-07`, §7, `CA-08`

## Contexto

Na hora programada, o agendador notifica e chama o comando de desligar do sistema operacional **já com o atraso embutido** (`backend/internal/api/v1/autoshutdown/scheduler.go:74-103`; `shutdown_exec.go:35-47`: `shutdown /s /t N` no Windows, `shutdown -h +minutos` no Linux). Depois disso o servidor não tem como voltar atrás: não existe rota de cancelar (só `GET`/`PUT /auto-shutdown/settings` e `GET /auto-shutdown/suggested-time`, `backend/internal/app/routes.go:166-175`) e o código nunca emite `shutdown /a` nem `shutdown -c`. Desligar o interruptor na tela só evita os próximos dias. Quem está assistindo a um filme vê uma notificação e o servidor cai.

## Objetivo

No horário, todos os clientes mostram uma contagem regressiva com "Cancelar desta vez", e cancelar impede o desligamento daquele dia.

## O que fazer

1. **Contagem no servidor**: o agendador entra em estado `warning` com `shutdown_at = agora + período de aviso` e só chama o comando do sistema operacional (sem atraso) quando a contagem termina sem cancelamento.
2. **`GET /auto-shutdown/status`** (contrato de 23 §7).
3. **`POST /auto-shutdown/cancel`**: volta a `idle`, registra o cancelamento do dia; `409` fora do período de aviso. A execução do dia seguinte continua programada.
4. **Período de aviso configurável** em `/auto-shutdown/settings` (`grace_seconds`, 30 a 600, padrão 60), se o campo ainda não for exposto.
5. **Web — faixa em todas as telas** (`RF-07`): o shell consulta o status a cada 30 s (a cada 5 s em `warning`) e mostra a faixa com a contagem e "Cancelar desta vez"; ao cancelar, a faixa some em todos os dispositivos na consulta seguinte.
6. **Web — seção de Configurações** (`RF-06`): mostra a próxima execução e o texto dizendo que o computador será desligado.

## Critérios de aceite

- [ ] Teste do agendador com relógio falso: entra em `warning`, chama o desligamento só ao fim da contagem; com cancelamento, não chama.
- [ ] Teste: cancelar fora do aviso responde `409`; o dia seguinte dispara normalmente depois de um cancelamento.
- [ ] Teste de handler decodificando `status`; seam test do `POST /auto-shutdown/cancel`; render sem mock da faixa.
- [ ] `23/CA-08` — no sistema rodando: no horário aparece o aviso com contagem; "Cancelar desta vez" impede o desligamento; sem cancelar, o computador desliga.
- [ ] A faixa aparece em um segundo dispositivo aberto em outra tela.
- [ ] `make ci` verde.

## Fora de escopo

- Adiar por atividade em andamento (`23/RN-12`, `CA-09`): sub-item próprio depois desta task, porque exige cada feature informar atividade.
- Faixa de aviso no Android e no kiosk.
- "Sugerir horário".
