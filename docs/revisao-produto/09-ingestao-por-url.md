# 09 — Downloads / ingestão por URL

Escopo: `backend/internal/api/v1/ingest/` (fetch, targets, presets, ciclo de vida do yt-dlp), step `remote_fetch` em `backend/internal/worker/engine/step_remote_fetch.go`, `frontend/src/service/ingest.ts`, `frontend/src/components/settings/YtDlpSettingsSection.tsx`, popup do plugin (`plugin/src/background/fetcher.js`). Fora de escopo: `distribution` (`/downloads`) é o catálogo de apps cliente do próprio KuraNAS, não ingestão.

## Diagnóstico (2026-11-10)

- `POST /ingest/fetch` valida URL (http/https), preset (lista fechada: `audio_mp3`, `video_1080`, `video_best`) e destino (raiz habilitada + subpasta confinada) e enfileira um job `remote_fetch`; `GET /ingest/targets` e `/ingest/presets` alimentam o formulário. Sem tabela própria: tudo vive em `jobs`/`job_steps`.
- O step baixa com yt-dlp para um diretório temporário e move o resultado para o destino; a indexação fica com o watcher/scan da raiz.
- **Cancelar o job não interrompe o yt-dlp**: o processo roda com `context.Background()` e o cancelamento só alcança steps ainda na fila; o download termina e o arquivo entra na biblioteca mesmo assim.
- **Sem timeout**: um download travado ocupa o slot para sempre.
- **Sobrescreve em silêncio**: `moveFile` usa `os.Rename` sobre o destino; `--no-overwrites` só vale para o temporário, que nasce vazio. Dois downloads com o mesmo título, ou um arquivo já existente com o mesmo nome, perdem conteúdo.
- Falha só aparece em `last_error` do job e no log; ninguém é avisado. A mensagem é a última linha crua do stderr do yt-dlp (inglês, pode trazer caminho do servidor).
- `scanner.Err()` é ignorado (linha de saída > 1 MB interrompe a leitura sem erro); se `out.Close()` falha no fallback de cópia, o arquivo parcial fica no destino.
- Temporário em `os.TempDir()`: em outro filesystem o "mover" vira cópia não atômica e o watcher pode enxergar arquivo parcial.
- Subpasta inválida responde a mesma mensagem de "selecione uma biblioteca habilitada".
- URL sem limite de tamanho; nenhuma checagem de espaço livre.
- Atualizador do yt-dlp: checksum verificado e `.bak` ok; corpo do download sem limite, sem trava contra dois updates simultâneos, janela sem binário durante a troca.
- Comentários dentro de corpo de função em `ingest/` e no step; comentário de `dto.go` e de um teste citam a rota antiga `/downloads/fetch`.
- **Frontend web não tem a funcionalidade**: `service/ingest.ts` só cobre status/atualização do yt-dlp (seção em Configurações). Não há onde colar uma URL, escolher destino/preset, acompanhar progresso, cancelar, repetir ou ver histórico.
- Android não consome `/ingest`. O único cliente completo é o popup do plugin, que recebe o `job_id` e não acompanha o progresso.
- Chaves `DOWNLOAD_FILE` e `DOWNLOADS_PAGE_EYEBROW` estão órfãs (conferido em 2026-11-16).

## Lacunas (uma alteração cada)

