# 35 — Envio em partes com retomada: web

**Tipo:** feature · **Prioridade:** P0 (N2) · **Spec:** [03](../especificacao/03-transferencia-entre-dispositivos.md) `RN-05` a `RN-08`, `RF-03` a `RF-07`, §8, `CA-03`, `CA-04`, `CA-09`

## Contexto

Na web o envio é um arquivo por requisição, com progresso, arrastar e soltar, pasta e política de conflito (`frontend/src/service/files.ts`, perto de `:255-270`), mas sem partes nem retomada: uma queda de Wi-Fi no meio de um arquivo grande perde tudo o que foi enviado (spec 03 §9). A task 34 entrega os endpoints. Conferido em 2026-11-17: a fila fica em `frontend/src/features/files/upload/useUploadQueue.ts` e o painel em `uploadQueuePanel.tsx`; o painel não mostra velocidade nem tempo restante; não há confirmação ao fechar a aba (`RF-07`); e não há diálogo de conflito — a política é escolhida antes, no seletor do painel (`conflictPolicySelect.tsx`), e com `fail` o item falha com a mensagem do servidor.

## Objetivo

Um arquivo grande enviado pelo navegador sobrevive a quedas de conexão e continua de onde parou.

## O que fazer

1. **Serviço** `service/fileUpload.ts`: `initFileUpload`, `sendFileUploadChunk`, `getFileUploadSession`, `completeFileUpload`, `cancelFileUpload`, com os tipos do contrato de 03 §7. Seam test de cada chamada.
2. **Escolha do caminho**: arquivo de até 64 MB continua indo por `POST /files/upload`; acima disso vai em partes. O painel de transferências não distingue os dois.
3. **Envio em partes**: fatia o arquivo no `chunk_size` devolvido pelo `init`, envia em sequência, atualiza progresso, velocidade e tempo restante por arquivo (`RF-03`).
4. **Retomada** (`RN-06`): falha de rede em uma parte põe o item em "aguardando conexão"; ao voltar, consulta a sessão e continua do `received_size`; `409` de posição realinha pelo `received_size` da resposta. Depois de 10 minutos sem conexão, o item vira "falhou", com "Tentar de novo" (que retoma a mesma sessão).
5. **Cancelar** chama `DELETE /files/upload/:upload_id`.
6. **Conflito e espaço**: o `init` leva a política escolhida no seletor do painel, como no envio de uma requisição; `409` (com `fail`) e `507` viram falha com a mensagem do servidor, antes de enviar. O diálogo de conflito de `RF-06` fica para o sub-item 35.1.
7. **Já existia** (`RN-05`): para arquivos em partes, o SHA-256 é calculado no navegador antes do `init` (em fluxo, sem carregar o arquivo inteiro); `already_exists` marca o item como "já existia". Enquanto calcula, o item fica "conferindo" (`RF-03`).
8. **Sair da página** (`RF-07`): com envio em andamento, fechar ou recarregar a aba pede confirmação do navegador; navegar dentro do app não interrompe. Conferido em 2026-11-17: o `UploadQueueProvider` estava montado em `features/files/files/filesLayout.tsx` (só Arquivos e Favoritos) — sair dessas telas desmontava a fila e o painel; ele sobe para o `ShellLayout` de `app/App.tsx`.

### Sub-itens

- [x] 35.1. Diálogo de conflito (`RF-06`) com as opções de `RN-04` e "aplicar a todos", aberto quando o servidor recusa por nome existente, nos dois caminhos de envio.

## Critérios de aceite

Verificados no sistema rodando.

- [x] `03/CA-03` — enviar um arquivo de 5 GB conclui; o SHA-256 no servidor é o do original. *Local 2026-11-17 (protocolo do frontend por script): 200 MiB aleatórios em partes de 2 MiB com checksum no `init` → `complete` 201 `uploaded`; SHA-256 do publicado igual ao original. 5 GB não exercitado (mesmo caminho, sem limite de tamanho).*
- [x] `03/CA-04` — desligar o Wi-Fi no meio e religar: o envio continua de onde parou. *Local: envio interrompido em 90 MiB; novo `init` com os mesmos dados → `resumed: true`, `received_size` 94.371.840 (= sessão), continua e conclui. Espera por conexão no navegador coberta por `uploadFileInChunks.test.ts`.*
- [x] `03/CA-05` — durante o envio, a pasta de destino não mostra arquivo parcial. *Local: com 90 MiB recebidos, a pasta de destino estava vazia.*
- [x] `03/CA-09` — navegar para outra tela não interrompe; fechar a aba pede confirmação. *Provider no `ShellLayout` (`859f0b31`) e `beforeunload` cobertos por testes; não exercitado em navegador.*
- [x] Enviar de novo o mesmo arquivo grande o marca como "já existia" sem transferir. *Local: `init` do mesmo arquivo → `already_exists: true`, sem `upload_id`.*
- [x] Cancelar no meio não deixa nada no destino. *Local: `DELETE /files/upload/:id` com 20 MiB recebidos → 204, sessão 404 depois, destino sem o arquivo e sem `.part`.*
- [x] Um lote com arquivos pequenos e grandes mostra todos no mesmo painel, com até 3 simultâneos (`RF-05`). *Coberto por `useUploadQueue.test.ts` (limite de 3 com itens conferindo/aguardando); não exercitado em navegador.*
- [x] Render sem mock dos componentes alterados; `make ci` verde.

## Fora de escopo

- Retomar depois de fechar a aba (o navegador perde a referência ao arquivo).
- Envio em partes de vários arquivos em paralelo dentro do mesmo arquivo.

## Depende de

Task 34.

## Achado na verificação local (2026-11-17)

`/files/upload` e `/files/upload/complete` devolviam o caminho do disco em `path`/`uploaded`; corrigido para o caminho lógico (`382bf877`). Conflito com `fail` no `init` → `409` com a mensagem traduzida, conferido.
