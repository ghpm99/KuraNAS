# 36 — Android: baixar arquivos em fluxo, para a pasta Downloads

**Tipo:** bug · **Prioridade:** P0 (N2) · **Spec:** [03](../especificacao/03-transferencia-entre-dispositivos.md) `RN-18`, `RN-19`, `RN-23`, `RF-17`, `RF-18`, `CA-19`, `CA-20`

## Contexto

O app baixa carregando **o arquivo inteiro na memória**: `feature/files/data/FilesRepository.kt:66-69` faz `body.bytes()` sobre `GET api/v1/files/blob/{id}` (declarado sem `@Streaming` em `FilesApi.kt:76-77`) e só então `FileViewerViewModel.kt:64-84` grava em Downloads. Um vídeo de 1 GB derruba o app. Não há progresso, notificação nem retomada. O endpoint próprio de download com retomada, `GET /files/download/:id`, não é usado pelo app.

## Objetivo

Baixar no celular um arquivo de qualquer tamanho que subiu pelo PC, com progresso, e abri-lo ao terminar.

## O que fazer

1. **API**: `downloadFile(id)` sobre `GET api/v1/files/download/{id}` com `@Streaming`, aceitando cabeçalho `Range`.
2. **Gravação em fluxo**: o corpo é copiado em blocos direto para o destino do `MediaStore` (pasta Downloads; fotos e vídeos em coleção que a galeria do aparelho enxerga — `RN-23`), marcado como pendente até terminar. Nome original preservado, com acentos e espaços (`RN-18`).
3. **Progresso** por bytes recebidos sobre o tamanho do arquivo, na tela e em notificação.
4. **Retomada** (`RN-19`): queda de conexão no meio continua de onde parou com `Range`; três tentativas, depois falha com a mensagem do servidor ou de conexão.
5. **Conclusão**: notificação "Abrir" que entrega o arquivo ao aplicativo adequado (`RF-17`).
6. **"Baixar" em seleção múltipla** no explorer: um download por arquivo, em fila.
7. **"Abrir com…" e "Compartilhar"** (`RF-18`): baixa em fluxo para o cache do app e entrega por `FileProvider`.
8. O visualizador interno de imagem e de texto continua usando `/files/blob/{id}`, limitado a arquivos de até 20 MB; acima disso oferece "Baixar".

## Critérios de aceite

Verificados no sistema local (decisão do dono, 2026-11-17): servidor local + testes do app com MockWebServer; o que depende de aparelho está anotado.

- [x] `03/CA-19` — baixar um PDF enviado pelo PC: fica em Downloads e abre pela notificação. *Coleção Downloads para `application/pdf` e notificação `ACTION_VIEW` cobertas por testes; abrir pelo aparelho não exercitado.*
- [x] `03/CA-20` — baixar uma foto a faz aparecer na galeria do aparelho. *MIME pela extensão (`MimeTypeMap`) → `image/*` vai para `MediaStore.Images` (`Pictures/KuraNAS`); galeria do aparelho não exercitada.*
- [x] Baixar um arquivo de 2 GB conclui sem o app fechar; o tamanho no aparelho é igual ao do servidor. *Cópia em blocos de 64 KB, sem `bytes()` (teste com servidor real em blocos); 2 GB no aparelho não exercitado.*
- [x] Desligar o Wi-Fi no meio e religar: o download continua, sem recomeçar. *Servidor local responde `206` com `Content-Range` a `Range: bytes=100-199`; app retoma com `Range: bytes=<recebido>-` (teste com `DISCONNECT_DURING_RESPONSE_BODY`), 4 quedas com avanço ainda concluem.*
- [x] Um arquivo com acento e espaço no nome é salvo com o nome correto. *Nome vem do item da listagem, não do `Content-Disposition` (que o servidor envia sem `filename*`); teste com acento e espaço.*
- [x] Arquivo ausente no servidor mostra a mensagem do `404`, sem criar arquivo vazio (`RN-22`). *Servidor local: `404 {"error":"Arquivo não encontrado"}`; app só cria a entrada pendente na primeira resposta bem-sucedida (teste).*
- [x] Testes: cópia em fluxo com `Range` a partir de um ponto; view model cobre progresso, falha e conclusão.
- [x] `assembleDebug` e `testDebugUnitTest` verdes.

## Fora de escopo

- Baixar pasta como zip no app.
- Downloads com o app fechado (task 37 traz a execução em segundo plano).
- Conteúdo offline de música e vídeo (`24 §4`).
