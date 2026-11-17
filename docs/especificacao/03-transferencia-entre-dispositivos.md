# 03 — Transferência entre dispositivos

**Prioridade:** P0 · **Plataformas:** web, Android · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Levar arquivos de qualquer dispositivo da casa para o servidor e do servidor para qualquer dispositivo, de forma tão simples quanto "tirei a foto, está no PC".

## 2. Cenários de uso

- **Celular → servidor → PC.** O dono tira fotos no celular; elas chegam ao servidor (sozinhas ou com um toque em "Compartilhar") e ele as vê na galeria do PC.
- **PC → servidor → celular.** Arrasta um PDF para o navegador no PC; no celular, abre o app, encontra o arquivo e baixa para a pasta Downloads do aparelho.
- **PC → PC.** Sobe uma pasta inteira de um computador e baixa como zip no outro.
- Envia um vídeo de 4 GB pelo Wi-Fi; a conexão cai no meio e o envio continua de onde parou.

## 3. Escopo

Envio pela web (arquivos e pastas), envio pelo Android (manual, por "Compartilhar" e automático da câmera), download na web e no Android (arquivo, vários, pasta), fila de transferências com progresso, conflitos, retomada.

## 4. Fora de escopo

- Sincronização bidirecional de pastas (estilo Dropbox/Syncthing).
- Envio a partir do app legado ([25](25-app-legado-kiosk.md)).
- Transferência direta entre dois dispositivos sem passar pelo servidor.
- Acesso de fora da rede local.
- Backup automático de outros dados do celular (contatos, mensagens, aplicativos).
- Cliente de sincronização para Windows/Linux (o PC usa o navegador ou WebDAV — [22](22-acesso-pela-rede.md)).

## 5. Regras de negócio

### Envio (comum a web e Android)

- **RN-01** — O destino de todo envio é uma pasta dentro de uma raiz habilitada. Na web, por padrão, a pasta que está aberta.
- **RN-02** — Um arquivo enviado só aparece no destino quando está completo e íntegro. Envio interrompido não deixa arquivo parcial visível.
- **RN-03** — Arquivo enviado aparece na listagem da pasta e na biblioteca do seu tipo (galeria, música, vídeos, documentos) sem ação manual, em até 1 minuto, e com miniatura e metadados em até 5 minutos quando a fila de processamento está vazia.
- **RN-04** — Envio nunca sobrescreve em silêncio. Quando o nome já existe no destino, vale a política escolhida: manter os dois (sufixo ` (2)`), substituir, ou pular. Padrão: perguntar na web; manter os dois nos envios automáticos.
- **RN-05** — Arquivo idêntico (mesmo conteúdo, conferido por checksum) já presente no destino é pulado e contado como "já existia".
- **RN-06** — Não há limite de tamanho imposto pelo sistema; o limite é o espaço livre. Arquivos grandes são enviados em partes e o envio **retoma** de onde parou após queda de conexão ou reinício do aplicativo.
- **RN-07** — Todo envio declara o tamanho do arquivo antes do primeiro byte; quando o espaço livre do destino é menor, o envio é recusado ali, com mensagem.
- **RN-08** — Em um lote, a falha de um arquivo não cancela os outros. Ao fim, o resumo diz quantos foram, quantos já existiam, quantos falharam e por quê, com "Tentar de novo" para os que falharam.
- **RN-09** — Envio de pasta preserva a estrutura de subpastas.
- **RN-10** — As datas originais do arquivo (modificação e, em fotos, a data em que foi tirada) são preservadas.
- **RN-11** — Nomes de arquivo inválidos para o servidor são ajustados de forma previsível (caracteres proibidos viram `_`) e o ajuste é informado.

### Envio automático da câmera (Android)

- **RN-12** — Desligado por padrão. Ligado, envia fotos e vídeos novos da câmera para uma pasta escolhida, organizados em subpastas `AAAA/MM` pela data em que foram tirados.
- **RN-13** — Só roda quando o aparelho está na rede do servidor. Opções: só no Wi-Fi (padrão, ligado), só carregando (padrão, desligado), incluir vídeos (padrão, ligado).
- **RN-14** — Cada foto é enviada uma única vez, mesmo que o app seja reinstalado ou o servidor reindexado (RN-05). Apagar a foto no celular não a apaga no servidor, e vice-versa.
- **RN-15** — Ao ligar, o dono escolhe entre enviar tudo o que já está no aparelho ou só o que for tirado a partir de agora.
- **RN-16** — Funciona com o app fechado. Fora de casa, fica pendente sem gastar bateria com tentativas contínuas e retoma ao voltar.
- **RN-17** — O dono vê a situação: última sincronização, quantos pendentes, falhas.

