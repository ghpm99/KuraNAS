# Melhorias do sistema — board de execução

Backlog de melhorias levantado na análise de maturidade do sistema (2026-06-11), estendido com a visão de armazenamento do dono (tiering + backup, 2026-06-11) e com a demanda e-mail + kiosk (2026-06-12). Cada arquivo é **uma task autocontida**: contexto, objetivo, o que/como fazer, critérios de aceite e fora de escopo.

**Este README é a fonte de verdade do andamento.** Um agente apontado para este arquivo deve conseguir continuar o trabalho de onde parou, mesmo após interrupção — o protocolo abaixo existe para isso.

**O comportamento esperado de cada feature está em [`docs/especificacao/`](../especificacao/README.md).** Este board lista o trabalho; a spec diz o que o trabalho precisa atingir. Toda lacuna ou task nova cita a regra (`RN`/`RF`/`CA`) que satisfaz, e só é fechada com o critério verificado no sistema rodando.

## Protocolo de execução (para agentes)

1. **Leia o board.** Se existe task `em execução`, **retome-a**: abra o arquivo da task, veja quais critérios de aceite já estão marcados (`[x]`), confira no `git log` o que já foi commitado para ela, e continue do primeiro critério desmarcado. Não recomece do zero o que os checkboxes dizem que está feito — mas verifique que o último commit compila (`make ci`) antes de confiar.
2. **Se não há task em execução**, pegue a primeira `pendente` de cima para baixo cuja coluna **Depende de** esteja toda `concluída`.
3. **Ao iniciar**: mude o status dela para `em execução` neste README e commite essa mudança isolada (`docs(melhorias): inicia task NN`). É esse commit que marca o estado para uma retomada futura.
4. **Durante**: trabalhe em commits lógicos (workflow do projeto). Ao satisfazer um critério de aceite, **marque o checkbox `[x]` no arquivo da task no mesmo commit do código** que o satisfaz — os checkboxes são o progresso fino que permite retomar no meio.
5. **Ao concluir**: todos os checkboxes marcados + `make ci` verde → status `✅ concluída (data)` na tabela + commit (`docs(melhorias): conclui task NN`).
6. **Se travar** (decisão pendente do dono, dependência externa): status `🚫 bloqueada` com o motivo na coluna Notas, commite, e passe para a próxima `pendente` elegível.

**Ordem entre os dois boards** (decidida em 2026-11-16): as tasks `pendente` deste board, de cima para baixo, vêm antes de qualquer item de [`docs/revisao-produto/`](../revisao-produto/README.md). As tasks 20 a 33 corrigem defeitos que quebram o uso básico; as 34 a 46 entregam o que falta das necessidades N1–N7. O item 09 do board de revisão (download por URL, P2), que estava em execução, fica parado até a task 39 terminar — as mudanças já iniciadas da lacuna 10 podem ser concluídas e commitadas antes de começar a 20, para não deixar trabalho pela metade.

**Antes de escrever código de uma task**, confira no código os arquivos e linhas citados no Contexto dela (foram levantados por leitura em 2026-11-16 e o código pode ter andado). Se algo divergir, corrija o arquivo da task primeiro, em commit próprio.

Regras invariantes:

- **Uma task `em execução` por vez.** Não pular a ordem sem registrar o motivo em Notas.
- Toda task termina **verde no `make ci`**; o **contrato HTTP não muda** salvo a task dizer o contrário (frontend, 2 apps Android e plugin consomem a API); **i18n obrigatório** em toda string visível; commits lógicos diretos em `develop`, sem `Co-Authored-By`.
- Mudança de status é sempre um commit — o board nunca fica só na memória de quem executa.

## Board

