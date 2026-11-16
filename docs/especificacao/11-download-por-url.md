# 11 — Download por URL

**Prioridade:** P2 · **Plataformas:** web, plugin, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Guardar no servidor um vídeo ou áudio de uma página da internet colando o endereço, para assistir ou ouvir depois pela biblioteca, sem depender do site.

## 2. Cenários de uso

- O dono cola o endereço de um vídeo, escolhe "Vídeos/Shows" e "1080p"; minutos depois o arquivo está na biblioteca de vídeos.
- Cola o endereço de uma música, escolhe "só áudio (MP3)" e a pasta de músicas.
- Na própria página, usa o botão do plugin para mandar o endereço ao servidor.

## 3. Escopo

Pedir um download (URL, destino, formato), acompanhar o andamento, cancelar, repetir, histórico, notificação de conclusão e falha, manutenção da ferramenta de download.

## 4. Fora de escopo

- **Conteúdo protegido por DRM.** O sistema não decripta nem contorna proteção, e não tem ponto de extensão para isso (decisão registrada).
- Download de arquivo genérico por HTTP, torrent, magnet.
- Agendamento, assinatura de canais, download automático de novidades.
- Login em sites, cookies de conta.
- Atualização automática da ferramenta de download (é manual, com verificação de integridade).

## 5. Regras de negócio

- **RN-01** — O endereço deve ser `http` ou `https`, com no máximo 2 048 caracteres. Fora disso: `400` com mensagem própria.
- **RN-02** — O destino é uma raiz habilitada mais uma subpasta opcional, que não pode sair da raiz. Destino inválido e subpasta inválida têm mensagens distintas.
- **RN-03** — O formato vem de uma lista fechada: só áudio (MP3), vídeo até 1080p, vídeo na melhor qualidade. Valor fora da lista: `400`.
- **RN-04** — Pedir um download cria um job em segundo plano e responde na hora com o identificador dele ([15](15-notificacoes-e-jobs.md)).
- **RN-05** — Um endereço que aponta para uma playlist baixa só o item do endereço. *(ver D-02)*
- **RN-06** — O arquivo só aparece no destino quando está completo; arquivo parcial nunca é visto pela biblioteca.
- **RN-07** — O download nunca sobrescreve: se o nome já existe no destino, recebe sufixo ` (2)`, ` (3)`…
- **RN-08** — Cancelar interrompe o download em andamento e descarta o que foi baixado.
- **RN-09** — Um download que dura mais que o limite (`INGEST_FETCH_TIMEOUT_MINUTES`, padrão 360 minutos) é interrompido, falha com motivo próprio e libera a vez.
- **RN-10** — Conclusão e falha geram notificação. A mensagem de falha é traduzida e não contém caminho interno do servidor.
- **RN-11** — O arquivo baixado entra na biblioteca pela indexação normal ([21](21-indexacao.md)), no tipo correspondente.
- **RN-12** — Sem a ferramenta de download instalada, a tela informa e aponta para Configurações; pedidos são recusados com mensagem clara.
- **RN-13** — Só uma atualização da ferramenta roda por vez; o binário novo é conferido por checksum antes de substituir o atual, que fica como reserva.

## 6. Requisitos por tela

### Baixar de URL (web)

- **RF-01** — Formulário: endereço (com colar), destino (raiz e subpasta, com seletor de pasta), formato. O último destino e formato usados vêm preenchidos.
- **RF-02** — Validação no campo antes de enviar; erro do servidor aparece no formulário.
- **RF-03** — Ao enviar, o pedido entra na lista "Em andamento" e o formulário fica pronto para o próximo.
- **RF-04** — **Em andamento**: título (quando conhecido) ou endereço, destino, formato, barra de progresso com porcentagem, velocidade e tempo restante, e "Cancelar". Atualiza sozinha.
- **RF-05** — **Histórico**: lista paginada, mais recente primeiro, com estado (concluído, falhou, cancelado), data e destino. Concluído tem "Abrir" (leva ao arquivo); falhou mostra o motivo e "Tentar de novo".

