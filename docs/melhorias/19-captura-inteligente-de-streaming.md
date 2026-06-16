# 19 — Captura inteligente de streaming (Fase 2 da ingestão de mídia)

**Tipo:** feature (ingestão de mídia) · **Prioridade:** P3 · **Depende de:** — (a Fase 1, o domínio `ingest` + `/ingest/fetch` via yt-dlp, já está em `develop`; ver `project-plugin-media-ingest` e os commits `cdcc33c`/`ddb3742`)

## Contexto

O dono assina várias plataformas (Netflix, Prime, Crunchyroll, Mercado Play…) mas títulos sai do catálogo ou migram, e ele quer arquivar localmente — **só para rever, sem compartilhar** — obras que paga para assistir. A **Fase 1** resolve as fontes **sem DRM** (YouTube etc.): o servidor puxa por URL com `yt-dlp` direto na biblioteca. Esta Fase 2 é para o conteúdo **com DRM (Widevine)**, onde puxar por URL não existe: o único caminho é **capturar a saída renderizada** (o "analog hole" — gravar o frame que o player já decodificou e mostrou na tela).

O plugin já faz uma versão crua disso: o modo híbrido (`hybrid-state.js` + `offscreen/recorder.js`) usa `tabCapture` → `MediaRecorder` → upload em chunks (`/captures/upload/*`). Os problemas hoje: exige dar play manualmente com delay para não gravar os controles, grava a aba inteira (UI inclusa), não conhece fronteira de episódio e, se interrompido, gera arquivos duplicados.

A ideia do dono: um gravador **inteligente** que siga o player sozinho — se ele dormir vendo o episódio, para quando o episódio acaba e volta quando o próximo começa; e se a gravação for interrompida no mesmo episódio, não gera dois arquivos.

## Restrição dura (decisão registrada)

**Slots verdes apenas. Nenhuma quebra de DRM em lugar nenhum do repositório — nem um ponto de extensão.** Esta task é só **captura de saída** (gravar o frame já decodificado), que **não é** circumvenção de DRM. Extração de chave Widevine/CDM, `mp4decrypt` com chave, qualquer pipeline de decriptação: **fora**, e fora também de um gancho/interface que os receba — publicar isso num repo público é o ato de tráfico que a DMCA §1201(a)(2) e a Lei 9.610/98 Art. 107 descrevem, independente do uso pessoal. Ver "Decisões registradas" no README.

## Objetivo

Assistindo ou não, o gravador segue o player: **para** quando o episódio termina (sobrevive ao dono dormir), **começa um arquivo novo** quando o autoplay engata o próximo, e um episódio **interrompido e retomado** continua o **mesmo** arquivo em vez de duplicar. Os frames saem limpos (sem os controles do player).

## O que fazer

1. **Adapter de página por serviço** (content script) que normaliza o estado do player num contrato único.
2. **Máquina de estados de gravação chaveada por episódio** no service worker, dirigida pelos eventos do adapter.
3. **Sessão de upload idempotente por episódio** no backend, para retomar/anexar em vez de duplicar.
4. **Captura limpa**: gravar só em tela cheia.

Entregar com **um serviço cobaia** (sugestão: Crunchyroll, citado pelo dono) — a arquitetura é genérica, mas só um adapter concreto nesta task.

## Como fazer

- **Adapter de página** (`content/adapters/<servico>.js`, MAIN world, no padrão de `content/title-detector.js`): lê o `<video>` e o DOM e emite, para a `bridge.js`, um estado normalizado:
  - `isPlaying`, `currentTime`, `duration` (do elemento `<video>`);
  - `episodeId` **estável** extraído da URL/DOM (série + temporada + episódio, ou o content-id do serviço) — é a chave de tudo;
  - `title` legível (para o nome do arquivo);
  - `isFullscreen`.
  Eventos derivados: `episodeStarted(episodeId)`, `episodeEnded(episodeId)`. Um **registry** mapeia hostname → adapter; host sem adapter não arma nada (degrada para o modo manual atual). O adapter é a única peça específica de site — é o que vai quebrar quando o site mudar, então fica isolado e pequeno.
- **Máquina de estados** (`src/background/capture-session.js`, no espírito de `hybrid-state.js`), chaveada por `episodeId`:
  - `IDLE → RECORDING(id)` quando `isPlaying && isFullscreen` e há `episodeId`;
  - `episodeEnded` **ou** `currentTime` cruzando perto de `duration` → finaliza e volta a `IDLE` (**dono dormiu → episódio acaba → para**);
  - `episodeStarted(novoId)` com `novoId ≠ atual` → finaliza o atual e começa o novo (**autoplay → arquivo novo**);
  - reaparição do **mesmo** `episodeId` (retomada após interrupção) → continua a **mesma** captura lógica (ver idempotência abaixo), não abre arquivo novo.
  - Pausa curta não finaliza (janela de graça, como o `HYBRID_STOP_GRACE_MS` atual); só `episodeEnded`/fim real encerra.
