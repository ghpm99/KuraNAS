# 35 — Envio em partes com retomada: web

**Tipo:** feature · **Prioridade:** P0 (N2) · **Spec:** [03](../especificacao/03-transferencia-entre-dispositivos.md) `RN-05` a `RN-08`, `RF-03` a `RF-07`, §8, `CA-03`, `CA-04`, `CA-09`

## Contexto

Na web o envio é um arquivo por requisição, com progresso, arrastar e soltar, pasta e política de conflito (`frontend/src/service/files.ts`, perto de `:255-270`), mas sem partes nem retomada: uma queda de Wi-Fi no meio de um arquivo grande perde tudo o que foi enviado (spec 03 §9). A task 34 entrega os endpoints.

## Objetivo

Um arquivo grande enviado pelo navegador sobrevive a quedas de conexão e continua de onde parou.

## O que fazer

1. **Serviço** `service/fileUpload.ts`: `initFileUpload`, `sendFileUploadChunk`, `getFileUploadSession`, `completeFileUpload`, `cancelFileUpload`, com os tipos do contrato de 03 §7. Seam test de cada chamada.
2. **Escolha do caminho**: arquivo de até 64 MB continua indo por `POST /files/upload`; acima disso vai em partes. O painel de transferências não distingue os dois.
3. **Envio em partes**: fatia o arquivo no `chunk_size` devolvido pelo `init`, envia em sequência, atualiza progresso, velocidade e tempo restante por arquivo (`RF-03`).
4. **Retomada** (`RN-06`): falha de rede em uma parte põe o item em "aguardando conexão"; ao voltar, consulta a sessão e continua do `received_size`; `409` de posição realinha pelo `received_size` da resposta. Depois de 10 minutos sem conexão, o item vira "falhou", com "Tentar de novo" (que retoma a mesma sessão).
5. **Cancelar** chama `DELETE /files/upload/:upload_id`.
6. **Conflito e espaço**: `409` do `init` abre o diálogo de conflito existente (`RF-06`) e repete o `init` com a política escolhida; `507` vira falha com a mensagem do servidor, antes de enviar.
7. **Já existia** (`RN-05`): para arquivos em partes, o SHA-256 é calculado no navegador antes do `init` (em fluxo, sem carregar o arquivo inteiro); `already_exists` marca o item como "já existia".

## Critérios de aceite

Verificados no sistema rodando.

- [ ] `03/CA-03` — enviar um arquivo de 5 GB conclui; o SHA-256 no servidor é o do original.
- [ ] `03/CA-04` — desligar o Wi-Fi no meio e religar: o envio continua de onde parou.
- [ ] `03/CA-05` — durante o envio, a pasta de destino não mostra arquivo parcial.
- [ ] `03/CA-09` — navegar para outra tela não interrompe; fechar a aba pede confirmação.
- [ ] Enviar de novo o mesmo arquivo grande o marca como "já existia" sem transferir.
- [ ] Cancelar no meio não deixa nada no destino.
- [ ] Um lote com arquivos pequenos e grandes mostra todos no mesmo painel, com até 3 simultâneos (`RF-05`).
- [ ] Render sem mock dos componentes alterados; `make ci` verde.

## Fora de escopo

- Retomar depois de fechar a aba (o navegador perde a referência ao arquivo).
- Envio em partes de vários arquivos em paralelo dentro do mesmo arquivo.

## Depende de

Task 34.
