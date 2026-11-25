# 10 — Capturas (plugin)

Escopo: `backend/internal/api/v1/captures/` (upload em partes, sessões em disco, promoção para a biblioteca de vídeos, lista/detalhe/exclusão), step `capture_process` em `backend/internal/worker/engine/step_capture_promote.go`, queries em `backend/pkg/database/queries/captures/`, `frontend/src/service/captures.ts`, `frontend/src/components/captures/` e `frontend/src/pages/captures/`, e a extensão `plugin/` (uploader, downloader, gravação híbrida, descoberta do servidor, popup). Fora de escopo: ingestão por URL no servidor (`ingest`, item 09), o catálogo de downloads do próprio KuraNAS (`distribution`), notificações como tela (item 13). Android não consome `/captures` e captura pelo Android é fora de escopo (`12 §4`).

## Diagnóstico (2026-11-25)

Spec: `docs/especificacao/12-capturas.md`. A seção "Situação em 2026-11-10" da spec está parcialmente defasada: `DELETE` já preserva o vídeo e devolve `library_file_kept`, o item já traz `failure_reason` e `chunk` fora da posição já responde `409` com `received_size`.

**Backend — o que já cumpre a spec**
- Envio em partes com `409` + `received_size` para parte fora da posição (`handler.go:139-146`) — `RN-01`.
- Sessão aberta com a mesma `episode_key` é retomada (`service.go:128-137, 189-233`) — `RN-02` (primeira metade).
- Concluir com zero bytes recusado com `400` (`service.go:280-287`) — `RN-03` (primeira metade).
- Staging em `<capturas>/.uploads/<id>/` fora das raízes indexadas — `RN-04`.
- Promoção em segundo plano (`capture_process`), destino sem sobrescrita, remux ffmpeg, pôster como fonte da miniatura, rollback na falha de colocação — `RN-05` a `RN-09`, `CA-11`.
- Exclusão preserva vídeo promovido — `RN-11`. Notificações de envio e promoção — `RN-12`.
- `CA-13`: busca por `drm|decrypt|widevine|clearkey` não achou código de decriptação.

**Backend — lacunas**
- Promoção só falha "bonito" na etapa de colocar o arquivo; erros anteriores em `PromoteCapture` (`service_promote.go:46-102`) deixam a captura `uploaded` para sempre, sem motivo nem notificação — `RN-09`, `RN-12`, `CA-10`.
- Status `promoting` existe (`model.go:13`) mas nada o grava — `RF-03`, `CA-07`.
- Não existe `POST /captures/:id/promote` — `§8`, `RF-04`, `CA-10`.
- `GET /captures` sem filtros `status`/`platform` (`handler.go:200-206`) — `§8`, `00/RN-52`.
- Ordenação sem desempate por `id` (`get_captures.sql`) — `00/RN-17`.
- Tamanho diferente do declarado no `complete` vira `500` genérico (`service.go:289-297`) — `RN-03`, `00/RN-40`.
- `upload_id` inexistente em `chunk`/`complete` vira `500`; `DELETE` de inexistente vira `500`; `GET /captures/:id` responde `404` para qualquer erro — `00/RN-40`.
- Erro cru (`uploadErr.Error()`, `promoteErr.Error()`) gravado no metadata das notificações (`service.go:404`, `service_promote.go:307`), que chega ao cliente — `00/RN-40`.
- Rollback da promoção não apaga o pôster gravado — `RN-09`.
- `RN-02` (segunda metade) divergente: o servidor regrava episódio completo por decisão antiga; `AlreadyComplete` nunca é devolvido (🚫 D1).
- Sessões abandonadas em `.uploads/` nunca expiram (🚫 D3). Captura em envio não tem registro (🚫 D2).
- Testes de handler não fixam os campos da lista, detalhe, `init` e `complete` — `00/CA-16`.
- Comentários de corpo no pacote `captures` e em `dto.go` (cita o sidecar `metadata.json` aposentado); testes de promoção gravam em `captures/etc/kuranas/thumbnails/` fora de `t.TempDir()`.

