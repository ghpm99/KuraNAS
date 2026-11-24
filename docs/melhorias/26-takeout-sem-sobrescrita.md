# 26 — Takeout: preservar a estrutura do zip e nunca sobrescrever

**Tipo:** bug crítico (perda de dado) · **Prioridade:** P0 (N7) · **Spec:** [13](../especificacao/13-takeout.md) `RN-04`, `RN-05`, `RN-06`, `RN-08`, `RN-11`, `CA-03`, `CA-04`, `CA-09`

## Contexto

`backend/internal/api/v1/takeout/extractor.go:47-50` monta o destino como `<biblioteca>/takeout/<nome base do arquivo>`, descartando as pastas do zip (`:136` usa `filepath.Base(entry.Name)`), e grava com `os.WriteFile` (`:167`) sem verificar se o nome existe. Um Takeout do Google Fotos tem centenas de `IMG_0001.JPG` em álbuns diferentes: **o último vence e os outros são perdidos**. Enviar o mesmo zip duas vezes regrava tudo. Não há contagem do que foi importado, pulado ou ignorado.

## Objetivo

Toda foto e todo vídeo do zip chega ao servidor, na pasta de origem, sem sobrescrever nada e sem duplicar o que já existe.

## O que fazer

1. **Destino com estrutura** (`RN-05`): `<biblioteca do tipo>/takeout/<caminho da pasta de origem dentro do zip>/<arquivo>`, removendo o prefixo comum `Takeout/Google Fotos/` (e `Takeout/Google Photos/`). `safeExtractionPath` continua barrando caminho que sai do destino e links simbólicos (`RN-08`).
2. **Sem sobrescrita** (`RN-06`): se o destino existe e o conteúdo é idêntico (mesmo tamanho e mesmo SHA-256), a entrada é pulada e contada como "já existente"; se é diferente, o arquivo novo recebe sufixo ` (2)`, ` (3)`…
3. **Contagem** (`RN-11`): o extrator devolve `imported`, `already_existing`, `ignored` (não é foto nem vídeo, ou entrada maliciosa) e `failed_entries`, com `total_entries`.
4. **Registro da importação**: tabela `takeout_import` com as colunas do item de `GET /takeout/imports` (13 §7); a linha é criada no `complete`, atualizada pelo step e fechada com o resumo.
5. **Notificação** de conclusão e de falha com o resumo traduzido (hoje a de falha usa o erro cru — `backend/internal/worker/engine/step_takeout.go:33`).

## Critérios de aceite

- [x] Teste do extrator: zip com `Álbum A/IMG_0001.JPG` e `Álbum B/IMG_0001.JPG` de conteúdos diferentes gera os dois arquivos, cada um em sua pasta.
- [x] Teste do extrator: mesmo nome, mesma pasta, conteúdo diferente gera o segundo com ` (2)`.
- [x] Teste do extrator: extrair o mesmo zip duas vezes não cria nem regrava arquivo; a segunda contagem é toda "já existente".
- [x] Teste do extrator: entrada com `../` e link simbólico são ignorados e contados.
- [x] `13/CA-03`, `CA-04`, `CA-09` — no sistema rodando, com um zip real de Takeout. *Task 47 (2026-11-24), backend local, zip montado no formato do Takeout (`Takeout/Google Fotos/<álbum>/…` com sidecars JSON e um `.html`) enviado por `/takeout/upload/init|chunk|complete`: as duas `IMG_0001.png` de "Viagem 2019" e "Natal 2020" foram importadas em `Imagens/takeout/<álbum>/` com conteúdos e datas próprios (CA-03); o resumo (`takeout_import` 4) deu 3 entradas = 2 importadas + 1 ignorada (CA-09); o mesmo zip de novo (import 5) deu 0 importadas, 2 já existentes, 1 ignorada, sem duplicar arquivo (CA-04). Um zip real exportado do Google não foi usado. As rotas de histórico `GET /takeout/imports` ainda não existem (spec 13 §8, "A criar").*
- [x] A notificação de falha não contém erro interno nem caminho do servidor.
- [x] `make ci-backend` verde (cobertura ≥ 80 %).

## Fora de escopo

- Aplicar as datas e a localização dos JSON (task 27).
- Extração em fluxo (task 28).
- Endpoints e tela de histórico (`GET /takeout/imports`), retomada e cancelamento do envio: item 11 do board de revisão, que usa a tabela criada aqui.
