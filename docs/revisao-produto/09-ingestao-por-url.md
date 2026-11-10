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
- Chaves `DOWNLOAD_FILE` e `DOWNLOADS_PAGE_EYEBROW` aparentam estar órfãs (a confirmar).

## Lacunas (uma alteração cada)

### Backend
- [x] 01. Cancelar o job interrompe o yt-dlp e descarta o parcial (teste com binário falso)
- [x] 02. Timeout do step `remote_fetch` (configurável, com padrão generoso)
- [ ] 03. Mover para o destino nunca sobrescreve: nome livre com sufixo ` (2)`, ` (3)`…
- [ ] 04. `scanner.Err()` tratado e parcial removido quando o fechamento da cópia falha
- [ ] 05. Temporário do download criado dentro da raiz de destino (pasta oculta ignorada por scan/watcher), para o mover ser atômico
- [ ] 06. Falha e conclusão do download geram notificação in-app com texto i18n
- [ ] 07. Erro do download guardado no job sem caminho do servidor
- [ ] 08. Subpasta inválida com mensagem própria
- [ ] 09. URL com limite de tamanho (400 com mensagem própria)
- [ ] 10. Atualização do yt-dlp: corpo limitado e uma atualização por vez
- [ ] 11. Teste do step de ponta a ponta com binário falso (sucesso, falha, sem arquivo)
- [ ] 12. Testes de `/ingest/ytdlp/*`, `/targets` e `/presets` fixam os nomes de campo da resposta
- [ ] 13. Comentários de corpo removidos de `ingest/` e do step; rota citada corrigida
- [ ] 14. Chaves de tradução órfãs removidas e paridade pt-BR × en-US conferida

### Frontend
- [ ] 15. `service/ingest.ts`: `fetchFromUrl`, `getIngestTargets`, `getIngestPresets` + tipos
- [ ] 16. Tela "Baixar de URL": URL, destino, subpasta e preset, com validação e erro do servidor (seam test do `POST /ingest/fetch`)
- [ ] 17. Entrada no menu e rota da tela
- [ ] 18. Downloads em andamento e recentes na tela, com progresso (depende de listar jobs por tipo — conferir `/jobs`)
- [ ] 19. Cancelar download em andamento (depende da 01)
- [ ] 20. Repetir um download que falhou
- [ ] 21. Render sem mock de `YtDlpSettingsSection` (conferir) e erro compartilhado no lugar do extrator local

### Plugin
- [ ] 22. Popup acompanha o progresso do download enviado (bump de `version_name`)

### Android
- [ ] 23. Tela "Baixar de URL" (destino, preset) e acompanhamento pelo `JobsApi` existente

### Decisões do dono
- 🚫 24. Bloquear URLs de hosts internos/loopback (SSRF) ou aceitar, dado que o acesso já é por whitelist de IP
- 🚫 25. Playlists: manter só o vídeo (`--no-playlist`) ou oferecer baixar a playlist inteira
- 🚫 26. Presets extras (720p, só áudio em outros formatos, legendas)
- 🚫 27. Teto de tamanho/duração por download e checagem de espaço livre antes de baixar
- 🚫 28. Reinício do servidor com download em andamento: recomeçar do zero (hoje) ou marcar como falho e avisar