**Frontend**
- `useCapturesScreen.ts:28-32`: uma única `useQuery` — só os 50 primeiros, sem scroll infinito — `00/RN-02`, `CA-06`.
- Sem busca, filtro, recarregar, revalidação por foco nem atualização automática — `RF-02`, `RF-03`, `CA-07`.
- Tipo `Capture` sem `failure_reason`; sem "Assistir", "Ver na pasta", "Tentar de novo" — `RF-04`, `RF-05`.
- Erro de carga com texto fixo `CAPTURES_ERROR`, sem mensagem do servidor nem "Tentar de novo" — `00/RN-10`, `RN-41`. Payload sem `items` vira vazio (teste fixa isso) — `00/RN-11`.
- Excluir sem retorno visível de sucesso/erro; `library_file_kept` ignorado — `00/RN-33`, `CA-14`.
- Estado vazio não explica o que são capturas — `RF-06`. Data com `toLocaleDateString()` — `00/RN-44`.
- `CAPTURES_PAGE_EYEBROW` sem uso. Sem render da tela sem mock. Comentário em `service/captures.ts:48-50`.

**Plugin**
- Downloads (HLS, DASH, direto, blob) não enviam metadados nem `episode_key` (`uploader.js:51-62`) — `RN-16`, `CA-05`.
- Gravação híbrida desiste após 3 tentativas de 200/400 ms e descarta o resto em silêncio (`background.js:~400-427`) — `RN-01`, `CA-04`.
- "Salvar" do popup grava qualquer endereço sem validar `/health` — `RN-15`, `RF-12`, `CA-12`. Popup não mostra a situação da conexão; `403` tratado como inalcançável — `RF-07`, `§9`.
- "Baixar"/"Capturar" descartam a resposta — `RF-11`.
- README não cita a limitação do quadro preto em conteúdo protegido — `§9`. Comentários de corpo em vários arquivos.

**Android**: não consome `/captures`; fora de escopo.

**Não verificado (só no sistema rodando)**: `CA-01` a `CA-05`, `CA-12`; `RN-13`/`RN-14` (máquina de estados da gravação); reexecução do job `capture_process` ao reiniciar o servidor; rota do player para "Assistir".

## Lacunas (uma alteração cada)

### Backend
- [x] 01. `PromoteCapture` grava `promoting` ao começar (método novo no repositório, `.sql` próprio que também zera `failure_reason`) — `RF-03`, `CA-07`. Pronto quando: teste do serviço cobre `uploaded` → `promoting` → `promoted`/`failed`.
- [x] 02. Falha de promoção antes da colocação do arquivo marca a captura `failed` com motivo i18n, preserva o recebido e emite a notificação de falha (chaves próprias nos dois idiomas) — `RN-09`, `RN-12`, `§9`, `CA-10`. Depende da 01.
- [x] 03. `POST /captures/:id/promote`: `202` com `{"job_id"}` via `CreateCaptureProcessJob`; `409` quando não está `failed`; `404` quando não existe — `§8`, `RF-04`, `CA-10`. Seam: teste do handler decodifica `job_id`.
- [x] 04. `GET /captures` aceita `status` (`uploaded|promoting|promoted|failed`) e `platform`, `400` em valor inválido — `§8`, `RF-02`, `00/RN-52`.
- [x] 05. `get_captures.sql` ordena por `created_at DESC, id DESC` — `00/RN-17`.
- [x] 06. Concluir com tamanho diferente do declarado responde `400` com mensagem própria; a sessão continua em disco — `RN-03`, `00/RN-40`.
- [x] 07. `upload_id` inexistente ou inválido em `chunk` e `complete` responde `404` com mensagem própria — `00/RN-40`.
- [x] 08. `DELETE /captures/:id` de inexistente responde `404`; `GET /captures/:id` só responde `404` quando não existe (falha de banco vira `500`) — `00/RN-40`.
- [ ] 09. Notificações de falha deixam de gravar o erro cru no metadata; gravam o motivo traduzido; o erro cru fica só em `applog` — `00/RN-40`, `RN-42`.
- [ ] 10. Rollback da promoção também remove o pôster gravado para o `file_id` desfeito — `RN-09`.
- [ ] 11. Testes de handler decodificam a resposta real e fixam os campos da lista, detalhe, `init` e `complete` — `00/CA-16`. Só teste.
- [ ] 12. Comentários de corpo removidos do pacote `captures` e de `step_capture_promote.go`; testes deixam de gravar `etc/` dentro do pacote. Sem mudança de comportamento.
- [ ] 13. `CAPTURES_PAGE_EYEBROW` removida dos dois idiomas (conferir uso antes) — `00/RN-43`.

