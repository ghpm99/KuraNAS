# 11 — Takeout (importação do Google Fotos)

Escopo: `backend/internal/api/v1/takeout/` (sessão de envio em partes, extrator, sidecars, `ImportRepository`), step `takeout_extract` em `backend/internal/worker/engine/step_takeout*.go` e `step_image_takeout_sidecar.go`, queries em `backend/pkg/database/queries/takeout/`, rotas em `backend/internal/app/routes.go:663-672`, `frontend/src/service/takeout.ts`, `frontend/src/components/takeout/` e `frontend/src/pages/takeout/`. Fora de escopo: jobs e notificações como tela (item 13), indexação de imagens (item 03), filtro de pasta da galeria e do catálogo de vídeos (itens 03 e 06). Android não tem Takeout; a spec é só web.

## Diagnóstico (2026-11-25)

Spec: `docs/especificacao/13-takeout.md`. A seção "Situação em 2026-11-10" está defasada: as tasks 26, 27 e 28 de `docs/melhorias/` resolveram estrutura, sem sobrescrita, datas, descrição, localização e extração em fluxo; a tabela `takeout_import` e o `import_id` no `complete` já existem.

**Backend — já cumpre**: destino `takeout/<pastas>/` (`RN-05`); sem sobrescrita, idêntico pulado e ` (2)` no conflito (`RN-06`, `CA-03`, `CA-04`); entradas perigosas ignoradas (`RN-08`); extração em fluxo (`RN-09`); contagens sem o `.json` (`RN-11`, `CA-09`); data, descrição e GPS (`RN-07`, `CA-02`); limpeza do zip e da sessão (`RN-10`, `CA-10`); job com progresso (`RN-03`); falha com chave i18n e erro cru só no `applog`; notificação de conclusão; `/jobs?type=takeout_import`; menu e rota `/takeout`.

**Backend — lacunas**
- `init` aceita `size` 0 e qualquer `file_name`; extensão só conferida no `complete` e sensível a maiúsculas (`service.go:56-82, ~112`) — `RN-01`, `§7`.
- Sem retomada: `init` sempre gera `upload_id` novo (`service.go:66`) — `RN-02`, `CA-05`.
- Faltam `GET`/`DELETE /takeout/upload/:upload_id`, `GET /takeout/imports` e `GET /takeout/imports/:id` (`routes.go:669-671`) — `RF-02`, `RF-03`, `RF-05`, `CA-06`.
- Sem `507` por espaço livre no `init`; `chunk` aceita além do `size` declarado — `RN-09`, `RN-02`.
- "Importação iniciada" é notificada no `init`, antes de qualquer byte; conclusão sem `metadata` com `import_id` — `RN-11`, `CA-11`.
- Falha no meio da extração descarta o parcial: arquivos gravados ficam sem indexar e sem contagem (`extractor.go:~158`, `step_takeout.go:35-45`) — `RN-12`.
- Falha só distingue zip inválido; sem espaço/sem permissão cai em "veja o log" — `§8`, `RN-12`. Texto de zip inválido não diz "corrompido" — `CA-08`.
- `classifyFile` tem lista própria de vídeo (ignora `.3gp`, `.mpg`, `.m2ts`…, aceita `.ts`) em vez de `utils.VideoFormats` — `RN-04`.
- `complete` repetido cria segundo job — `RN-06`, `RN-10`.
- Erros: `SETTINGS_LIBRARY_SAVE_ERROR` de outra feature em `init`/`chunk`; sessão inexistente `400` (deveria `404`); `TAKEOUT_UPLOAD_INCOMPLETE` mostra `%d` literal; `file_name` vazio vira `500` — `00/RN-40`.
- Textos pt-BR sem acento; chaves sem uso — `00/RN-43`.

**Frontend**: um único arquivo, sem fila; sem explicação de como exportar; erro descarta a mensagem do servidor e mostra `%s` literal; sem retomada; sem velocidade, tempo restante ou "Cancelar"; após o `complete` só "job #N" literal, sem etapa de importação, resumo, atalhos nem histórico; sair durante o envio não pede confirmação; rótulos de outra feature (`SETTINGS_RESET`); sem render sem mock (`00/RN-30`).

**Não verificado**: `CA-01` (10 GB), `CA-05` e `CA-06` ponta a ponta; parâmetro de pasta das telas de imagens e vídeos; se o scan ignora `.takeout-*.part` órfão; reexecução do job ao reiniciar.

## Lacunas (uma alteração cada)

