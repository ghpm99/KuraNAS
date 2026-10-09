# Revisão de produto — board de execução

Revisão funcionalidade por funcionalidade (iniciada em 2026-10-08) com foco em **produto**: o que todo sistema daquele tipo tem, o que o KuraNAS se propõe a fazer (botões, colunas, rotas, funções, qualquer rastro de intenção) e o que ele realmente faz. Cada funcionalidade vira um arquivo `NN-<nome>.md` com a análise e a lista de lacunas; cada lacuna é resolvida uma por vez, em commits lógicos.

**Este README é a fonte de verdade do andamento.**

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
| 05 | [Música](05-musica.md) | em execução | |
| 06 | Vídeos | pendente | |
| 07 | Favoritos | pendente | |
| 08 | Lixeira | pendente | |
| 09 | Downloads / ingestão por URL | pendente | |
| 10 | Capturas (plugin) | pendente | |
| 11 | Takeout | pendente | |
| 12 | Analytics | pendente | |
| 13 | Notificações e jobs | pendente | |
| 14 | Diário de atividades | pendente | |
| 15 | Assistente | pendente | |
| 16 | Configurações — raízes, tiering, backup, watch folders, bibliotecas, acesso, desligamento, atualizador, IA | pendente | |
| 17 | E-mail | pendente | |
| 18 | Home | pendente | |
