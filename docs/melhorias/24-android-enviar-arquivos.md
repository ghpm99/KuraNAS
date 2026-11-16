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

Verificados em aparelho real.

- [ ] `03/CA-13` — na galeria do aparelho, selecionar 5 fotos → Compartilhar → KuraNAS: as 5 chegam à pasta escolhida e aparecem na galeria da web.
- [ ] No explorer, "Enviar" com 3 arquivos os coloca na pasta aberta e a lista os mostra sem sair da tela.
- [ ] Enviar um arquivo cujo nome já existe cria o segundo com sufixo; nada é sobrescrito.
- [ ] Em um lote com um arquivo que falha, os outros chegam e o resumo aponta o que falhou e por quê.
- [ ] Um vídeo de 200 MB é enviado sem o app fechar por falta de memória.
- [ ] Textos em `strings.xml` (pt-BR e en-US).
- [ ] Testes: repositório monta a requisição com a parte `files` e decodifica a resposta real; view model cobre sucesso, falha parcial e sem servidor.
- [ ] `assembleDebug` e `testDebugUnitTest` verdes.

## Fora de escopo

- Envio com o app fechado, notificação de progresso, retomada após queda (task 37).
- Envio automático da câmera (task 38).
- Diálogo de conflito com escolha (no aparelho o padrão é manter os dois).

## Depende de

Task 22 (sem servidor conectado não há para onde enviar) e task 23 (a escolha de destino navega pelo explorer).
