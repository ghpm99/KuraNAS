# Frontend KuraNAS

Aplicação web do KuraNAS construída com React + TypeScript + Vite.

## Stack

- React 19
- TypeScript
- Vite
- MUI
- React Query
- Jest + Testing Library
- ESLint

## Estrutura

```text
frontend/
├── src/
│   ├── app/            # composição de rotas e inicialização da aplicação
│   ├── components/     # composição de tela, layout e domínios não migrados
│   ├── features/       # ownership por domínio (files, music, videos)
│   ├── pages/          # wrappers finos de rota
│   ├── service/        # clientes e serviços de API
│   ├── shared/         # utilitários compartilhados cross-feature
│   ├── types/          # tipos compartilhados
│   └── utils/          # utilitários
├── public/
├── jest.config.js
├── eslint.config.js
└── vite.config.ts
```

## Setup

```bash
cd frontend
yarn
```

## Scripts

Dev server:

```bash
yarn dev
```

Build de produção:

```bash
yarn build
```

Preview local do build:

```bash
yarn preview
```

Lint:

```bash
yarn lint
```

Testes:

```bash
yarn test --watchAll=false
```

Teste em watch:

```bash
yarn test:watch
```

Cobertura:

```bash
yarn coverage
```

Typecheck da configuração de testes:

```bash
yarn typecheck:test
```

Format:

```bash
yarn format
```

## Variáveis de Ambiente

Variável suportada:

- `VITE_API_URL`: URL base da API (sem `/api/v1` no final).

Exemplo:

```dotenv
VITE_API_URL=http://localhost:8000
```

Comportamento da URL base:

- Se `globalThis.__KURANAS_API_URL__` existir em runtime, ela tem prioridade.
- Senão, usa `VITE_API_URL`.
- Sem variável, fallback para caminho relativo (`/api/v1`).

## API e i18n

- O frontend consome a API via `src/service/index.ts` usando `getApiV1BaseUrl()` (`src/service/apiUrl.ts`).
- Textos visíveis devem vir de tradução via `useI18n()`.
- Não adicionar texto hardcoded em componentes.
- Novas mensagens devem ser adicionadas primeiro em `backend/translations` e consumidas por chave.

## Padrões de Implementação

Antes de alterar código frontend, siga:

- `/docs/standards/frontend-standards.md`

Pontos obrigatórios:

- domínios `files/music/videos` devem evoluir prioritariamente em `src/features/*`;
- lógica e chamadas HTTP em hooks/providers, não em componentes de render;
- uso de alias `@/...`;
- testes com cobertura mínima global configurada em `jest.config.js` (90% lines/functions/statements e 89% branches).
