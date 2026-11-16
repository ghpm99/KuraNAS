# 12 — Capturas

**Prioridade:** P2 · **Plataformas:** plugin (captura), web (gestão) · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Guardar na biblioteca de vídeos o que o dono assiste no navegador — a mídia de uma página ou a gravação do que está sendo exibido — com nome e organização corretos, sem trabalho manual.

## 2. Cenários de uso

- O dono assiste a um episódio em tela cheia; a extensão grava e, ao fim, o episódio aparece em Vídeos, na série e temporada certas.
- Dorme no meio: a gravação para sozinha quando o episódio acaba; o próximo episódio vira outro arquivo.
- Em uma página com um vídeo sem proteção, clica em "Baixar" na extensão e o arquivo vai para o servidor.
- Uma captura deu errado: ele vê o motivo na tela de Capturas e tenta de novo.

## 3. Escopo

Extensão do navegador (detecção de mídia, download de mídia sem proteção, gravação da saída renderizada, envio em partes com retomada, descoberta do servidor), tela de Capturas (listar, acompanhar, abrir, repetir, excluir), promoção da captura para a biblioteca de vídeos.

## 4. Fora de escopo

- **Decriptação ou contorno de DRM, em qualquer forma, e qualquer ponto de extensão para isso** — decisão registrada. Conteúdo protegido só é guardado pela gravação do que já está na tela.
- Captura no nível do sistema operacional (fora do navegador).
- Edição da gravação (cortar abertura, juntar partes).
- Adaptadores de título/episódio para mais de um serviço além do já suportado (cada novo serviço é incremento próprio).
- Tradução da interface da extensão (decisão do projeto: o plugin ainda não segue a regra de i18n).
- Captura pelo Android.

## 5. Conceitos

- **Captura**: um arquivo de mídia enviado pela extensão, com seus metadados (título, temporada, episódio, plataforma, endereço de origem, pôster).
- **Sessão de envio**: envio em partes de uma captura, que pode ser retomado.
- **Chave do episódio**: identificador que a extensão atribui a um episódio, para que uma gravação interrompida e retomada continue no mesmo arquivo.
- **Promoção**: mover a captura concluída para a biblioteca de vídeos, com nome e pasta definidos pelos metadados.
- **Estados**: enviando → recebida → processando → na biblioteca | falhou.

## 6. Regras de negócio

### Envio

- **RN-01** — A captura é enviada em partes; a parte seguinte só é aceita na posição exata onde a anterior terminou. Queda de conexão retoma do ponto recebido.
- **RN-02** — Uma sessão aberta com a mesma chave de episódio é retomada, não duplicada. Episódio já capturado por completo não é regravado automaticamente; regravar é ação explícita.
- **RN-03** — Concluir um envio com zero bytes ou com tamanho diferente do declarado é recusado com mensagem.
- **RN-04** — Capturas em envio ficam em área própria, fora das raízes indexadas: nunca aparecem na biblioteca antes da promoção.

### Promoção

- **RN-05** — Captura concluída é promovida automaticamente, em segundo plano, para a biblioteca de vídeos ([21](21-indexacao.md)).
- **RN-06** — Destino: com episódio, `<título>/Temporada N/E<nn> - <título do episódio>`; sem episódio, `<título> (<ano>)`; sem metadados, o nome informado. Nome já existente recebe sufixo; nada é sobrescrito.
- **RN-07** — O arquivo promovido é ajustado para tocar no player do sistema (web e Android) antes de entrar na biblioteca.
- **RN-08** — O pôster, quando a extensão o informa, é guardado junto ao vídeo como imagem local.
- **RN-09** — Promoção que falha deixa a captura no estado "falhou", com o motivo, e o arquivo recebido é preservado para nova tentativa.
- **RN-10** — Depois de promovido, o vídeo é um item normal da biblioteca ([06](06-videos.md)).

### Gestão

- **RN-11** — Excluir uma captura ainda não promovida apaga o arquivo recebido. Excluir uma captura já promovida remove só o registro da captura; **o vídeo na biblioteca não é afetado** — para removê-lo, exclui-se o vídeo (que vai para a lixeira).
- **RN-12** — Início de envio, conclusão, falha de envio, promoção e falha de promoção geram notificação ([15](15-notificacoes-e-jobs.md)).

### Extensão

- **RN-13** — A gravação da tela só arma com o vídeo em tela cheia (quadro limpo, sem controles), para quando o episódio termina, e começa arquivo novo quando o próximo episódio inicia.
- **RN-14** — Em site sem adaptador, a captura inteligente não arma e a extensão oferece o modo manual, sem erro.
- **RN-15** — A extensão encontra o servidor sozinha na rede local; o endereço pode ser fixado manualmente. Um endereço só é aceito depois de confirmado como servidor KuraNAS.
- **RN-16** — Toda captura enviada leva os metadados que a extensão conseguir detectar (título, temporada, episódio, plataforma, endereço), seja gravação ou download.
- **RN-17** — Cada versão da extensão tem número de versão próprio, visível no popup.

## 7. Requisitos por tela

### Capturas (`/captures`)

- **RF-01** — Lista paginada, mais recente primeiro. Cada item: miniatura/pôster, título, temporada e episódio, plataforma, tamanho, data, estado, link para o endereço de origem.
- **RF-02** — Busca por nome e filtro por estado e plataforma.
- **RF-03** — Itens em envio ou em processamento mostram progresso e se atualizam sozinhos.
- **RF-04** — Ações por item: "Assistir" (abre o vídeo promovido), "Ver na pasta", "Tentar de novo" (quando falhou), "Excluir" (com confirmação que diz o que será apagado — RN-11).
- **RF-05** — Item que falhou mostra o motivo.
- **RF-06** — Estado vazio explica o que são capturas e aponta para a instalação da extensão ([23/RF-12](23-configuracoes-e-sistema.md)).

