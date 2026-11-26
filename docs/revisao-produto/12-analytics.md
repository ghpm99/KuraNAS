# 12 — Analytics

Escopo: `backend/internal/api/v1/analytics/`, queries em `backend/pkg/database/queries/analytics/`, rotas em `backend/internal/app/routes.go:544-558` e `/files/top-files-by-size`, `frontend/src/service/analytics.ts`, `frontend/src/components/analytics/`, `frontend/src/components/providers/analyticsProvider/` e `frontend/src/pages/analytics/`. Fora de escopo: indexação e jobs como tela (item 13, spec 21), lixeira (item 08), Home (item 18). Android não tem tela de Analytics; a Home do app lê `storage`, `types` e `recent-files` (`HomeApi.kt:18-25`), então o contrato só cresce.

## Diagnóstico (2026-11-25)

Spec: `docs/especificacao/16-analytics.md`. A seção "Situação em 2026-11-10" está defasada: as tasks 30 e 31 de `docs/melhorias/` resolveram o "Reindexar", os links das listas e criaram `/analytics/largest` e `/analytics/duplicates`.

**Backend — já cumpre**: um endpoint por informação (13 rotas, `RN-03`); `period` validado com `400` i18n e erro interno sem vazar (`00/RN-40`); caminhos relativos à raiz (`RN-06`, `CA-05`); duplicados paginados com filtro e ordem estável (`RN-05`, `CA-08`); tipos pelas listas canônicas; Reindexar por `POST /indexing/reindex` (`RN-07`, `CA-09`); telas de maiores e duplicados com scroll infinito e seam test (`RF-07`, `RF-09`).

**Backend — lacunas**
- Espaço total/livre só do `EntryPoint`; falha de `statfs` vira total = usado em silêncio (`service.go:343-353`) — `RN-02`, `CA-01`.
- Faltam `/analytics/trash`, `/analytics/processing/failures`, `POST /analytics/processing/retry-failed` — §7, `RF-02`, `RF-10`, `CA-11`. Sem espaço por tier em `storage` — `RF-02`.
- `type_distribution.sql` devolve `archive` além das cinco categorias — `RF-04`, `CA-02`.
- `/analytics/library` sem números por tipo de mídia; `/analytics/processing` só `metadata`/`thumbnail` — `RF-10`.
- `health.recent_errors` cru do banco na tela — `00/RN-40`, `RN-41`.
- Uso de IA em memória, zera ao reiniciar, sem provedor/função/período — `RF-12`, §8.
- `ERROR_ANALYTICS_LOAD` cita "overview"; textos pt-BR sem acento — `00/RN-43`. Comentários em função e em `.sql`; `log.Printf` em vez de `applog`.
- Testes de handler só conferem `200`; nenhum fixa os campos — `00/CA-16`.

**Frontend — lacunas**
- `AnalyticsProvider` junta 13 consultas: só `storage` vira erro; os demais blocos caem em zero/"ok"/"Sem dados" (`analyticsProvider/index.tsx:41-139`) — `RN-03`, `CA-04`. Nenhum bloco tem "Tentar de novo".
- O provedor envolve a página inteira: `/analytics/duplicates` e `/largest` disparam 13 consultas sem usar — `RN-03`, `00/RN-52`.
- Período em `useState`, não na URL — `RN-04`, `CA-03`.
- Visão geral sem tier, lixeira, total de pastas, gráfico, pastas que mais cresceram, bloco de maiores — `RF-02`, `RF-03`, `RF-06`, `RF-07`.
- Rótulos de tipo e extensão crus (`archive`, `<none>`) — `RF-04`, `RF-05`, i18n.
- Bloco de duplicados sem cópias, caminhos nem link — `RN-06`, `RF-08`.
- Biblioteca sem tabela por tipo, lista de falhas, "Reprocessar falhas", "Reindexar" — `RF-10`, `RF-11`, `CA-11`. Sem uso de IA por provedor/função — `RF-12`.
- Sem render sem mock de `AnalyticsOverviewScreen`, `AnalyticsLibraryScreen`, `AnalyticsContent` — `00/RN-30`.

**Não verificado**: `CA-01` a `CA-04`, `CA-09` a `CA-11` no sistema rodando; se as queries excluem backup/temporários (`RN-01`); se o `payload` de `worker_step` traz o id do arquivo (necessário para a 06); se falha reprocessada com sucesso ainda conta; decode do Android contra `storage`, `types`, `recent-files`.

## Lacunas (uma alteração cada)

### Spec
- [x] 01. Spec 16: §8 atualizada (tasks 30 e 31 resolvidas) e §7 documenta `tiers: {hot_bytes, cold_bytes}` em `/analytics/storage` — `RF-02`. Commit próprio, antes do código.