### Backend
- [x] 01. Cancelar o job interrompe o yt-dlp e descarta o parcial (teste com binário falso)
- [x] 02. Timeout do step `remote_fetch` (configurável, com padrão generoso)
- [x] 03. Mover para o destino nunca sobrescreve: nome livre com sufixo ` (2)`, ` (3)`…
- [x] 04. Saída do yt-dlp além do buffer de progresso não trava mais o download (pipe drenado); parcial removido quando o fechamento da cópia falha
- [x] 05. Cópia entre filesystems grava em nome provisório (`.part`) e renomeia no fim: o nome final nunca aparece parcial (pasta temporária por raiz descartada — exigiria nova pasta ignorada em scan, watcher, backup e WebDAV)
- [x] 06. Falha e conclusão do download geram notificação in-app com texto i18n
- [x] 07. Erro do download guardado no job sem caminho do servidor
- [x] 08. Subpasta inválida com mensagem própria
- [x] 09. URL com limite de tamanho (400 com mensagem própria)
- [x] 10. Atualização do yt-dlp: corpo do binário limitado a 256 MB, listagem de checksums a 1 MB, e segunda atualização simultânea recusada com `409` e mensagem própria — `11/RN-13`, `CA-11`. Pronto quando: teste do serviço cobre binário acima do limite e duas chamadas concorrentes; teste do handler cobre o `409`.
- [x] 11. Teste do step `remote_fetch` de ponta a ponta com binário falso: sucesso (arquivo no destino), falha (motivo sem caminho do servidor) e saída sem arquivo — `11/RN-06`, `RN-07`, `RN-10`. Só teste; nenhum código de produção muda.
- [x] 12. Testes de `/ingest/ytdlp/status`, `/ingest/ytdlp/update`, `/ingest/targets` e `/ingest/presets` decodificam a resposta real do handler e fixam cada nome de campo — `00/CA-16`, `11 §7`.
- [x] 13. Comentários de corpo removidos de `backend/internal/api/v1/ingest/` e de `step_remote_fetch.go`; o comentário de `dto.go` e o do teste que citam `/downloads/fetch` passam a citar `/ingest/fetch`. Sem mudança de comportamento.
- [x] 14. Chaves `DOWNLOAD_FILE` e `DOWNLOADS_PAGE_EYEBROW` removidas de `pt-BR.json` e `en-US.json` (conferido em 2026-11-16: nenhum uso em `frontend/`, `backend/`, `android/` e `mobile/`); as chaves de download e do yt-dlp existem nos dois idiomas — `00/RN-43`.
- [x] 29. Registro de pedidos: migração da tabela `ingest_download` (colunas do item de `11 §7`, mais o caminho final interno, que nunca sai na API) e gravação do registro em `POST /ingest/fetch`, que passa a devolver também `download_id` — `11/RN-14`, `RN-16`. Seam: teste do handler decodifica `job_id` e `download_id`; consumidor a conferir: `plugin/src/background/fetcher.js` (campo novo, nada removido).
- [x] 30. O step `remote_fetch` mantém o registro: estado `running` ao começar; `progress`, `speed_bytes_per_second` e `eta_seconds` lidos da linha `[download]` do yt-dlp; `title` lido do nome do arquivo gerado; ao terminar, `completed`/`failed`/`canceled`, `ended_at`, caminho final e `error` traduzido — `11/RN-14`, `RN-15`, `RN-10`. Depende da 29.
- [x] 31. `GET /ingest/downloads`: lista paginada no envelope padrão, mais recente primeiro com desempate por `id`, filtro `state=active|finished` (`400` em valor inválido), `file_id` resolvido por `JOIN` do caminho final com a tabela de arquivos, destino como raiz + subpasta — `11/RN-14`, `RN-16`, `RF-04`, `RF-05`, `00/RN-17`, `RN-49` a `RN-52`. Um `.sql` próprio. Depende da 29.
- [x] 32. Ao iniciar o servidor, todo registro `queued` ou `running` vira `failed` com motivo traduzido "interrompido pelo reinício do servidor", e o job correspondente não recomeça — `11/RN-15`, `11 §8`. Depende da 29. *(comportamento provisório de D-05; a lacuna 28 continua aguardando o dono)*
- [x] 33. `POST /ingest/downloads/:id/retry`: cria pedido novo com o mesmo endereço, destino e formato; `201` com o item novo; `409` quando o pedido não está `failed` nem `canceled`; `404` quando não existe — `11/RN-17`, `CA-08`. Depende da 29.
- [x] 34. `POST /ingest/fetch` recusa com `503` e mensagem própria quando o yt-dlp não está instalado, em vez de enfileirar um job que vai falhar — `11/RN-12`, `CA-10`.
- [ ] 35. A notificação de conclusão leva ao arquivo baixado — `11/CA-07`. Depende da 30 e do destino explícito de notificação (`15 §8`, "a criar": tipo e identificador do alvo), que é lacuna do item 13 deste board; fica 🚫 até lá.

