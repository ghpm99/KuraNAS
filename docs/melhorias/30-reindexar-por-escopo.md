# 30 — "Reindexar" faz o que o rótulo diz

**Tipo:** bug · **Prioridade:** P1 · **Spec:** [21](../especificacao/21-indexacao.md) `RN-15`, `RF-02`, `RF-07`, §8, `CA-11`; [16](../especificacao/16-analytics.md) `RN-07`, `RF-11`, `CA-09`

## Contexto

- **"Reescanear pasta"** (barra de ações de Arquivos, `frontend/src/components/actionBar/actionBar.tsx:105-110`) chama `POST /files/update` com `data=manual-rescan` (`service/files.ts:218-226`) e recarrega a lista em seguida (`fileProvider/index.tsx:297-300`). O handler (`backend/internal/api/v1/files/listing.go:133-155`) ignora o valor e empilha uma varredura de **todas as raízes** (`backend/internal/worker/engine/worker.go:398-415`), responde `200` sem corpo — sem `job_id` — e a tela recarrega antes de o trabalho terminar.
- **"Reindexar" do Analytics** (`components/analytics/AnalyticsOverviewScreen.tsx:60-62`) só refaz as consultas da própria tela (`analyticsProvider/index.tsx:129-131`). Não inicia indexação nenhuma.
- Configurações não tem botão de reindexar.

## Objetivo

Reindexar uma pasta, uma raiz ou tudo inicia um job daquele escopo, mostra o andamento e atualiza a tela quando termina.

## O que fazer

1. **`POST /indexing/reindex`** (contrato de 21 §8): `scope` `all` | `root` | `folder`; `202` com `job_id`; `400`, `404` e `409` conforme o contrato. O job de pasta varre só a pasta e suas subpastas.
2. **`POST /files/update`** passa a responder `{"job_id"}` e continua equivalendo a `scope: all`.
3. **Arquivos — "Reindexar esta pasta"** (`RF-07`): o botão chama o endpoint novo com a pasta aberta, mostra o progresso do job e recarrega a listagem **quando o job conclui**. Na raiz do explorer, o escopo é a raiz aberta. O rótulo `FILES_RESCAN_FOLDER` vira "Reindexar esta pasta".
4. **Analytics — "Reindexar"** (`16/RF-11`): chama `scope: all`, mostra o job iniciado com link para Tarefas; o botão que só recarrega os números chama-se "Atualizar" (`16/RN-07`).
5. **Configurações → Indexação** (`RF-02`): "Reindexar tudo" e reindexar por raiz, com confirmação e progresso.

## Notas da execução (2026-11-17)

- A web não tem nenhuma tela que liste jobs (o sino lista só notificações, e a reindexação não cria notificação), então o link "ver tarefa" do Analytics não foi colocado: o andamento aparece no próprio botão e nos avisos de início e fim. Volta quando a tela de Tarefas existir (item 13 do board de revisão).
- Na raiz do explorer (sem pasta aberta) "Reindexar esta pasta" usa `scope: all`, porque a raiz do explorer mostra todas as raízes juntas.

## Sub-itens levantados na execução (2026-11-17)

- [x] 30.1. Reindexar só reprocessa arquivo novo ou com tamanho/data alterados (o diff da varredura). Arquivo inalterado que ficou sem miniatura, sem metadados ou sem texto enquanto a etapa estava desligada não é recuperado — e a spec 21 §8 diz que "o que ficou sem processar entra pela reindexação". **Decidido pelo dono (2026-11-17)**: reindexar enfileira, para os arquivos do escopo que não mudaram, só as etapas cujo resultado falta — miniatura ausente, linha de metadados ausente, texto de documento não indexado —, respeitando as preferências de indexação. Nunca regrava metadados existentes (`CA-08`).

## Critérios de aceite

- [x] Teste de handler: os três escopos devolvem `job_id`; escopo inválido `400`; pasta inexistente `404`; segundo pedido do mesmo escopo em andamento `409`.
- [x] Teste do worker: o job de pasta não visita arquivos fora dela.
- [x] `21/CA-11` — no sistema rodando: "Reindexar esta pasta" reindexa só aquela pasta, mostra progresso e a listagem se atualiza ao concluir. *Task 47 (2026-11-24), backend local: `POST /indexing/reindex {"scope":"folder","folder_id":988}` → `202 {"job_id":6291}`, cujas etapas (`scan_filesystem`, `diff_against_db`, `mark_deleted`, `fill_missing_stages`) têm só o caminho `Imagens/teste-task47`; o job é acompanhável em `/jobs/:id` e a listagem da pasta traz o arquivo novo ao concluir. Barra de progresso na tela não exercitada (sem navegador).*
- [x] `16/CA-09` — "Reindexar" no Analytics inicia um job visível em `GET /jobs`. *Task 47 (2026-11-24), backend local: `{"scope":"all"}` → job 6286 (`reindex_folder`), listado em `GET /jobs?type=reindex_folder`; repetir enquanto não há outro em andamento abre novo job.*
- [x] `21/CA-08` — reindexar não remove estrelas, playlists, álbuns, progresso nem metadados corrigidos. *Task 47 (2026-11-24), backend local: antes e depois de reindexar tudo: 200 estrelas, 2 playlists de música (uma manual com 1 faixa e uma por regra), 6 correções de metadados de música, 1 correção de episódio (o vídeo 958 continua como T1E11) e 31 linhas de progresso de vídeo (11 assistidos) — tudo igual.*
- [x] Seam test de cada chamada nova no frontend; consumidores de `POST /files/update` conferidos (web; Android e plugin não usam).
- [x] `make ci` verde.

## Fora de escopo

- Tela de Tarefas na web (item 13 do board de revisão). Até ela existir, o link de "ver tarefa" leva ao sino, que já lista jobs em andamento.
- Reprocessar falhas de processamento (`16/RF-10`).