### Backend
- [x] 01. `init` recusa `size` ausente ou ≤ 0 e `file_name` sem `.zip` (sem diferenciar maiúsculas) com `400` e mensagem própria; `complete` compara a extensão sem diferenciar maiúsculas — `RN-01`, `§7`, `CA-08`.
- [x] 02. `init` com o mesmo `file_name` e `size` de uma sessão aberta devolve o mesmo `upload_id`, `received_size` e `resumed: true`; nova sessão devolve `received_size: 0`, `resumed: false` — `RN-02`, `§7`, `CA-05`. Seam no handler.
- [x] 03. `GET /takeout/upload/:upload_id` devolve `{upload_id, size, received_size}`; `404` quando não existe — `RN-02`, `§7`.
- [x] 04. `DELETE /takeout/upload/:upload_id` apaga a sessão, `204`; `404` quando não existe — `RF-02`, `§7`.
- [x] 05. `init` responde `507` quando o espaço livre é menor que `size` — `RN-09`, `§7`, `§8`. Depende da 01.
- [x] 06. `chunk` recusa parte que ultrapassa o `size` declarado — `RN-02`.
- [x] 07. Erros do envio com chaves `TAKEOUT_*` próprias, sessão inexistente `404`, `TAKEOUT_UPLOAD_INCOMPLETE` com `i18n.Translate`, `file_name` vazio `400` — `00/RN-40`, `RN-41`. Depende da 01.
- [x] 08. Notificação "Importação iniciada" sai do `init` para o `complete`, com `event` e `import_id` no metadata — `RN-11`, `CA-11`.
- [ ] 09. Notificações de conclusão e falha levam `metadata` `{event, import_id, job_id}` — `RN-11`, `CA-11`.
- [x] 10. `complete` repetido responde `409` e não cria segundo job — `RN-06`, `RN-10`. Depende da 07.
- [ ] 11. `GET /takeout/imports` paginado (`created_at DESC, id DESC`) com os campos de `§7` — `RF-05`, `00/RN-17`, `RN-52`. Seam + teste `_Postgres`.
- [ ] 12. `GET /takeout/imports/:id`; `404` quando não existe — `RF-03`, `CA-06`. Depende da 11.
- [ ] 13. O extrator devolve o `ExtractResult` acumulado junto do erro — `RN-12`.
- [ ] 13.1. Na falha, o step grava os sidecars e enfileira a indexação do que já foi gravado antes de fechar como falha — `RN-12`, `CA-09`. Depende da 13.
- [ ] 13.2. `CloseImportFailed` grava também as contagens — `RN-12`. Depende da 13.
- [ ] 14. Falta de espaço ou de permissão durante a importação falha com motivo traduzido próprio — `§8`, `RN-12`. Depende da 13.
- [ ] 15. Mensagens de zip inválido dizem que o arquivo está corrompido ou ilegível — `RN-12`, `CA-08`.
- [ ] 16. `classifyFile` usa `utils.VideoFormats` e a classificação de imagem compartilhada — `RN-04`, `06/RN-01`.
- [ ] 17. Textos pt-BR com acento nas chaves `TAKEOUT_*` e remoção das chaves sem uso — `00/RN-43`.
- [ ] 18. Testes de handler fixam os campos de `init`, `chunk` e `complete` — `00/CA-16`. Só teste.

### Frontend
- [ ] 19. `service/takeout.ts`: status e cancelamento da sessão, lista e detalhe de importações; `received_size`/`resumed` no tipo do `init`; payload sem `items` é erro — `00/RN-11`, `§7`. Depende da 02, 03, 04, 11, 12. Seam test.
- [x] 20. Erro do envio mostra a mensagem do servidor; fallback `TAKEOUT_UPLOAD_ERROR` sem `%s` — `RF-04`, `00/RN-41`.
- [ ] 21. Fila de vários zips, um por vez, recusa individual de não-`.zip` — `RF-01`, `RN-01`.
- [ ] 22. Área de soltar e `<input>` aceitam vários arquivos — `RF-01`. Depende da 21.
- [x] 23. Explicação curta de como gerar a exportação no Google — `RF-01`.
- [ ] 24. Retomada do envio pelo `received_size` com espera crescente — `RN-02`, `CA-05`. Depende da 02 e 19.
- [ ] 25. "Cancelar" por zip chama `DELETE /takeout/upload/:upload_id` — `RF-02`. Seam. Depende da 04 e 19.
- [ ] 26. Velocidade e tempo restante na barra do envio — `RF-02`.
- [ ] 27. Etapa de importação: consulta `GET /takeout/imports/:id` a cada 2 s e mostra processados/total; remove o `job #N` literal — `RF-02`, `CA-06`. Depende da 12 e 19.
- [ ] 28. Rótulos próprios de iniciar e limpar; progresso por interpolação — `00/RN-43`, `RN-44`.
- [ ] 29. Resumo ao concluir e atalhos para as fotos e vídeos importados — `RF-03`, `CA-11`. Depende da 27.
- [ ] 30. Falha mostra o `error` do servidor e "Tentar de novo" — `RF-04`, `RN-06`. Depende da 27.
- [ ] 31. Histórico de importações com scroll infinito — `RF-05`, `00/RN-02 a RN-10`. Depende da 11 e 19.
- [ ] 32. Importação em andamento aparece ao reabrir a tela, com atualização a cada 2 s — `CA-06`, `RF-06`. Depende da 31.
- [ ] 33. Sair da tela durante o envio pede confirmação — `RF-06`.
- [ ] 34. Notificação de conclusão leva às fotos importadas — `CA-11`. Depende da 09 e 29.
- [ ] 35. Teste de render sem mock de `TakeoutImportScreen` e `TakeoutDropZone` — `00/RN-30`.

Ordem: backend 01 → 18 (13 antes de 13.1, 13.2 e 14); frontend 20, 23, 28, 33, 35 independem do backend; 19 depois de 02, 03, 04, 11 e 12; 21 antes de 22 e 24; 27 antes de 29, 30, 32.

### Decisões do dono
- 🚫 D1. Expirar sessões abandonadas em `.takeout_uploads/` depois de N dias, ou manter.
- 🚫 D2. `13/D-01` e `D-02` (álbuns do Google viram álbuns; organização por data) seguem com resposta provisória.