### Frontend
- [ ] 14. `service/captures.ts`: filtros `status`/`platform`, `failure_reason` no tipo, `promoteCapture(id)`, adaptador de paginação compartilhado, payload sem `items` é erro — `00/RN-11`, `RN-53`, `§8`. Depende da 03 e 04. Seam test.
- [ ] 15. Lista com scroll infinito conforme `00` — `RF-01`, `00/RN-02 a RN-09`, `CA-06`.
- [x] 16. Erro da primeira carga mostra a mensagem do servidor com "Tentar de novo"; forma inválida é erro — `00/RN-10`, `RN-11`, `RN-41`.
- [ ] 17. "Recarregar" no cabeçalho e revalidação ao voltar o foco — `00/RN-12 a RN-14`, `CA-07`.
- [ ] 18. "Tentar de novo" no item `failed` chama `POST /captures/:id/promote` — `RF-04`, `CA-10`, `00/RN-34`. Seam test. Depende da 03, 14 e 19.
- [ ] 19. Item `failed` mostra o `failure_reason` — `RF-05`, `CA-10`.
- [ ] 20. Atualização a cada 2 s enquanto houver item `uploaded` ou `promoting` — `RF-03`, `CA-07`. Depende da 01.
- [ ] 21. "Assistir" no item promovido com `file_id` — `RF-04`, `CA-08`.
- [ ] 21.1. "Ver na pasta" leva à pasta do vídeo promovido — `RF-04`.
- [ ] 22. Exclusão com retorno visível de sucesso (citando `library_file_kept`) e erro — `RN-11`, `CA-09`, `00/RN-33`, `CA-14`.
- [ ] 23. Busca por nome e filtro por estado na URL — `RF-02`, `00/RN-19 a RN-22`. Depende da 04 e 15.
- [ ] 24. Estado vazio explica o que são capturas e aponta para a extensão — `RF-06`.
- [ ] 25. Data relativa até 7 dias pelo formatador compartilhado — `00/RN-44`.
- [x] 26. Teste de render de `CapturesScreen` sem mock — regra do projeto, `00/RN-30`.
- [x] 27. Comentário de `service/captures.ts:48-50` removido.

### Plugin (toda lacuna sobe `version_name`)
- [x] 28. `uploadToKuraNAS` envia `metadata` e `episode_key` no `init` — `RN-16`, `CA-05`.
- [x] 29. Downloads de HLS, DASH, direto e blob usam os metadados detectados da aba — `RN-16`, `CA-05`. Depende da 28.
- [x] 30. Gravação híbrida sobrevive à queda de rede: reabre `init` com a mesma `episode_key`, realinha o offset e segue até um tempo máximo — `RN-01`, `RN-02`, `CA-04`.
- [x] 31. "Salvar" do popup só persiste o endereço depois de confirmar `/health` com "kuranas" — `RN-15`, `RF-12`, `CA-12`.
- [x] 32. Popup mostra a situação da conexão (conectado, não encontrado, `403`) — `RF-07`, `§9`. Depende da 31.
- [x] 33. "Baixar" e "Capturar" mostram o resultado no popup — `RF-11`. Depende da 29.
- [x] 34. `plugin/README.md` explica a limitação em conteúdo protegido — `§9`, `§4`.
- [x] 35. Comentários de corpo removidos de `plugin/`.

Ordem de execução: backend 01 → 13; frontend 14 → 27 (18 depois da 19 e 20); plugin 28 → 29 → 31 → 32 → 30 → 33 → 34 → 35. Backend e plugin são independentes; o frontend depende da 03 e 04.

### Decisões do dono
- 🚫 D1. Regravar episódio já capturado por completo: hoje regrava sozinho; `RN-02` diz que só por ação explícita.
- 🚫 D2. Captura "em envio" na tela: o registro só nasce no `complete`; mostrar progresso exige criar o registro no `init` ou expor as sessões.
- 🚫 D3. Expirar sessões abandonadas em `.uploads/` e capturas `failed` depois de N dias, ou manter.
- 🚫 D4. Filtro de plataforma: lista fixa, texto livre ou endpoint de plataformas distintas.
- 🚫 D5. Validação da captura no navegador real com o serviço do dono (`12 D-01`): quadro preto na gravação da aba levaria a captura ao nível do SO.
