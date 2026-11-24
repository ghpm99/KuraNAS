# 25 — Capturas: excluir uma captura promovida não apaga o vídeo da biblioteca

**Tipo:** bug crítico (perda de dado) · **Prioridade:** P0 (N7 — "sem perder arquivo") · **Spec:** [12](../especificacao/12-capturas.md) `RN-11`, `RF-04`, §8, `CA-09`

## Contexto

`DeleteCapture` (`backend/internal/api/v1/captures/service.go:467-481`) apaga a linha da captura e faz `os.Remove(model.FilePath)`, sem olhar o estado. Na promoção (`captures/service_promote.go:88-96`) a captura recebe `FileID`, `Status = CaptureStatusPromoted` e **`FilePath = finalPath`** — o caminho do vídeo já dentro da biblioteca. Resultado: excluir uma captura promovida apaga do disco o vídeo da biblioteca, sem passar pela lixeira, e deixa a linha do arquivo órfã até o watcher perceber. O erro do `os.Remove` é ignorado. A tela exclui sem confirmação. Não há teste de serviço para esse caminho.

## Objetivo

Excluir uma captura promovida remove só o registro da captura; o vídeo continua na biblioteca.

## O que fazer

1. **Serviço**: `DeleteCapture` só remove o arquivo do disco quando a captura **não** está promovida (sem `FileID` e estado diferente de promovida). Em captura promovida, apaga só a linha. Falha ao remover o arquivo de uma captura não promovida é registrada em log (hoje é descartada).
2. **Resposta**: `DELETE /captures/:id` devolve `{"deleted": true, "library_file_kept": true|false}` (contrato de 12 §8).
3. **Web — confirmação** (`RF-04`): excluir pede confirmação em diálogo da aplicação. O texto diz, conforme o estado: "o arquivo recebido será apagado" ou "o vídeo continua na biblioteca; para removê-lo, exclua o vídeo". Textos no catálogo de traduções.

## Critérios de aceite

- [x] Teste de serviço: excluir captura promovida não remove o arquivo de `FilePath` e remove a linha da captura.
- [x] Teste de serviço: excluir captura não promovida remove o arquivo recebido e a linha.
- [x] Teste de handler decodifica `deleted` e `library_file_kept`; seam test do `DELETE` no frontend.
- [x] `12/CA-09` — no sistema rodando: excluir uma captura promovida não remove o vídeo de Vídeos nem do disco; a confirmação diz isso. *Local 2026-11-17: captura MP4 real promovida para `Videos/Captura boa.mp4` (file_id 910); `DELETE /captures/19` → `{deleted:true, library_file_kept:true}`, arquivo no disco e linha ativa mantidos, captura 404; texto da confirmação coberto pelos testes do frontend.*
- [x] Excluir uma captura que falhou na promoção apaga o arquivo recebido. *Não verificável localmente hoje: nenhuma promoção falha — ver 25.1.* *Task 47 (2026-11-24), depois do 25.1, backend local: captura 21 com 200 KB aleatórios como `.mp4` → `failed` com o motivo traduzido, `file_path` no arquivo recebido e nada em Vídeos; `DELETE /captures/21` → `{"deleted":true,"library_file_kept":false}` e o arquivo recebido sumiu do disco.*
- [x] `make ci` verde.

### Sub-itens

- [x] 25.1. Promoção que não consegue ajustar o vídeo falha de verdade (`12/RN-07`, `RN-09`): hoje, se o remux do ffmpeg falha, `placeRecordingInLibrary` (`captures/service_promote.go:128`) move o arquivo assim mesmo e a captura fica `promoted` — verificado localmente: um `.mp4` com bytes aleatórios entrou em Vídeos. Além disso, o rollback não devolve o `file_path` da captura ao arquivo recebido (já foi trocado pelo destino em `UpdateCapturePromotion`), então "Tentar de novo" leria um caminho inexistente, e o estado "falhou" é gravado sem o motivo.

## Fora de escopo

- "Tentar de novo", filtros, paginação e atualização automática da tela de Capturas (item 10 do board de revisão).
- Recuperar vídeos já apagados por este defeito.
