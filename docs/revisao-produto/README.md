# Revisão de produto — board de execução

Revisão funcionalidade por funcionalidade (iniciada em 2026-10-08) com foco em **produto**: o que todo sistema daquele tipo tem, o que o KuraNAS se propõe a fazer (botões, colunas, rotas, funções, qualquer rastro de intenção) e o que ele realmente faz. Cada funcionalidade vira um arquivo `NN-<nome>.md` com a análise e a lista de lacunas; cada lacuna é resolvida uma por vez, em commits lógicos.

**Este README é a fonte de verdade do andamento.**

**O comportamento esperado de cada feature está em [`docs/especificacao/`](../especificacao/README.md).** Este board lista o trabalho; a spec diz o que o trabalho precisa atingir. Toda lacuna ou task nova cita a regra (`RN`/`RF`/`CA`) que satisfaz, e só é fechada com o critério verificado no sistema rodando.

**Ordem entre os boards** (2026-11-16): as tasks pendentes de [`docs/melhorias/`](../melhorias/README.md) (20 a 46) vêm antes dos itens deste board. O item 09 retoma depois delas.

## Protocolo

1. Pegue a primeira funcionalidade `pendente` (ou retome a `em execução`).
2. Análise completa (critérios básicos × intenção × realidade) → `NN-<nome>.md` com as lacunas numeradas e checkbox.
3. Resolva as lacunas **sequencialmente**, uma por vez; marque `[x]` no mesmo commit do código.
4. Lacuna que exige decisão do dono ou ambiente externo → `🚫` com motivo, segue para a próxima.
5. Fim da funcionalidade: `make ci` verde → `✅ concluída (data)`.

Regras invariantes: contrato HTTP só cresce (campos/params novos, nunca remover/renomear — frontend, 2 apps Android e plugin consomem a API); i18n obrigatório; sem comentários no código; teste de render sem mock em todo componente; seam test em toda mutação nova; commits lógicos em `develop`.

## Board

| # | Funcionalidade | Status | Notas |
|---|---|---|---|
| 01 | [Arquivos](01-arquivos.md) — navegação, CRUD, upload/download, detalhes, paginação, preview por tipo, localização física (tiering) | ✅ concluída (2026-10-08) | 26 lacunas resolvidas; `make ci` verde |
| 02 | [Shell e layout](02-shell-layout.md) — responsividade, sidebar, ocultação adaptativa, scroll | ✅ concluída (2026-10-08) | shell não montava desde 717b3f4; 24 lacunas; `make ci` verde |
| 03 | [Imagens](03-imagens.md) — galeria, paginação, busca por conteúdo, álbuns, timeline | ✅ concluída (2026-10-09) | 19 lacunas; HEIC/AVIF dependem de ffmpeg ≥ 7.1 no servidor; `make ci` verde |
| 04 | [Busca global](04-busca-global.md) | ✅ concluída (2026-10-09) | 13 lacunas; índices trigram/fold, conteúdo de documentos; `make ci` verde |
| 05 | [Música](05-musica.md) | ✅ concluída (2026-10-09) | 21 lacunas (sub-itens em commits próprios); transcodificação depende de ffmpeg; `make ci` verde |
| 06 | [Vídeos](06-videos.md) | ✅ concluída (2026-10-15) | 85 lacunas (sub-itens em commits próprios); 58.1 🚫 aguarda decisão de autoplay; `make ci` verde |
| 07 | [Favoritos](07-favoritos.md) | ✅ concluída (2026-11-10) | 31 lacunas (sub-itens em commits próprios); 24–26 🚫 aguardam decisão do dono; `make ci` verde |
| 08 | [Lixeira](08-lixeira.md) | ✅ concluída (2026-11-10) | 30 lacunas (web, backend e Android); 31–36 🚫 aguardam decisão do dono; Android validado só por build e testes unitários; `make ci` verde |
| 09 | [Downloads / ingestão por URL](09-ingestao-por-url.md) | em execução | parado atrás das tasks 20–46 de `docs/melhorias/`; lacunas 01–09 resolvidas, ordem das demais no fim do arquivo |
| 10 | [Capturas (plugin)](10-capturas.md) | em execução | em paralelo com o fim do 09 (pedido do dono 2026-11-25); 35 lacunas, D1–D5 🚫 |
| 11 | Takeout | pendente | |
| 12 | Analytics | pendente | |
| 13 | Notificações e jobs | pendente | |
| 14 | Diário de atividades | pendente | |
| 15 | Assistente | pendente | |
| 16 | Configurações — raízes, tiering, backup, watch folders, bibliotecas, acesso, desligamento, atualizador, IA | pendente | |
| 17 | E-mail | pendente | |
| 18 | Home | pendente | |