### Frontend
- [x] 15. `service/ingest.ts`: `fetchFromUrl`, `getIngestTargets`, `getIngestPresets`, `getIngestDownloads` (paginado, usando o adaptador de paginação compartilhado — `00/RN-53`) e `retryIngestDownload`, com tipos iguais ao item de `11 §7`. Depende da 31 e da 33.
- [x] 16. Tela "Baixar de URL" — formulário: endereço (com botão colar), destino (raiz + subpasta pelo seletor de pasta), formato; validação no campo antes de enviar (`http`/`https`, até 2 048 caracteres); erro do servidor exibido no formulário; ao enviar, o formulário limpa o endereço e fica pronto para o próximo — `11/RF-01` a `RF-03`, `CA-05`. Seam test do `POST /ingest/fetch`; teste de render sem mock.
- [x] 16.1. Último destino e último formato usados vêm preenchidos (guardados no dispositivo) — `11/RF-01`.
- [x] 16.2. Sem yt-dlp instalado (`GET /ingest/ytdlp/status`), a tela mostra o aviso com atalho para `/settings#ferramenta-de-download` e desabilita o envio — `11/RN-12`, `CA-10`.
- [x] 17. Rota `/ingest` e item "Baixar de URL" no grupo Ferramentas do menu — `01/RF-01`, `11/RF-01`.
- [x] 18. Bloco "Em andamento" (`state=active`): título ou endereço, destino, formato, barra com porcentagem, velocidade e tempo restante; consulta a cada 2 s só enquanto há item ativo — `11/RF-04`, `CA-01`. Depende da 15.
- [x] 18.1. Bloco "Histórico" (`state=finished`): scroll infinito conforme `00`, estado, data e destino; "Abrir" (habilitado com `file_id`) e "Ver na pasta"; motivo nos falhos — `11/RF-05`, `00/CA-01` a `CA-12`, `CA-16`. Depende da 15.
- [x] 19. "Cancelar" no item em andamento chama `POST /jobs/:id/cancel`, desabilita só aquele item durante a chamada e o item passa ao histórico como cancelado — `11/RN-08`, `CA-03`, `00/RN-34`. Seam test.
- [x] 20. "Tentar de novo" nos itens falhos e cancelados chama `POST /ingest/downloads/:id/retry`; o item novo aparece em "Em andamento" — `11/RN-17`, `CA-08`. Seam test. Depende da 33.
- [x] 21. `YtDlpSettingsSection`: teste de render sem mock (conferido em 2026-11-16: os dois testes existentes mockam `useYtDlpSettings`, `@/service` ou o i18n — nenhum renderiza sem mock) — regra do projeto, `01/RN-02`.
- [x] 21.1. `YtDlpSettingsSection` mostra a mensagem do servidor na falha de carregar e na falha de "Atualizar" (hoje usa o texto fixo `SETTINGS_YTDLP_LOAD_ERROR`), pelo componente de erro compartilhado — `00/RN-41`, `11/RF-06`.

### Plugin
- [x] 22. Depois de enviar, o popup consulta `GET /jobs/:id` a cada 2 s e mostra estado e porcentagem até o job terminar; bump de `version_name` — `11/RF-07`, `CA-12`.

### Android
- [x] 23. Tela "Baixar de URL" com endereço, destino e formato, e a lista de pedidos por `GET /ingest/downloads` (paginada, com cancelar e tentar de novo); entrada no menu "Mais" — `11/RF-08`, `24/RF-03`. Depende da 31 e da 33.
- [x] 23.1. O app aparece em "Compartilhar" de outros aplicativos para texto com endereço (`ACTION_SEND`, `text/plain`) e abre a tela com o endereço preenchido — `11/RF-08`.

Ordem de execução: 10 → 14 (backend já aberto), 29 → 34, depois frontend 15 → 21, plugin 22, Android 23 → 23.1. A 35 espera o item 13 do board.

### Decisões do dono
- 🚫 24. Bloquear URLs de hosts internos/loopback (SSRF) ou aceitar, dado que o acesso já é por whitelist de IP
- 🚫 25. Playlists: manter só o vídeo (`--no-playlist`) ou oferecer baixar a playlist inteira
- 🚫 26. Presets extras (720p, só áudio em outros formatos, legendas)
- 🚫 27. Teto de tamanho/duração por download e checagem de espaço livre antes de baixar
- 🚫 28. Reinício do servidor com download em andamento: recomeçar do zero (hoje) ou marcar como falho e avisar
