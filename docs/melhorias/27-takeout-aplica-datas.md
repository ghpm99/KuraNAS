# 27 — Takeout: aplicar a data, a descrição e a localização de cada foto

**Tipo:** bug · **Prioridade:** P0 (N3 — achar foto antiga) · **Spec:** [13](../especificacao/13-takeout.md) `RN-07`, `CA-02`; [04](../especificacao/04-imagens.md) `RN-02`, §5 ("data da foto")

## Contexto

O extrator lê os `.json` que acompanham cada mídia e monta `Metadata *TakeoutMetadata` em cada `ExtractedFile` (`takeout/extractor.go:104-124`, `:171-183`; campos em `takeout/model.go:16-28`: `PhotoTakenTime`, `CreationTime`, `GeoData`…). Quem consome o resultado só usa `DestinationPath` (`backend/internal/worker/engine/step_takeout.go:40-48`): **os metadados são descartados**. Fotos sem EXIF (capturas de tela, imagens recebidas por mensagem, fotos antigas editadas) entram na galeria com a data da importação, todas amontoadas no dia do envio. Além disso, o JSON é associado pelo nome base, que colide entre pastas.

## Objetivo

Cada foto importada aparece na galeria no dia em que foi tirada.

## O que fazer

1. **Associação correta**: o JSON é ligado à mídia pelo caminho completo dentro do zip (mesma pasta), cobrindo os formatos de nome do Takeout: `foto.jpg.json`, `foto.jpg.supplemental-metadata.json` e o nome truncado em 46 caracteres.
2. **Data do arquivo**: depois de gravar a mídia, a data de modificação do arquivo vira `photoTakenTime` (ou `creationTime`, na falta dele), por `os.Chtimes`. Como a "data da foto" é o EXIF e, sem EXIF, a data de modificação (`04 §5`), a galeria passa a posicionar a foto certo sem mudar a indexação.
3. **Descrição e localização**: gravadas no complemento de metadados da imagem **só quando o arquivo não as tem** — o EXIF do arquivo sempre prevalece.
4. O JSON em si não é importado como arquivo e não entra na contagem de "ignorados".

## Critérios de aceite

- [ ] Teste do extrator: mídia sem EXIF com JSON ao lado termina com data de modificação igual a `photoTakenTime`.
- [ ] Teste do extrator: dois `IMG_0001.JPG` em pastas diferentes recebem cada um a data do seu próprio JSON.
- [ ] Teste do extrator: os três formatos de nome de JSON são reconhecidos.
- [ ] Mídia com EXIF mantém a data do EXIF na galeria.
- [ ] `13/CA-02` — no sistema rodando: uma foto sem EXIF, cuja data está só no JSON, aparece na galeria na data correta.
- [ ] `make ci-backend` verde.

## Fora de escopo

- Criar álbuns a partir dos álbuns do Google (`13/D-01`).
- Organizar o destino por data (`13/D-02`).
- Gravar EXIF dentro do arquivo.

## Depende de

Task 26 (a associação por caminho usa a estrutura preservada).
