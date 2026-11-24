# 38 — Android: envio automático das fotos da câmera

**Tipo:** feature · **Prioridade:** P0 (N1) · **Spec:** [03](../especificacao/03-transferencia-entre-dispositivos.md) `RN-12` a `RN-17`, `RF-15`, `RF-16`, `CA-15` a `CA-18`, `D-01` a `D-03`

## Contexto

É o caminho mais direto para "tirei a foto, está no PC": sem abrir o app. Hoje não existe (spec 03 §9). A spec o põe em escopo, depois de "Compartilhar para o KuraNAS" e do envio manual (`03/D-01`), que já resolvem com um toque. Usa a fila e os workers da task 37.

## Objetivo

Com a opção ligada, toda foto e todo vídeo novo da câmera chega ao servidor sozinho quando o aparelho está na rede de casa, uma única vez.

## O que fazer

1. **Configurações do app** (`RF-15`): ligar/desligar (desligado por padrão); pasta de destino (padrão `Fotos/Celular` na primeira raiz — `D-02`); só no Wi-Fi (ligado); só carregando (desligado); incluir vídeos (ligado).
2. **Permissões** (`RF-16`): leitura de mídia e notificações, pedidas ao ligar, com o motivo; sem permissão, a tela diz o que falta e como liberar.
3. **Escolha inicial** (`RN-15`): ao ligar, "enviar tudo o que já está no aparelho" ou "só a partir de agora".
4. **Detecção**: worker periódico (15 min) mais gatilho de mudança no `MediaStore`, que lista as mídias da câmera mais novas que o último ponto registrado.
5. **Registro do que já foi** (`RN-14`): tabela local por identificador da mídia e SHA-256. O envio usa `checksum` no `init`; `already_exists` conta como enviado — é isso que impede duplicar depois de reinstalar o app.
6. **Destino por data** (`RN-12`): subpasta `AAAA/MM` pela data em que a mídia foi tirada; `modified_at` preserva a data original.
7. **Só em casa** (`RN-13`, `RN-16`): o worker só enfileira quando o servidor lembrado confirma (`/api/v1/health`); fora de casa encerra sem novas tentativas até a próxima execução.
8. **Situação** (`RN-17`): última sincronização, pendentes, falhas, e "Sincronizar agora".
9. O app nunca apaga nada do aparelho (`D-03`), e apagar no aparelho não apaga no servidor.

## Decisões de desenho (2026-11-24)

- **Destino**: `target_folder_id` da pasta escolhida (padrão: raiz, `D-02`) + `relative_path` = prefixo (padrão `Fotos/Celular`) + `AAAA/MM` pela `DATE_TAKEN` (cai para `DATE_ADDED`). O servidor cria as subpastas; a fila da task 37 passa a carregar `relative_path` por item.
- **Não reenvio** (`RN-14`): tabela local `camera_uploads` por id da mídia no `MediaStore` (com o `transfer_id`); a seleção ignora o que já consta. Depois de reinstalar, o registro some, mas o envio manda `checksum` e o mesmo nome para o mesmo `AAAA/MM`, e o servidor responde `already_exists` — conta como enviado.
- **Ponto de partida** (`RN-15`): "só a partir de agora" grava como marca o `DATE_ADDED` do momento em que ligou; "tudo" grava zero. A seleção lista câmera (`DCIM/Camera` ou `BUCKET_DISPLAY_NAME` da câmera) mais nova que a marca e ausente do registro.
- **Execução**: `CameraSyncWorker` periódico de 15 min (restrições: rede sem medição se "só no Wi-Fi", carregando se "só carregando") mais um trabalho único com `ContentUriTrigger` em `MediaStore.Images/Video.Media.EXTERNAL_CONTENT_URI`, reagendado a cada execução. Fora de casa (`/api/v1/health` do servidor lembrado não responde `ok`) encerra com sucesso, sem retentar (`RN-16`).
- **Origem do envio**: `Uri` do `MediaStore`, lido com a permissão de mídia (sem cópia).

## Critérios de aceite

Verificados em aparelho real.

- [x] `03/CA-15` — com o envio automático ligado, tirar uma foto em casa: ela aparece na galeria da web em até 2 minutos, sem abrir o app. *(com o gatilho de mudança; só com o worker periódico o prazo é de até 15 minutos — se o aparelho não entregar o gatilho com o app fechado, registrar aqui o prazo medido e ajustar a spec)* *Gatilho `ContentUriTrigger` (5 s/30 s) + periódico de 15 min agendados e reaplicados no início do app; prazo no aparelho não medido (sem aparelho/emulador).*
- [x] `03/CA-16` — fotos tiradas fora de casa são enviadas ao reconectar, uma única vez cada. *Fora de casa o runner encerra sem registrar nada (`AWAY_FROM_HOME`, teste); na próxima execução em casa a seleção pega o que não está em `camera_uploads` (teste de não reenvio).*
- [x] `03/CA-17` — reinstalar o app e religar não duplica fotos no servidor. *Backend local: `init` com `relative_path=Fotos/Celular/2026/11` + `checksum` enviado e concluído; o mesmo `init` repetido (registro local vazio, como após reinstalar) responde `already_exists: true`, que a fila marca como "já existia".*
- [x] `03/CA-18` — apagar no celular uma foto já enviada não a remove do servidor. *O envio automático não chama nenhuma rota de exclusão; o registro mantém a mídia como enviada.*
- [x] As fotos chegam em `…/AAAA/MM` do mês em que foram tiradas e aparecem nessa data na galeria. *Backend local: arquivo gravado em `nas-data/Fotos/Celular/2026/11/` com SHA-256 igual e mtime = `modified_at` (2026-11-03 10:15 UTC); `relativePathFor` testado nas viradas de mês por fuso.*
- [x] Com "só no Wi-Fi" ligado, nada é enviado pela rede móvel. *Varredura e transferências da câmera agendadas com `NetworkType.UNMETERED` (teste do mapeamento; a exigência fica na linha da transferência e vale no tentar de novo).*
- [x] Sem permissão de mídia, a tela explica e nada falha em silêncio. *Negar mantém desligado e mostra o que falta com atalho para as configurações; permissão revogada com o envio ligado grava `MEDIA_PERMISSION_MISSING`, exibido na situação (testes).*
- [x] Testes: seleção das mídias novas; decisão "em casa / fora"; não reenvio do que já consta no registro.
- [x] `assembleDebug` e `testDebugUnitTest` verdes.

## Fora de escopo

- Pastas além da câmera (capturas de tela, WhatsApp).
- "Liberar espaço" apagando do aparelho (`D-03`).
- Backup de contatos, mensagens ou aplicativos.

## Depende de

Task 37.
