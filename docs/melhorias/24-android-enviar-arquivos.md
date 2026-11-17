# 24 — Android: enviar arquivos e "Compartilhar para o KuraNAS"

**Tipo:** feature (lacuna de necessidade central) · **Prioridade:** P0 (N1) · **Spec:** [03](../especificacao/03-transferencia-entre-dispositivos.md) `RN-01`, `RN-04`, `RN-08`, `RF-11`, `RF-12`, §7, `CA-13`; [24](../especificacao/24-app-android.md) `CA-09`

## Contexto

"Tirar foto no celular e ver no PC" (N1) não é possível pelo app. `uploadFile` está declarado em `feature/files/data/FilesApi.kt:52-56` e nunca é chamado; não há seletor de arquivos (o único `ActivityResultContracts` do app é o de permissão, em `MainActivity.kt`), nem `ACTION_SEND`/`ACTION_SEND_MULTIPLE` no `AndroidManifest.xml`. A declaração existente está errada: usa a parte `file`, e o servidor lê a parte **`files`** (`backend/internal/api/v1/files/operations.go:80`).

O servidor já aceita o envio (`POST /api/v1/files/upload`, contrato em 03 §7): parte `files`, `target_folder_id`, `on_conflict` (`fail`, `skip`, `replace`, `rename`), resposta `202` com o resultado por arquivo.

## Objetivo

Enviar um ou vários arquivos do aparelho para uma pasta do servidor, de dentro do explorer e pelo "Compartilhar" de qualquer aplicativo.

Esta task entrega o envio **com o app aberto**, por `POST /files/upload`. Envio em segundo plano, retomada e arquivos grandes são da task 37.

## O que fazer

1. **API e repositório**: `uploadFile` passa a usar a parte `files`, `target_folder_id` e `on_conflict=rename` (padrão dos envios do aparelho — `RN-04`), lendo o conteúdo por fluxo a partir do `Uri` (sem carregar o arquivo na memória); o DTO de resposta inclui `files[{name, path, status, error}]`.
2. **"Enviar" no explorer** (`RF-11`): botão na pasta aberta que abre o seletor do sistema com seleção múltipla e envia para a pasta atual.
3. **Progresso e resultado**: durante o envio, a tela mostra quantos de quantos e o arquivo atual; ao fim, o resumo (enviados, renomeados, com falha e motivo — `RN-08`) e a lista da pasta recarregada.
4. **Alvo de compartilhamento** (`RF-12`): `intent-filter` de `ACTION_SEND` e `ACTION_SEND_MULTIPLE` para `*/*`; ao receber, abre a escolha de destino (raízes e pastas, navegando pelo explorer), com o último destino usado pré-selecionado.
5. **"Sempre enviar para cá"**: opção na escolha de destino; marcada, os próximos compartilhamentos vão direto para aquela pasta (com como desfazer em Configurações).
6. **Sem servidor ou fora da whitelist**: a tela de envio diz o motivo e não perde a seleção (`03 §8`).

## Critérios de aceite

Verificados no sistema local (decisão do dono, 2026-11-17): API local com o mesmo multipart do app + testes unitários; o que depende de aparelho está anotado.

- [x] `03/CA-13` — na galeria do aparelho, selecionar 5 fotos → Compartilhar → KuraNAS: as 5 chegam à pasta escolhida e aparecem na galeria da web. *Local: `POST /files/upload` com `files`, `target_folder_id=884`, `on_conflict=rename` → arquivos na pasta e com `image_metadata` (galeria); intent de compartilhamento coberto por `ShareIntentUrisTest`; fluxo na galeria do aparelho não exercitado (sem aparelho).*
- [x] No explorer, "Enviar" com 3 arquivos os coloca na pasta aberta e a lista os mostra sem sair da tela. *Local: a verificação achou o explorer listando por rota errada (task 23, corrigido em `246be07f`); agora `/files/tree?file_parent=884` lista os enviados; recarga após envio coberta por `UploadViewModelTest`.*
- [x] Enviar um arquivo cujo nome já existe cria o segundo com sufixo; nada é sobrescrito. *Local: 2º envio de `foto.jpg` → `status: renamed`, `foto (2).jpg`; o original intacto.*
- [x] Em um lote com um arquivo que falha, os outros chegam e o resumo aponta o que falhou e por quê. *Um arquivo por requisição; falha parcial coberta por `UploadViewModelTest` e pela resposta real com `failed` em `UploadFilesRepositoryTest`.*
- [x] Um vídeo de 200 MB é enviado sem o app fechar por falta de memória. *`UriRequestBody` lê o `Uri` por fluxo em `writeTo` (teste com `InputStream` contador); 200 MB no aparelho não exercitado.*
- [x] Textos em `strings.xml` (pt-BR e en-US).
- [x] Testes: repositório monta a requisição com a parte `files` e decodifica a resposta real; view model cobre sucesso, falha parcial e sem servidor.
- [x] `assembleDebug` e `testDebugUnitTest` verdes.

## Fora de escopo

- Envio com o app fechado, notificação de progresso, retomada após queda (task 37).
- Envio automático da câmera (task 38).
- Diálogo de conflito com escolha (no aparelho o padrão é manter os dois).

## Depende de

Task 22 (sem servidor conectado não há para onde enviar) e task 23 (a escolha de destino navega pelo explorer).