### Popup da extensão

- **RF-07** — Situação da conexão com o servidor e versão da extensão.
- **RF-08** — Título detectado; lista das mídias detectadas na aba, com "Baixar" (escolha de qualidade quando houver) ou "Capturar".
- **RF-09** — Controles de gravação: armar, desarmar, parar; indicador claro de "gravando".
- **RF-10** — "Salvar no servidor por URL" ([11/RF-07](11-download-por-url.md)).
- **RF-11** — Progresso do envio em andamento e resultado.
- **RF-12** — Endereço do servidor (automático ou manual).

## 8. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| POST | `/captures/upload/init` | abre ou retoma sessão (`name`, `size`, `episode_key`, `metadata`) |
| POST | `/captures/upload/chunk` | envia parte (`upload_id`, `offset`) |
| POST | `/captures/upload/complete` | conclui |
| POST | `/captures/upload` | envio em requisição única (compatibilidade) |
| GET | `/captures` | lista paginada (`name`, `media_type`) |
| GET, DELETE | `/captures/:id` | detalhe, excluir |

**A criar** (não existe hoje):

| Método | Rota | Uso |
|---|---|---|
| POST | `/captures/:id/promote` | repete a promoção de uma captura que falhou (RN-09, RF-04) |

- **`POST /captures/:id/promote`**: sem corpo. `202` com `{"job_id"}`; `409` quando a captura não está no estado "falhou"; `404` quando não existe.
- **`GET /captures`** passa a aceitar `status` (os valores que o campo `status` do item já usa) e `platform`, além de `name` e `media_type`. Valor inválido → `400`.
- **`DELETE /captures/:id`** (RN-11): em captura **não** promovida apaga o registro e o arquivo recebido; em captura promovida apaga **só o registro** — o arquivo da biblioteca (para o qual `file_path` aponta depois da promoção) e o item `file_id` não são tocados. A resposta `{"deleted": true}` ganha `library_file_kept` (`true` quando havia vídeo promovido).
- **Formato de `chunk` com posição errada**: passa a responder `409` com `{"error", "received_size"}` (hoje é `500`), como em [03 §7](03-transferencia-entre-dispositivos.md).

## 9. Estados e erros

| Situação | O que o usuário vê |
|---|---|
| Servidor não encontrado pela extensão | aviso no popup e campo para informar o endereço |
| Dispositivo fora da whitelist | aviso de acesso não liberado |
| Quadro preto na gravação (proteção do site) | a captura é concluída, mas o conteúdo é inútil — limitação conhecida, informada na documentação da extensão |
| Biblioteca de vídeos não configurada | promoção falha com motivo e atalho para Configurações |
| Sem espaço | falha com motivo |

## 10. Situação em 2026-11-10

- A lista carrega só as 50 primeiras capturas; sem busca, filtro, recarregar, "tentar de novo" ou confirmação ao excluir; estados não se atualizam sozinhos; não há como abrir o vídeo promovido.
- Excluir uma captura promovida **apaga o arquivo da biblioteca** (contraria RN-11).
- Download comum pela extensão não envia metadados nem chave de episódio (contraria RN-16).

## 11. Critérios de aceite

Além dos CA de [00](00-padroes-transversais.md) na lista:

- [ ] **CA-01** — Gravar um episódio em tela cheia até o fim: a gravação para sozinha e o episódio aparece em Vídeos na série e temporada certas, e toca na web e no Android.
- [ ] **CA-02** — Com reprodução automática do site, dois episódios seguidos geram dois arquivos.
- [ ] **CA-03** — Interromper a gravação de um episódio e retomar resulta em um único arquivo.
- [ ] **CA-04** — Derrubar a rede durante o envio e religar: o envio continua de onde parou.
- [ ] **CA-05** — Baixar pela extensão um vídeo sem proteção o coloca na biblioteca com o título detectado.
- [ ] **CA-06** — A tela de Capturas com mais de 50 itens rola até o fim.
- [ ] **CA-07** — Uma captura em processamento muda para "na biblioteca" na tela sem recarregar.
- [ ] **CA-08** — "Assistir" em uma captura promovida abre o vídeo no player.
- [ ] **CA-09** — Excluir uma captura promovida não remove o vídeo da biblioteca; a confirmação diz isso.
- [ ] **CA-10** — Uma promoção que falha mostra o motivo e "Tentar de novo" funciona depois de corrigida a causa.
- [ ] **CA-11** — Capturar duas vezes o mesmo título não sobrescreve o primeiro arquivo.
- [ ] **CA-12** — Com o servidor em outro IP, a extensão o encontra sozinha ou aceita o endereço manual só se for um servidor KuraNAS.
- [ ] **CA-13** — Não há no repositório código que decripte conteúdo protegido.

## 12. Decisões em aberto

- **D-01** — Validação da captura inteligente no navegador real, com o serviço usado pelo dono (pendente na task 19): se a gravação da aba entregar quadro preto, a captura precisa migrar para o nível do sistema operacional, o que é outra feature. *Provisório: gravação da aba.*

## 13. Rastreabilidade

`docs/melhorias/19-captura-inteligente-de-streaming.md` (bloqueada, aguardando validação do dono); issues #93 e #100; `docs/revisao-produto` item 10 (pendente).
