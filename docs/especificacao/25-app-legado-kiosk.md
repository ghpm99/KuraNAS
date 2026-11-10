# 25 — App legado (kiosk)

**Prioridade:** P2 · **Plataformas:** Android antigo (`mobile/`, a partir do Android 4.1)

## 1. Objetivo

Dar uso a um tablet antigo fixado na parede: um painel sempre ligado com hora, data, as últimas notificações do servidor e os últimos e-mails, sem precisar de toque.

## 2. Cenários de uso

- O tablet fica na parede da sala; o dono olha e vê a hora e se chegou algo importante.
- O servidor é desligado à noite: o painel indica "offline" e volta sozinho de manhã.

## 3. Escopo

Conexão com o servidor (descoberta e endereço manual), tela única de painel (relógio, notificações, e-mails), modo quiosque, comportamento offline.

## 4. Fora de escopo

- Qualquer navegação de mídia ou de arquivos (removida deste app por decisão registrada; isso é do app `android/`).
- Interação com os itens: marcar como lido, abrir, rolar.
- Modo retrato; outras resoluções além da do tablet alvo.
- Atualização em tempo real por conexão permanente (o painel consulta periodicamente).
- Mudanças no servidor motivadas por este app.

## 5. Regras de negócio

- **RN-01** — O app encontra o servidor sozinho (última conexão lembrada, anúncio na rede, consulta por broadcast, varredura da rede) e aceita endereço manual. Um endereço só é aceito depois de confirmado como servidor KuraNAS.
- **RN-02** — O relógio atualiza a cada segundo e nunca atrasa visivelmente; data por extenso no idioma do app.
- **RN-03** — Notificações são consultadas a cada 60 s e e-mails a cada 120 s, em instantes defasados, com respostas pequenas (até 8 itens, menos de 4 KB por consulta).
- **RN-04** — Servidor fora do ar: cada painel mostra "offline" e mantém o último conteúdo; as tentativas se espaçam progressivamente (até 5 min) e a retomada é automática. Nunca trava nem fecha.
- **RN-05** — E-mails suspeitos ou maliciosos são visualmente distintos e não mostram resumo. Nada é renderizado como HTML.
- **RN-06** — Com a feature de e-mail desligada no servidor, o painel de e-mails mostra "não configurado", não "offline".
- **RN-07** — A tela fica sempre ligada, em tela cheia, sem barras do sistema; se as barras reaparecem, o modo quiosque é reaplicado.
- **RN-08** — Painel sem itens mostra texto de vazio. Todo texto vem do catálogo de traduções, com reserva local quando o servidor não responde.
- **RN-09** — O app é leve o bastante para rodar de forma contínua e fluida no tablet alvo (Galaxy Tab 2).

## 6. Requisitos por tela

- **RF-01** — **Conexão**: situação da busca, campo de endereço manual, "Conectar", "Tentar de novo".
- **RF-02** — **Painel** (paisagem): relógio grande e data; painel de notificações (tipo, título, mensagem, há quanto tempo); painel de e-mails (remetente, assunto, resumo ou selo de alerta, há quanto tempo); indicador de offline por painel.
- **RF-03** — Um gesto discreto (toque longo em um canto) abre a troca de servidor.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET | `/health` | confirmação do servidor |
| GET | `/notifications?page=1&page_size=8` | painel de notificações |
| GET | `/email/messages?page=1&page_size=8` | painel de e-mails |
| GET | `/configuration/translation` | textos |

Toda mudança nesses formatos confere este consumidor.

## 8. Situação em 2026-11-10

- Código pronto e testes verdes; faltam as validações no ambiente do dono: instalar e rodar no tablet, e medir o tamanho das respostas no servidor real (tasks 17 e 18).
- E-mail desligado no servidor aparece como "offline" (contraria RN-06).

## 9. Critérios de aceite

- [ ] **CA-01** — No tablet alvo, o app instala, encontra o servidor e mostra o painel.
- [ ] **CA-02** — O relógio fica correto após 24 h ligado.
- [ ] **CA-03** — Uma notificação nova no servidor aparece no painel em até 60 s; um e-mail novo analisado, em até 120 s.
- [ ] **CA-04** — Cada consulta do painel pesa menos de 4 KB (medido no servidor real).
- [ ] **CA-05** — Desligar o servidor mostra "offline" nos dois painéis sem fechar o app; religar retoma sozinho.
- [ ] **CA-06** — Um e-mail suspeito aparece destacado e sem resumo.
- [ ] **CA-07** — Com o e-mail desligado no servidor, o painel de e-mails diz "não configurado".
- [ ] **CA-08** — A tela não apaga e as barras do sistema não ficam visíveis.
- [ ] **CA-09** — Após 7 dias ligado continuamente, o app continua fluido.

## 10. Decisões em aberto

Nenhuma.

## 11. Rastreabilidade

`docs/melhorias/` tasks 17 e 18 (bloqueadas, aguardando validação no tablet do dono).