- **Captura limpa por fullscreen**: o gatilho de gravação inclui `isFullscreen` — em tela cheia o player ocupa todo o frame e os controles somem após ociosidade, então a saída sai sem chrome. Sair de fullscreen no meio entra na janela de graça (não corta na hora).
- **Backend — sessão idempotente por episódio**: `/captures/upload/init` ganha um campo opcional `episode_key` (string estável vinda do adapter: `<servico>:<episodeId>`). Comportamento:
  - se já existe sessão **incompleta** com aquele `episode_key`, `init` **retoma** essa sessão (devolve o mesmo `upload_id` e o offset já recebido) em vez de criar outra;
  - se já existe captura **completa** com aquele `episode_key`, `init` responde "já existe" e o cliente não regrava.
  Mantém o contrato atual retrocompatível (sem `episode_key` = comportamento de hoje). Migração: coluna `episode_key` em `captures` (e/ou na tabela de sessão de upload, conforme onde o estado de upload mora hoje) + índice único parcial por `episode_key` enquanto a sessão está aberta.
- **Nome/ível do arquivo**: derivado do `title` + `episodeId` do adapter (sanitizado), para cair organizado na biblioteca; reaproveitar `sanitizeFileName`.
- **i18n**: chaves novas no catálogo (estado "gravando episódio X", "retomando", "episódio já arquivado", erros) — plugin ainda fora da regra de i18n obrigatória, então literais no popup seguem o padrão atual; mensagens vindas do backend exibidas como chegam.
- **Testes**: máquina de estados isolada e testável (transições IDLE↔RECORDING, dormir = episodeEnded, autoplay = novo id, retomada = mesma sessão) com `node --test`, como `tests/hybrid-state.test.js`; adapter com fixtures de DOM/estado; backend: `init` idempotente (retoma sessão aberta, recusa duplicado completo) nos testes de `captures`.

## Critérios de aceite

- [ ] Com o vídeo em tela cheia tocando, a gravação **inicia sozinha**; sem fullscreen, não grava (frames limpos, sem controles).
- [ ] Episódio chega ao fim (ou o dono dorme e o player para no fim) → a gravação **finaliza** e volta a ocioso, **sem** cortar conteúdo antes do fim.
- [ ] Autoplay do próximo episódio → **arquivo novo** separado, sem ação manual.
- [ ] Gravação interrompida e retomada **no mesmo episódio** → **um único arquivo** (sessão idempotente por `episode_key`), nunca dois.
- [ ] Episódio já arquivado por completo não é regravado.
- [ ] Host sem adapter não arma a captura inteligente (degrada para o modo manual atual sem erro).
- [x] Contrato `/captures/upload/*` retrocompatível: requisições sem `episode_key` funcionam como hoje.
- [ ] **Nenhum** código de decriptação de DRM no repo (nem ponto de extensão) — revisão explícita.
- [ ] Testes verdes: `npm test` (plugin) e `make ci` (backend).

## Riscos / notas de realidade

- **Frame preto por DRM**: `tabCapture` de vídeo Widevine frequentemente entrega **frame preto** (caminho de vídeo protegido). Onde a captura atual funciona, é DRM por software (L3). Se a cobaia escolhida entregar preto, o caminho confiável passa a ser **captura no nível do SO** (gravar a saída composta da tela, ainda analog hole, ainda slot verde) em vez de `tabCapture` — isso muda a peça de captura, não a máquina de estados nem a idempotência (que continuam valendo). Avaliar na cobaia antes de fechar a abordagem de captura.
- **Fragilidade do adapter**: cada serviço muda seu DOM; manter o adapter mínimo e isolado, com o estado normalizado bem definido, para que só ele precise de manutenção.

## Fora de escopo

- **Qualquer decriptação de DRM** (extração de chave Widevine/CDM, `mp4decrypt` com chave, etc.) e qualquer interface/gancho que a receba — linha dura do projeto.
- Adapters para mais de um serviço além da cobaia (cada um vira incremento próprio depois).
- Captura no nível do SO como implementação (só avaliada como risco aqui; se necessária, vira task própria por mexer fora do navegador).
- Edição pós-gravação (cortar abertura/encerramento, trim) — evolução futura.