| # | Task | Tipo | Prioridade | Depende de | Status | Notas |
|---|---|---|---|---|---|---|
| 01 | [Indexação de diretórios](01-indexacao-de-diretorios.md) | bug crítico | P0 | — | ✅ concluída (2026-06-11) | causa raiz do bug reportado |
| 02 | [Filtro deleted_at](02-filtro-deleted-at.md) | bug | P0 | — | ✅ concluída (2026-06-11) | |
| 03 | [Debounce do watcher perde eventos](03-debounce-watcher-perde-eventos.md) | bug | P1 | — | ✅ concluída (2026-06-11) | |
| 04 | [Whitelist de IPs](04-whitelist-de-ips.md) | segurança | P0 | — | ✅ concluída (2026-06-11) | decisão: sem autenticação |
| 05 | [Operações atualizam banco sincronamente](05-operacoes-atualizam-banco-sincronamente.md) | consistência | P1 | 01, 02 | ✅ concluída (2026-06-11) | |
| 06 | [Watcher por eventos (fsnotify)](06-watcher-por-eventos-fsnotify.md) | performance | P2 | 01, 03 | 🚫 bloqueada | código pronto e CI verde; falta só validação manual no Windows (dono) |
| 07 | [Remover pipeline legado](07-remover-pipeline-legado.md) | dívida técnica | P2 | 01 | ✅ concluída (2026-06-11) | |
| 08 | [Decompor get_files query](08-decompor-get-files-query.md) | dívida técnica | P2 | 02 | ✅ concluída (2026-06-11) | |
| 09 | [Lixeira](09-lixeira.md) | feature | P2 | 02, 05 | ✅ concluída (2026-06-12) | |
| 10 | [Múltiplas raízes](10-multiplas-raizes.md) | feature | P3 | 01, 05 | ✅ concluída (2026-06-12) | fundação da visão de armazenamento |
| 11 | [Acesso WebDAV](11-acesso-webdav.md) | feature | P3 | 04 | 🚫 bloqueada | código pronto e CI verde; falta validação manual do dono (montar via Explorer/davfs2, `WEBDAV_ENABLED=true`) |
| 14 | [Contas de e-mail + OAuth2](14-email-contas-oauth.md) | feature | P2 | 04 | 🚫 bloqueada | código pronto e CI verde; falta o dono registrar os apps OAuth (Google Cloud + Entra), configurar `EMAIL_TOKEN_KEY`/`EMAIL_*` e vincular as contas reais pela UI |
| 15 | [Sincronização de e-mail](15-email-sync-worker.md) | feature | P2 | 14 | ✅ concluída (2026-06-13) | `pkg/mailfetch` (Gmail/Graph, allowlist) + sanitizador + pré-filtro + job `email_sync`; migração saiu como **0040** (0039 já usada pela task 13) e ganhou coluna `prefilter_rules`; `make ci` verde |
| 16 | [Análise de e-mail por IA](16-email-analise-ia.md) | feature | P2 | 15 | ✅ concluída (2026-06-13) | step `email_analyze` (classificação fail-closed + resumo), migração **0041** `email_analysis`, `Manager.Named` com hot-swap, provedor selecionável na UI (default Ollama, aviso de privacidade); `make ci` verde |
| 17 | [Enxugar app legado](17-app-legado-limpeza.md) | dívida técnica | P2 | — | 🚫 bloqueada | código pronto, build e testes verdes, APK −4,7%; falta só instalar/rodar no tablet (dono) |
| 18 | [Tela kiosk do app legado](18-app-legado-kiosk.md) | feature | P2 | 16, 17 | 🚫 bloqueada | código pronto, `assembleDebug`+`testDebugUnitTest` verdes; faltam só os 2 critérios do ambiente do dono — medir payload < 4 KB no servidor real e rodar no Galaxy Tab 2 |
| 12 | [Backup orquestrado](12-backup-orquestrado.md) | feature | P3 | 10 | ✅ concluída (2026-06-12) | retenção ≠ espelho; segunda cópia (HD externo) documentada no README, fica com o SO |
| 13 | [Tiering quente/frio](13-tiering-quente-frio.md) | feature | P3 | 01, 05, 10 | ✅ concluída (2026-06-13) | path lógico × físico; job tier_migration + operações tiered + UI; `make ci` verde |
| 19 | [Captura inteligente de streaming](19-captura-inteligente-de-streaming.md) | feature | P3 | — | 🚫 bloqueada | código pronto, `npm test` (plugin) + `make ci-backend` verdes. Falta validação do dono no navegador real: (1) confirmar que `tabCapture` na cobaia (Crunchyroll) **não** entrega frame preto — se entregar, a peça de captura migra para nível de SO (analog hole continua, ver Riscos); (2) verificar in-loco arma-em-fullscreen / finaliza-no-fim / autoplay→arquivo-novo / retomada→arquivo-único. Máquina de estados + idempotência por `episode_key` + adapter Crunchyroll já cobertos por testes; **sem decrypt de DRM** |
| 20 | [Web: segunda página das listagens de arquivos](20-paginacao-web-le-has-next.md) | bug crítico | P0 | — | 🚫 bloqueada | código pronto e `make ci` verde (2026-11-16); faltam os CAs de tela no sistema rodando (dono) — pasta de teste `teste-paginacao-task20` com 760 itens, 200 favoritos e 200 recentes no ambiente de dev. spec 00 anexo A1–A3 |
| 21 | [Web: falha ao carregar mais não esvazia a lista](21-falha-de-pagina-nao-esvazia-a-lista.md) | bug crítico | P0 | 20 | 🚫 bloqueada | código pronto e `make ci` verde (2026-11-16); faltam os CAs de tela no sistema rodando (dono) — derrubar o servidor entre páginas em cada tela. spec 00 anexo A4 |
| 22 | [Android: encontrar o servidor](22-android-descoberta-do-servidor.md) | bug crítico | P0 | — | 🚫 bloqueada | código pronto, `testDebugUnitTest` (205) e `assembleDebug` verdes (2026-11-16); faltam os CAs no aparelho real na rede do servidor (dono): achar em 10 s, busca vazia termina, endereço errado recusado, reabrir conecta, IP trocado reencontrado, mDNS bloqueado acha por UDP. spec 24 |
| 23 | [Android: explorer paginado](23-android-explorer-paginado.md) | bug crítico | P0 | — | 🚫 bloqueada | código pronto, `testDebugUnitTest` (227) e `assembleDebug` verdes (2026-11-16); sub-item 23.1 levou o rodapé de erro às outras listas paginadas; faltam os CAs no aparelho (dono): pasta de 300 itens e raiz com mais de 50 rolam até o fim, servidor derrubado após a 1ª página, girar mantém itens e posição, operações atualizam a lista. | spec 24, 00 anexo A10 |
| 24 | [Android: enviar arquivos e Compartilhar](24-android-enviar-arquivos.md) | feature | P0 | 22, 23 | pendente | spec 03 |
| 25 | [Capturas: excluir não apaga o vídeo](25-excluir-captura-nao-apaga-video.md) | bug crítico | P0 | — | 🚫 bloqueada | código pronto e `make ci` verde (2026-11-16); faltam os CAs no sistema rodando (dono): excluir captura promovida mantém o vídeo em Vídeos e no disco e a confirmação diz isso; excluir captura que falhou na promoção apaga o arquivo recebido. | spec 12/RN-11 |
| 26 | [Takeout: estrutura e sem sobrescrita](26-takeout-sem-sobrescrita.md) | bug crítico | P0 | — | 🚫 bloqueada | código pronto e `make ci` verde (2026-11-16); faltam `13/CA-03`, `CA-04`, `CA-09` no sistema rodando com um zip real de Takeout (dono). | spec 13 |
| 27 | [Takeout: aplicar datas](27-takeout-aplica-datas.md) | bug | P0 | 26 | pendente | spec 13/RN-07 |
| 28 | [Takeout: extração em fluxo](28-takeout-extracao-em-fluxo.md) | bug | P1 | 26 | pendente | spec 13/RN-09 |
| 29 | [Indexação: preferências com efeito](29-preferencias-de-indexacao-com-efeito.md) | bug | P1 | — | 🚫 bloqueada | código pronto e `make ci` verde (2026-11-17); falta `21/CA-12` no sistema rodando (dono): desligar "gerar miniaturas" faz um arquivo novo aparecer sem miniatura e religar volta a gerar. Spec 21 §8 ganhou a exceção das filas de pendência; `WATCHER_RECONCILE_HOURS` removida. | spec 21/RN-14 |
| 30 | [Reindexar por escopo](30-reindexar-por-escopo.md) | bug | P1 | — | 🚫 bloqueada | código pronto e `make ci` verde (2026-11-17), inclusive o 30.1 (reindexar recupera só as etapas que faltam — decisão do dono); faltam `21/CA-11`, `16/CA-09` e `21/CA-08` no sistema rodando (dono). | spec 21/RN-15, 16/RN-07 |
| 31 | [Analytics: links levam à lista certa](31-analytics-links-levam-a-lista-certa.md) | bug | P2 | 20, 21 | pendente | spec 16/RN-06 |
| 32 | [Atualização confere checksum](32-atualizacao-confere-checksum.md) | segurança | P1 | — | 🚫 bloqueada | código pronto e `make ci` verde (2026-11-17); faltam `23/CA-11` (pacote adulterado recusado no sistema rodando) e publicar o próximo release com `SHA256SUMS` e atualizar a partir dele (dono). Atenção: servidor com este código recusa atualizar para release sem `SHA256SUMS`. | spec 23/RN-15 |
| 33 | [Desligamento: aviso e cancelar](33-cancelar-desligamento.md) | bug | P1 | — | 🚫 bloqueada | código pronto e `make ci` verde (2026-11-17); sub-item 33.1 (aviso 30 a 600 s) feito; faltam `23/CA-08` e a faixa num segundo dispositivo no sistema rodando (dono). | spec 23/RN-11 |
| 34 | [Envio em partes: backend](34-envio-em-partes-backend.md) | feature | P0 | — | ✅ concluída (2026-11-17) | spec 03 §7 |
| 35 | [Envio em partes: web](35-envio-em-partes-web.md) | feature | P0 | 34 | em execução | spec 03 |
| 36 | [Android: download em fluxo](36-android-download-em-fluxo.md) | bug | P0 | 22 | pendente | spec 03 |
| 37 | [Android: transferências em segundo plano](37-android-transferencias-em-segundo-plano.md) | feature | P0 | 24, 34, 36 | pendente | spec 03 |
| 38 | [Android: envio automático da câmera](38-android-envio-automatico-da-camera.md) | feature | P0 | 37 | pendente | spec 03 |
| 39 | [Android: navegação](39-android-navegacao.md) | bug | P1 | 22 | pendente | spec 24 |
| 40 | [Música: corrigir metadados](40-musica-edicao-de-metadados.md) | feature | P0 | — | pendente | spec 05 |
| 41 | [Música: playlist por regra](41-musica-playlist-por-regra.md) | feature | P0 | 40 | pendente | spec 05 |
| 42 | [Vídeos: tela da série](42-videos-tela-da-serie.md) | feature | P0 | — | pendente | spec 06 |
| 43 | [Vídeos: faixa de áudio](43-videos-faixa-de-audio.md) | feature | P1 | — | pendente | spec 06 |
| 44 | [Documentos: biblioteca no backend](44-documentos-biblioteca-backend.md) | feature | P0 | — | pendente | spec 07 |
| 45 | [Documentos: tela na web](45-documentos-tela-web.md) | feature | P0 | 44, 20, 21 | pendente | spec 07 |
| 46 | [Documentos: tela no Android](46-documentos-android.md) | feature | P1 | 44, 36, 39 | pendente | spec 07 |