### Download

- **RN-18** — Baixar um arquivo entrega o original, byte a byte, com o nome original (inclusive acentos e espaços).
- **RN-19** — O download é transmitido em fluxo, sem depender do tamanho do arquivo, e aceita retomada (o navegador ou o app continua de onde parou).
- **RN-20** — Baixar vários itens ou uma pasta entrega um zip com a estrutura de pastas, gerado em fluxo. O nome do zip é o da pasta ou `kuranas-AAAAMMDD-HHMM.zip`.
- **RN-21** — Arquivo no tier frio baixa normalmente.
- **RN-22** — Arquivo ausente do disco responde `404` com mensagem, nunca um arquivo vazio.
- **RN-23** — No Android, o arquivo baixado vai para a pasta Downloads do aparelho (fotos e vídeos também aparecem na galeria do aparelho) e uma notificação permite abri-lo.

## 6. Requisitos por tela

### Enviar (web)

- **RF-01** — Três formas: botão "Enviar" (arquivos ou pasta), arrastar e soltar sobre a listagem (a área de soltar fica destacada) e colar da área de transferência (imagem).
- **RF-02** — Disponível em Arquivos (para a pasta aberta) e na galeria de Imagens, em Música e em Vídeos (para a pasta padrão do tipo, com opção de escolher outra).
- **RF-03** — Painel de transferências, fixo no canto, que não bloqueia a navegação: um item por arquivo com nome, tamanho, barra de progresso, velocidade, tempo restante e estado (na fila, enviando, concluído, já existia, falhou, cancelado).
- **RF-04** — Ações: cancelar um arquivo, cancelar tudo, tentar de novo os que falharam, limpar concluídos. Clicar em um concluído leva ao arquivo.
- **RF-05** — Progresso geral do lote (arquivos e bytes). Até 3 envios simultâneos.
- **RF-06** — Diálogo de conflito com as opções de RN-04 e "aplicar a todos".
- **RF-07** — Sair da página ou fechar a aba com envio em andamento pede confirmação.
- **RF-08** — Ao concluir, a listagem mostra os arquivos novos sem recarregar.

### Baixar (web)

- **RF-09** — "Baixar" no menu do item, no visualizador, no painel de detalhes e na barra de lote. O download é gerenciado pelo navegador.
- **RF-10** — Antes de gerar um zip com mais de 1 GB ou mais de 1 000 arquivos, a tela informa quantidade e tamanho total e pede confirmação.

### Enviar (Android)

- **RF-11** — Em uma pasta do explorer: "Enviar" abre o seletor do sistema (fotos, vídeos, arquivos), com seleção múltipla.
- **RF-12** — **Compartilhar para o KuraNAS**: em qualquer aplicativo (galeria, câmera, arquivos), "Compartilhar → KuraNAS" envia um ou vários itens. Abre a escolha de destino, com o último destino usado como padrão e a opção "sempre enviar para cá".
- **RF-13** — Os envios rodam em segundo plano, com notificação de progresso e de resultado; continuam com o app fechado e retomam após perda de rede.
- **RF-14** — Tela "Transferências": em andamento, concluídas e com falha, com tentar de novo e cancelar.

### Envio automático (Android → Configurações)

- **RF-15** — Ligar/desligar; pasta de destino; opções de RN-13; escolha de RN-15; situação (RN-17); "Sincronizar agora".
- **RF-16** — Pede as permissões necessárias explicando o motivo; sem permissão, a tela diz o que falta.

### Baixar (Android)

- **RF-17** — "Baixar" no menu do item, no visualizador e em seleção múltipla; progresso em notificação; ao concluir, "Abrir".
- **RF-18** — "Abrir com…" e "Compartilhar" a partir de um arquivo do servidor (baixa para cache e entrega a outro aplicativo).

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| POST | `/files/upload` | envio de um arquivo (multipart) |
| GET | `/files/download/:id` | download com retomada |
| GET | `/files/download-zip` | vários itens ou pasta |
| GET | `/files/blob/:id` | conteúdo para visualização |
| GET | `/ingest/targets` ou `/storage-roots` | destinos possíveis |