### Backend
- [ ] 02. Espaço total e livre somam os volumes das raízes habilitadas sem contar o mesmo volume duas vezes; falha de `statfs` não vira total = usado em silêncio — `RN-02`, `CA-01`.
- [ ] 03. `storage` devolve `tiers.hot_bytes` e `tiers.cold_bytes` — `RF-02`, `RN-01`. Depende da 01.
- [x] 04. `GET /analytics/trash` → `{item_count, size_bytes}` pela mesma fonte de `/trash/summary` — `RF-02`, §7. Seam no handler.
- [x] 05. `type_distribution.sql` devolve só as cinco categorias (`archive` em `other`) — `RF-04`, `CA-02`. Teste `_Postgres` da soma.
- [x] 06. Query, repositório e service de falhas de processamento `{file_id, name, stage, error, failed_at}`, `error` traduzido, filtro `stage`, ordem estável — `RF-10`, §7, `00/RN-17`, `RN-41`.
- [ ] 07. Handler e rota `GET /analytics/processing/failures` paginados, `400` em `stage` inválido — `RF-10`, `CA-11`. Depende da 06.
- [ ] 08. `POST /analytics/processing/retry-failed` (`stage` opcional), `202` com `job_id` ou `200` com `job_id: null` — `RF-10`, §7. Depende da 06.
- [ ] 09. `/analytics/library` ganha `total`, `with_metadata`, `pending`, `failed` por tipo de mídia — `RF-10`.
- [x] 10. `health.recent_errors` com mensagem traduzida em vez do erro cru — `00/RN-40`, `RN-41`.
- [ ] 11. Uso de IA persistido por chamada (provedor, função, instante, sucesso, tokens, latência) — `RF-12`, §8.
- [ ] 12. `/analytics/ai-usage` aceita `period` e devolve `by_provider` e `by_function` — `RF-12`, `RN-04`. Depende da 11.
- [x] 13. `ERROR_ANALYTICS_LOAD` sem "overview" e textos `ANALYTICS_*` pt-BR com acento — `00/RN-43`.
- [x] 14. Remove comentários de `service.go` e dos `.sql` de saúde; `log.Printf` vira `applog` — só refatoração.
- [x] 15. Testes de handler fixam os campos de todos os endpoints de analytics — `00/CA-16`. Só teste.

### Frontend
- [ ] 16. `service/analytics.ts`: lixeira, falhas paginadas, reprocessar, tipos novos de `tiers`, biblioteca e IA; formato inválido é erro — `00/RN-11`, §7. Depende da 03, 04, 07, 08, 09, 12. Seam do reprocessar.
- [x] 17. Hook `useAnalyticsBlockQuery` (uma consulta por bloco, sem fallback de zero) e `AnalyticsSection` com "Tentar de novo" — `RN-03`, `00/RN-52`.
- [x] 18. Período em `?period=` na URL — `RN-04`, `CA-03`.
- [ ] 19. Bloco de indicadores com consulta própria (usado, livre, total, crescimento, arquivos e pastas) — `RF-02`, `RN-03`, `CA-04`. Depende da 17 e 18.
- [ ] 20. Indicadores de tier e de lixeira — `RF-02`, `RN-01`, `RN-03`. Depende da 16 e 17.
- [ ] 21. Uso ao longo do tempo como gráfico — `RF-03`, `RN-03`. Depende da 17 e 18.
- [ ] 22. Blocos de tipo e extensão com rótulos i18n — `RF-04`, `RF-05`, `CA-02`. Depende da 05 e 17.
- [ ] 23. Pastas que mais ocupam e que mais cresceram, cada linha abre a pasta — `RF-06`, `RN-06`, `CA-05`. Depende da 17 e 18.
- [ ] 24. Bloco "maiores arquivos" com os 10 primeiros e "ver todos" — `RF-07`, `CA-06`. Depende da 17.
- [ ] 25. Bloco de duplicados com cópias, caminhos e link — `RF-08`, `RN-05`, `RN-06`. Depende da 17.
- [ ] 26. Remove `AnalyticsProvider` e os fallbacks; as telas de maiores e duplicados deixam de disparar consultas sem uso — `RN-03`, `00/RN-52`. Depende de 19 a 25 e 29 a 31.
- [ ] 27. "Atualizar" recarrega só os blocos montados; "atualizado há…" da consulta mais antiga — `RF-01`, `RN-07`, `RN-08`. Depende da 26.
- [ ] 28. Sem texto cru em saúde, tipos e extensões — `00/RN-43`. Depende da 10 e 22.
- [ ] 29. Biblioteca: tabela por tipo de mídia — `RF-10`, `CA-10`. Depende da 09 e 17.
- [ ] 30. Biblioteca: lista de falhas com scroll infinito e filtro por etapa — `RF-10`, `CA-11`. Depende da 07 e 16.
- [ ] 31. Biblioteca: "Reprocessar falhas" com confirmação — `RF-10`, `RN-07`. Seam. Depende da 08, 16 e 30.
- [ ] 32. Biblioteca: bloco de saúde com "Reindexar" — `RF-11`, `CA-09`. Depende da 17.
- [ ] 33. Número de falhas vira link para o bloco de falhas — `RN-06`. Depende da 30.
- [ ] 34. Uso de IA por provedor e função — `RF-12`. Depende da 12, 16 e 17.
- [ ] 35. Render sem mock de `AnalyticsOverviewScreen`, `AnalyticsLibraryScreen` e `AnalyticsContent` — `00/RN-30`.
- [ ] 36. Render sem mock de cada componente novo, no mesmo commit do componente — `00/RN-30`.

Ordem: 01 → 02, 03; 04, 05, 06, 09, 10, 11, 13, 14, 15 independentes; 07 e 08 após 06; 12 após 11; 16 após 03, 04, 07, 08, 09, 12; 17 e 18 antes de 19–25; 26 após 19–25 e 29–31; 27 após 26.

### Decisões do dono
- 🚫 D1. `16/D-01`: tela de sugestões por IA ou remoção de `GET /analytics/insights`.
- 🚫 D2. Card "Arquivos em timeout recorrente" não está na spec 16: manter ou tirar.