Status possíveis: `pendente` · `em execução` · `✅ concluída (AAAA-MM-DD)` · `🚫 bloqueada`.

## Decisões registradas (valem para todas as tasks)

- **Sem autenticação** (2026-06-11): nada de login/senha/token. Controle de acesso é whitelist de IPs (task 04). TLS fora de escopo enquanto o produto for de rede interna.
- **Redundância de disco é do SO, não do app** (2026-06-11): HDs avulsos em pool espelhado via Windows Storage Spaces; o KuraNAS enxerga um volume comum. Detalhe na task 10.
- **Visão de armazenamento** (2026-06-11): SSD = tier quente; pool de HDs = tier frio + backup com retenção; HD externo 2 TB = segunda cópia desconectável (gerida pelo SO). Tasks 10 → 12/13.
- **Backup ≠ espelho**: backup tem retenção de versões; espelho propaga ransomware/exclusão acidental (task 12).
- **Tiering é transparente**: arquivo migrado não muda de lugar na árvore lógica — separação path lógico × localização física (task 13).
- **Regras duras de e-mail** (2026-06-12, valem para as tasks 14–18 e qualquer evolução futura — viabilidade da feature depende delas): (1) escopos OAuth somente leitura (`gmail.readonly`, `Mail.Read` + `offline_access`), nenhuma capacidade de envio jamais; (2) nunca buscar URL contida em e-mail, nunca baixar/armazenar/executar anexo (metadados apenas); (3) HTML → texto puro no backend antes de qualquer LLM, remoção de Unicode invisível, corpo ≤ 16 KB; (4) o LLM do pipeline de e-mail não tem ferramentas — entrada texto, saída JSON com schema validado; parse inválido = `suspicious` (fail-closed); (5) tokens cifrados em repouso (AES-GCM, chave `EMAIL_TOKEN_KEY` em env; sem a chave a feature não liga); (6) clients HTTP de e-mail com allowlist fixa de hosts; (7) spam barrado no pré-filtro determinístico não chega ao LLM. Reputação externa de URLs ficou fora do v1 (cria superfície e vaza dados). Pior caso aceito: e-mail mal classificado/resumo errado no painel — nunca execução, download ou exfiltração.
- **OAuth dos e-mails** (2026-06-12): Microsoft pessoal via Device Code Flow (public client, audience consumers, só `EMAIL_MS_CLIENT_ID`); Google via Authorization Code + PKCE com loopback `localhost:8000` (Device Flow do Google não aceita escopos Gmail) — vínculo feito em navegador na máquina do NAS ou via túnel SSH; consent screen publicada In production para o refresh token não expirar em 7 dias. Detalhe na task 14.
- **Privacidade da análise de e-mail** (2026-06-12): provedor de IA escolhível na UI (chave `email_ai_provider` na `configuration`), default Ollama local; nuvem só por escolha explícita com aviso de privacidade (task 16).
- **App legado é painel de parede** (2026-06-12): o `mobile/` perde todas as telas de navegação de mídia (removidas de vez, task 17) e vira discovery + kiosk (task 18); navegação fica com o app `android/` novo.
- **Ingestão de mídia: slots verdes apenas, sem quebra de DRM** (2026-06-15, vale para as tasks de ingestão — Fase 1 já em `develop`, Fase 2 na task 19): o produto puxa por URL fontes sem DRM (`yt-dlp` no servidor, domínio `ingest`, montado em `/ingest` porque `/downloads` é da `distribution`) e, para conteúdo com DRM, **só captura a saída renderizada** (analog hole — grava o frame já decodificado), nunca decripta o stream. **Nenhum** código de decriptação de DRM (extração de chave Widevine/CDM, `mp4decrypt` com chave) entra no repositório — **nem um ponto de extensão** para recebê-lo. Razão: o repo é público; distribuir tecnologia de circumvenção é o ato de tráfico da DMCA §1201(a)(2) e da Lei 9.610/98 Art. 107, independente do uso pessoal ser arquivar obras que o dono paga e não compartilha. Atualização do `yt-dlp` é manual e verificada por checksum (sem auto-update — cautela de supply-chain).