**Já existe em `POST /files/upload`** (multipart, vários arquivos por requisição): parte `files`; campos `target_folder_id` (vazio ou 0 = raiz principal), `on_conflict` (`fail` — padrão —, `skip`, `replace`, `rename`) e `relative_paths` (lista paralela a `files`, para recriar subpastas). Resposta `202`: `{"message", "uploaded": [...], "job_id", "files": [{"name", "path", "status", "error"}]}`, com `status` em `uploaded`, `skipped`, `replaced`, `renamed`, `failed`. O servidor não impõe limite de tamanho nesta rota e ela não muda; depois que o envio em partes existir, os clientes a usam só para arquivos de até 64 MB.

**A criar** — envio em partes com retomada, para arquivos acima de 64 MB e para todo envio do Android em segundo plano. Segue o desenho de `/captures/upload/*`:

| Método | Rota | Uso |
|---|---|---|
| POST | `/files/upload/init` | abre ou retoma uma sessão de envio |
| POST | `/files/upload/chunk` | envia uma parte |
| GET | `/files/upload/:upload_id` | quanto já foi recebido (para retomar após reinício do cliente) |
| POST | `/files/upload/complete` | conclui e publica o arquivo no destino |
| DELETE | `/files/upload/:upload_id` | cancela e apaga o que foi recebido → `204` |

- **`init`** — corpo JSON: `file_name` e `size` (obrigatórios), `target_folder_id`, `relative_path` (subpastas a recriar, opcional), `checksum` (SHA-256 em hexadecimal, opcional), `modified_at` (data ISO da última modificação do original, opcional), `on_conflict` (mesmos valores de `/files/upload`). Resposta `201`: `{"upload_id", "chunk_size", "received_size", "resumed", "already_exists"}`.
  - Sessão aberta com os mesmos `target_folder_id`, `relative_path`, `file_name` e `size` é retomada: `resumed: true` e `received_size` com o ponto de continuação.
  - Com `checksum` informado e arquivo de mesmo conteúdo já no destino: `already_exists: true`, sem `upload_id` — nada é enviado (RN-05).
  - Erros: nome já existe e `on_conflict=fail` → `409`; espaço livre menor que `size` → `507` (RN-07); nome inválido ou destino fora de raiz habilitada → `400`.
- **`chunk`** — multipart: `chunk` (arquivo), `upload_id`, `offset`. `chunk_size` é 2 MiB. Resposta `200`: `{"received_size"}`. `offset` diferente do já recebido → `409` com `{"error", "received_size"}`, para o cliente se realinhar. Sessão inexistente ou expirada → `404`.
- **`complete`** — corpo `{"upload_id"}`. Resposta `201`: `{"file": {"name", "path", "status"}, "job_id"}`, com `status` em `uploaded`, `replaced`, `renamed` ou `skipped` (com `on_conflict: skip` e destino já existente: nada é publicado, a sessão é descartada e `job_id` é `0`). Tamanho recebido diferente do declarado, ou `checksum` informado que não confere → `400`, e a sessão é descartada.
- As partes ficam em área própria, fora das raízes indexadas; o arquivo só aparece no destino no `complete`, por renomeação (ou cópia para nome provisório `.part` seguida de renomeação, entre volumes) — RN-02. `modified_at` é aplicado ao arquivo publicado (RN-10).
- Sessão sem nenhuma parte nova por 24 h expira e é apagada.
- `GET /files/upload/:upload_id` → `{"upload_id", "size", "received_size"}`.

## 8. Estados e erros

| Situação | O que o usuário vê |
|---|---|
| Conexão caiu durante o envio | item "aguardando conexão"; retoma sozinho |
| Sem espaço no servidor | falha com motivo, antes de enviar o primeiro byte (o cliente sempre informa o tamanho — RN-07) |
| Nome já existe | diálogo de conflito (web) / mantém os dois (automático) |
| Dispositivo fora da whitelist | mensagem de acesso não liberado |
| Servidor não encontrado (Android fora de casa) | pendente; nada é perdido |
| Arquivo de origem removido antes de enviar | item falho com motivo |

