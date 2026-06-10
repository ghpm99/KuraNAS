# Refatoração de organização do backend

Plano de reorganização **estrutural** do backend (`backend/`) para torná-lo previsível e navegável: cada pasta = um domínio, com conteúdo esperável. Não muda comportamento — só onde o código mora.

> A **regra** resultante é canônica e vive nos `CLAUDE.md` (raiz → "Backend domains: a generic file core + type extensions"; `backend/CLAUDE.md` → "Domain package organization"). Estes docs descrevem o **caminho** até lá, fase a fase. Quando uma fase terminar, marque o status aqui.

## Princípio que guia tudo

Organizar **por domínio, nunca por camada** (não existe pacote `handlers/`/`services/` — camada é *prefixo de arquivo*). Os domínios de mídia seguem **supertype → extensão**, espelhando o banco (uma tabela `files` + tabelas-complemento por tipo):

```
internal/api/v1/
  files/      NÚCLEO genérico (supertype). FileModel/FileDto, CRUD, tree, listing,
              operations, recent, reports, blob/thumbnail genérico.
              NÃO importa image/music/video.
  image/      extensão  ─┐
  music/      extensão   ├─ importam files, dão JOIN na tabela files.
  video/      extensão  ─┘  Dependência num sentido só: extensão → files.
internal/worker/
  job/    tipos/enums (pacote neutro, sem ciclo)
  engine/ pool, orquestrador, scheduler, executores
  steps/  um arquivo por step de job
  scan/   pipeline de scan/index de arquivo
```

Regras de ouro:
- **`files` (núcleo) não conhece as extensões.** Se precisar conhecer, a modelagem está errada.
- **Pacote não é dono de tabela.** A extensão dá `JOIN` em `files` à vontade; suas queries ficam em `pkg/database/queries/<domínio>/`.
- **Composição por tipo acontece na borda** (frontend ou handler que importa vários domínios), nunca pelo núcleo.

## Restrição inegociável: o contrato HTTP não muda

O backend é o ponto de integração de frontend, Android (x2) e plugin (ver `CLAUDE.md` raiz). **Nenhuma fase altera path, método ou shape de resposta de rota existente.** Mover código de `files` para `image`/`music`/`video` mantém a **mesma URL**; só muda quem a serve internamente. Mudança de contrato é outro trabalho, fora deste plano.

## Ordem, risco e status

Fases ordenadas do menor risco ao maior. Cada uma termina **verde no `make ci-backend`** (gofmt + vet + testes ≥80%) antes da próxima. Pode parar em qualquer fase.

| Fase | Arquivo | Escopo | Risco | Status |
|---|---|---|---|---|
| 0 | [phase-0-naming-conventions.md](phase-0-naming-conventions.md) | nomes `snake_case` + regra no CLAUDE.md | mínimo | ✅ concluída (2026-06-09) |
| 1 | [phase-1-worker-split.md](phase-1-worker-split.md) | `worker/` → job/engine/steps/scan | médio | ✅ concluída (2026-06-09) |
| 2 | [phase-2-image-extension.md](phase-2-image-extension.md) | criar `image/` (extrair de files) | baixo | ✅ concluída (2026-06-09) |
| 3 | [phase-3-music-extension.md](phase-3-music-extension.md) | mover música de files → `music/` | médio | ✅ concluída (2026-06-10) |
| 4 | [phase-4-video-extension.md](phase-4-video-extension.md) | mover vídeo de files → `video/` | médio | ✅ concluída (2026-06-10) |
| 5 | [phase-5-files-core-cleanup.md](phase-5-files-core-cleanup.md) | arrumar núcleo `files` por arquivo | baixo | ✅ concluída (2026-06-10) |

> **Refatoração concluída** (2026-06-10): todas as fases verdes no `make ci`. A regra resultante vive nos `CLAUDE.md`; estes docs ficam como registro do caminho.

## Decisão registrada

- **Não renomear `files` → `file`.** O diretório é importado por ~28 lugares; renomear seria churn cosmético. `files.FileModel`/`FileDto` permanece o núcleo estável — é isso que segura o raio de impacto.
- Cada fase é **um ou mais commits lógicos** (backend isolado), conforme o workflow do projeto (commits diretos em `develop`).
