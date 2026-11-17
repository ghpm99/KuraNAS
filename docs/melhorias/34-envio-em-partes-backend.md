# 34 — Envio em partes com retomada: backend

**Tipo:** feature · **Prioridade:** P0 (N1, N2) · **Spec:** [03](../especificacao/03-transferencia-entre-dispositivos.md) `RN-02`, `RN-04` a `RN-07`, `RN-10`, §7, `CA-03` a `CA-05`

## Contexto

Arquivos comuns só podem ser enviados por `POST /files/upload`, em uma requisição única: se a conexão cai aos 3,9 GB de um vídeo de 4 GB, o envio recomeça do zero. O envio em partes com retomada já existe para capturas e Takeout (`/captures/upload/*`, `/takeout/upload/*`; partes de 2 MiB, sessão em disco, `offset` conferido) — `backend/internal/api/v1/captures/service.go:113-329` é a referência mais completa (retomada, validação de `upload_id`). O `/files/upload` atual não confere espaço livre nem checksum e não recebe a data original do arquivo.

## Objetivo

Qualquer cliente envia um arquivo de qualquer tamanho em partes, retoma de onde parou e só vê o arquivo no destino quando está completo e íntegro.

## O que fazer

No domínio `files` (envio é comportamento genérico de arquivo). Um endpoint por commit, cada um com handler → service → repository e teste de handler decodificando a resposta real.

1. **Sessão de envio**: área própria fora das raízes indexadas (como a de capturas), com o estado da sessão em disco: destino, nome, tamanho declarado, checksum declarado, data original, política de conflito, bytes recebidos, última atividade. `upload_id` validado em toda leitura.
2. **`POST /files/upload/init`**: valida nome (`02/RN-05`), destino dentro de raiz habilitada, espaço livre (`507`), conflito (`409` com `on_conflict=fail`); retoma sessão aberta equivalente (`resumed`, `received_size`); com `checksum` e arquivo idêntico já no destino, responde `already_exists` sem abrir sessão.
3. **`POST /files/upload/chunk`**: grava a parte na posição exata; `offset` diferente → `409` com `received_size`.
4. **`GET /files/upload/:upload_id`**.
5. **`POST /files/upload/complete`**: confere tamanho e checksum; aplica `on_conflict`; publica no destino por renomeação (ou `.part` + renomeação entre volumes); aplica `modified_at`; registra o arquivo no índice na mesma requisição (`02/RN-04`) e enfileira o processamento, devolvendo `job_id`.
6. **`DELETE /files/upload/:upload_id`**.
7. **Expiração**: sessão sem parte nova há 24 h é apagada por rotina periódica.
8. **Capturas — alinhamento do erro de posição**: `POST /captures/upload/chunk` responde `409` com `received_size` no lugar do `500` atual (contrato de 12 §8). Consumidor a conferir: `plugin/src/background/uploader.js`.

## Critérios de aceite

- [x] Teste: enviar em partes, interromper, chamar `init` de novo e continuar do `received_size` produz arquivo com o mesmo SHA-256 do original.
- [x] Teste: `offset` errado responde `409` com `received_size`; nada é gravado.
- [x] Teste: `complete` com tamanho ou checksum divergente responde `400`, não publica e descarta a sessão.
- [x] Teste: `init` de conteúdo já existente (por checksum) responde `already_exists`.
- [x] Teste: espaço insuficiente responde `507` antes de qualquer parte.
- [x] Teste: durante o envio, o destino não contém o arquivo nem `.part` visível à listagem (`03/CA-05`).
- [x] Teste: `modified_at` informado é a data de modificação do arquivo publicado.
- [x] Teste contra Postgres: o arquivo aparece na listagem da pasta logo após o `complete`, com workers desligados.
- [x] `make ci-backend` verde (cobertura ≥ 80 %).

## Fora de escopo

- Clientes (tasks 35 e 37).
- Remover ou mudar `POST /files/upload` (o contrato só cresce).
- Envio de pasta como unidade: o cliente envia arquivo a arquivo com `relative_path`.
