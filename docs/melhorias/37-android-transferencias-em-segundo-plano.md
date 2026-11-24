# 37 — Android: transferências em segundo plano, com retomada

**Tipo:** feature · **Prioridade:** P0 (N1) · **Spec:** [03](../especificacao/03-transferencia-entre-dispositivos.md) `RN-06`, `RN-08`, `RF-13`, `RF-14`, §8, `CA-14`; [24](../especificacao/24-app-android.md) `RF-03`, `RF-06`

## Contexto

Depois das tasks 24 e 36 o app envia e baixa, mas só com a tela aberta e em requisição única: sair do app interrompe, e um arquivo grande não retoma. A dependência do WorkManager já está no projeto (`android/app/build.gradle.kts:85`), sem nenhum `Worker` escrito. A task 34 entrega o envio em partes.

## Objetivo

Envios e downloads continuam com o app fechado, mostram notificação de progresso e de resultado, e retomam sozinhos depois de uma queda de rede.

## O que fazer

1. **Fila persistente** (Room): uma linha por transferência — direção, origem (`Uri` com permissão persistida) ou `file_id`, destino, tamanho, bytes transferidos, `upload_id`, estado (`na fila`, `em andamento`, `aguardando conexão`, `concluída`, `já existia`, `falhou`, `cancelada`) e motivo da falha.
2. **Worker de envio**: usa `/files/upload/init|chunk|complete`, com `checksum` e `modified_at`; ao recomeçar, consulta `GET /files/upload/:upload_id` e continua do `received_size`. Restrição de rede conectada; nova tentativa com espera crescente.
3. **Worker de download**: o download em fluxo com `Range` da task 36, sob o mesmo modelo.
4. **Notificações**: progresso (uma por lote, com contagem e porcentagem) e resultado (enviados, já existiam, com falha); tocar abre a tela de Transferências.
5. **Tela "Transferências"** (`RF-14`): em andamento, concluídas e com falha; cancelar e tentar de novo por item; limpar concluídas.
6. **Ligação com as telas existentes**: "Enviar" do explorer e "Compartilhar para o KuraNAS" (task 24) e "Baixar" (task 36) passam a enfileirar na fila, no lugar da chamada direta.
7. **Indicador** de transferências em andamento na navegação (`24/RF-06`) e entrada "Transferências" no menu "Mais".

## Decisões de desenho (2026-11-24)

- **Fila** em Room (`transfers`), uma linha por arquivo, agrupada por `batch_id` (lote = uma ação do usuário). A máquina de estados é uma função pura (`TransferStateMachine`) testada à parte; o DAO só persiste.
- **Execução**: um `CoroutineWorker` por transferência (`@HiltWorker`), `ExistingWorkPolicy.KEEP` com nome único `transfer-<id>`, restrição `NetworkType.CONNECTED`, backoff exponencial, `setForeground` com tipo `dataSync`. O WorkManager já reagenda após reiniciar o aparelho.
- **Origem do envio**: `Uri` do seletor do explorer (`OpenMultipleDocuments`) recebe `takePersistableUriPermission`; `Uri` de "Compartilhar para o KuraNAS" (permissão temporária) é copiado para `filesDir/outbox/<id>` no ato de enfileirar e o arquivo é apagado quando a transferência termina ou é cancelada.
- **Envio**: sempre em partes (`init|chunk|complete`), `checksum` SHA-256 calculado no worker antes do `init`, `modified_at` da origem quando disponível, `on_conflict=rename` (automático mantém os dois, §8). Ao recomeçar com `upload_id` salvo, `GET /files/upload/:id` dá o `received_size`; `404` reabre com `init`. `409` no chunk realinha pelo `received_size` do corpo. `already_exists` vira estado `já existia`. Cancelar chama `DELETE /files/upload/:id`.
- **Download**: o worker reaproveita o `StreamingDownloader`, persistindo bytes recebidos e o `Uri` pendente do `MediaStore` para continuar com `Range` após o processo morrer.
- **Navegação**: a barra inferior de `24/RF-03` é da task 39; aqui "Transferências" entra como rota própria alcançável pelo drawer atual e pela notificação, com o indicador de andamento no drawer.

## Critérios de aceite

Verificados em aparelho real.

- [ ] `03/CA-14` — o envio continua com o app em segundo plano e mostra notificação de progresso e de conclusão.
- [ ] Fechar o app (remover dos recentes) durante o envio de 20 fotos: todas chegam.
- [ ] Desligar o Wi-Fi no meio de um vídeo de 1 GB e religar: continua de onde parou; o SHA-256 no servidor confere.
- [ ] Reiniciar o aparelho com transferência pendente: ela retoma depois de ligar.
- [ ] Um arquivo de origem apagado antes do envio vira item com falha e motivo, sem travar a fila.
- [ ] Cancelar um envio em andamento não deixa arquivo no destino.
- [ ] Testes: máquina de estados da fila; worker de envio retomando de `received_size`; realinhamento no `409` de posição.
- [ ] `assembleDebug` e `testDebugUnitTest` verdes.

## Fora de escopo

- Envio automático da câmera (task 38).
- Limitar por Wi-Fi ou por carregador (essas opções são do envio automático).

## Depende de

Tasks 24, 34 e 36.