## 9. Situação em 2026-11-10

- **O app Android não envia nada para o servidor**: a chamada de envio está declarada e nunca é usada; não há seletor de arquivos, "Compartilhar para o KuraNAS" nem envio automático da câmera. A necessidade N1 (foto do celular no PC) não é atendida pelo app — só abrindo a web no navegador do celular.
- O download no Android carrega o arquivo inteiro na memória, sem progresso nem retomada (contraria RN-19 e RF-17).
- Na web, o envio é um arquivo por requisição, com progresso, arrastar e soltar, pasta e política de conflito; não há envio em partes nem retomada (contraria RN-06).

## 10. Critérios de aceite

- [ ] **CA-01** — No PC, arrastar 20 arquivos para uma pasta: o painel mostra o progresso de cada um e, ao fim, os 20 estão na listagem sem recarregar.
- [ ] **CA-02** — Enviar uma pasta com subpastas recria a mesma estrutura no servidor.
- [ ] **CA-03** — Enviar um arquivo de 5 GB pela web conclui; o arquivo no servidor tem o mesmo checksum do original.
- [ ] **CA-04** — Desligar o Wi-Fi no meio de um envio grande e religar: o envio continua de onde parou, sem recomeçar.
- [ ] **CA-05** — Durante um envio, a pasta de destino não mostra arquivo parcial.
- [ ] **CA-06** — Enviar um arquivo cujo nome já existe oferece as três opções; "manter os dois" cria ` (2)`; "aplicar a todos" vale para o lote.
- [ ] **CA-07** — Em um lote de 10 com um arquivo inválido, 9 são enviados e o resumo aponta o que falhou e o motivo.
- [ ] **CA-08** — Uma foto enviada aparece na galeria em até 1 minuto, na data em que foi tirada, com miniatura.
- [ ] **CA-09** — Navegar para outra tela durante o envio não o interrompe; fechar a aba pede confirmação.
- [ ] **CA-10** — Baixar um arquivo com acento e espaço no nome o salva com o nome correto e conteúdo idêntico.
- [ ] **CA-11** — Baixar uma pasta de 3 GB como zip conclui sem erro e o zip abre com a estrutura correta.
- [ ] **CA-12** — Pausar e retomar um download grande no navegador continua de onde parou.
- [ ] **CA-13** — **No Android, na galeria do aparelho, selecionar 5 fotos → Compartilhar → KuraNAS: as 5 chegam à pasta escolhida e aparecem na galeria da web.**
- [ ] **CA-14** — No Android, o envio continua com o app em segundo plano e mostra notificação de progresso e de conclusão.
- [ ] **CA-15** — **Com o envio automático ligado, tirar uma foto em casa: ela aparece na galeria da web em até 2 minutos, sem abrir o app.**
- [ ] **CA-16** — Tirar fotos fora de casa e voltar: elas são enviadas ao reconectar, uma única vez cada.
- [ ] **CA-17** — Reinstalar o app e religar o envio automático não duplica fotos no servidor.
- [ ] **CA-18** — Apagar no celular uma foto já enviada não a remove do servidor.
- [ ] **CA-19** — **No Android, baixar um PDF enviado pelo PC: ele fica em Downloads do aparelho e abre pela notificação.**
- [ ] **CA-20** — No Android, baixar uma foto a faz aparecer na galeria do aparelho.

## 11. Decisões em aberto

- **D-01** — Envio automático da câmera: é o caminho mais direto para "tirei a foto, está no PC", mas é a parte mais cara desta spec. *Provisório: em escopo, implementado depois de "Compartilhar para o KuraNAS" (RF-12) e do envio manual (RF-11), que já resolvem a necessidade com um toque.*
- **D-02** — Pasta de destino padrão do envio automático. *Provisório: `Fotos/Celular/AAAA/MM` na primeira raiz, alterável.*
- **D-03** — Apagar do celular depois de enviar ("liberar espaço"). *Provisório: não; o sistema nunca apaga nada do aparelho.*

## 12. Rastreabilidade

`docs/revisao-produto/01-arquivos.md` lacunas 09 (download) e 11 (upload web). `docs/melhorias/`: 34 e 35 (envio em partes), 24 (enviar e Compartilhar no Android), 36 (download no Android), 37 (segundo plano e tela de transferências), 38 (envio automático da câmera).
