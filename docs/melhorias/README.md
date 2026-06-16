# Melhorias do sistema — board de execução

Backlog de melhorias levantado na análise de maturidade do sistema (2026-06-11), estendido com a visão de armazenamento do dono (tiering + backup, 2026-06-11) e com a demanda e-mail + kiosk (2026-06-12). Cada arquivo é **uma task autocontida**: contexto, objetivo, o que/como fazer, critérios de aceite e fora de escopo.

**Este README é a fonte de verdade do andamento.** Um agente apontado para este arquivo deve conseguir continuar o trabalho de onde parou, mesmo após interrupção — o protocolo abaixo existe para isso.

## Protocolo de execução (para agentes)

1. **Leia o board.** Se existe task `em execução`, **retome-a**: abra o arquivo da task, veja quais critérios de aceite já estão marcados (`[x]`), confira no `git log` o que já foi commitado para ela, e continue do primeiro critério desmarcado. Não recomece do zero o que os checkboxes dizem que está feito — mas verifique que o último commit compila (`make ci`) antes de confiar.
2. **Se não há task em execução**, pegue a primeira `pendente` de cima para baixo cuja coluna **Depende de** esteja toda `concluída`.
3. **Ao iniciar**: mude o status dela para `em execução` neste README e commite essa mudança isolada (`docs(melhorias): inicia task NN`). É esse commit que marca o estado para uma retomada futura.
4. **Durante**: trabalhe em commits lógicos (workflow do projeto). Ao satisfazer um critério de aceite, **marque o checkbox `[x]` no arquivo da task no mesmo commit do código** que o satisfaz — os checkboxes são o progresso fino que permite retomar no meio.
5. **Ao concluir**: todos os checkboxes marcados + `make ci` verde → status `✅ concluída (data)` na tabela + commit (`docs(melhorias): conclui task NN`).
6. **Se travar** (decisão pendente do dono, dependência externa): status `🚫 bloqueada` com o motivo na coluna Notas, commite, e passe para a próxima `pendente` elegível.

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