### Configurações → ferramenta de download

- **RF-06** — Mostra se está instalada, a versão e se há versão nova; "Atualizar" com progresso e resultado.

### Plugin

- **RF-07** — Botão no popup envia o endereço da aba atual com destino e formato escolhidos; o popup mostra o andamento do pedido enviado.

### Android

- **RF-08** — "Baixar de URL" com os mesmos campos; aceita endereço compartilhado de outro aplicativo ("Compartilhar → KuraNAS"); acompanha pelo andamento de jobs.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| POST | `/ingest/fetch` | pedir download (`url`, destino, formato) → `job_id` |
| GET | `/ingest/targets` | destinos possíveis |
| GET | `/ingest/presets` | formatos |
| GET | `/ingest/ytdlp/status` | situação da ferramenta |
| POST | `/ingest/ytdlp/update` | atualizar a ferramenta |
| GET | `/jobs` · `/jobs/:id` · POST `/jobs/:id/cancel` | andamento, histórico, cancelar |

`/downloads` pertence à distribuição dos aplicativos cliente ([23](23-configuracoes-e-sistema.md)), não a esta feature.

## 8. Estados e erros

| Situação | O que o usuário vê |
|---|---|
| Endereço inválido ou longo demais | mensagem no campo |
| Site sem suporte / conteúdo indisponível / protegido | falha no histórico com motivo traduzido |
| Sem espaço no destino | falha com motivo |
| Ferramenta ausente | aviso na tela com atalho para Configurações |
| Servidor reiniciado durante o download | o item aparece como falho, com "Tentar de novo" |

## 9. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) no histórico:

- [ ] **CA-01** — Colar um endereço válido, escolher destino e formato e enviar: o item aparece em andamento com progresso e, ao concluir, o arquivo está na pasta escolhida e na biblioteca.
- [ ] **CA-02** — "Só áudio" gera um MP3 que aparece em Música.
- [ ] **CA-03** — Cancelar interrompe o download; nenhum arquivo aparece no destino.
- [ ] **CA-04** — Baixar duas vezes o mesmo endereço para o mesmo destino resulta em dois arquivos, o segundo com ` (2)`.
- [ ] **CA-05** — Um endereço inválido é recusado no formulário, com mensagem.
- [ ] **CA-06** — Um endereço de site sem suporte termina em falha com motivo traduzido, sem caminho do servidor, e gera notificação.
- [ ] **CA-07** — A conclusão gera notificação que leva ao arquivo.
- [ ] **CA-08** — "Tentar de novo" em um item falho cria um novo pedido com os mesmos dados.
- [ ] **CA-09** — Durante o download, a pasta de destino não mostra arquivo parcial.
- [ ] **CA-10** — Sem a ferramenta instalada, a tela avisa e o envio é recusado com mensagem.
- [ ] **CA-11** — Atualizar a ferramenta mostra a versão nova; duas atualizações simultâneas não acontecem.
- [ ] **CA-12** — Enviar pelo plugin cria o pedido e o popup mostra o andamento.

## 10. Decisões em aberto

Correspondem às lacunas 24–28 do board.

- **D-01** — Bloquear endereços de hosts internos/loopback. *Provisório: aceitar, dado que só IPs da whitelist alcançam o servidor.*
- **D-02** — Baixar playlist inteira. *Provisório: só o item do endereço (RN-05).*
- **D-03** — Formatos extras (720p, outros formatos de áudio, legendas). *Provisório: os três de RN-03.*
- **D-04** — Limite de tamanho/duração e checagem de espaço livre antes de baixar. *Provisório: sem limite; falta de espaço vira falha com motivo.*
- **D-05** — Servidor reiniciado com download em andamento. *Provisório: marcar como falho e permitir repetir.*

## 11. Rastreabilidade

`docs/revisao-produto/09-ingestao-por-url.md` (em execução: lacunas 01–09 resolvidas, 10–23 pendentes, 24–28 aguardam decisão); decisão "Ingestão de mídia" em `docs/melhorias/README.md`.
