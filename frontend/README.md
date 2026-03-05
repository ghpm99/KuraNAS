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
│   ├── components/     # componentes por domínio e providers
│   ├── pages/          # páginas de rota (wrappers)
│   ├── service/        # clientes e serviços de API
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

- Dev server:
```bash
yarn dev
```
- Build de produção:
```bash
yarn build
```
- Preview local do build:
```bash
yarn preview
```
- Lint:
```bash
yarn lint
```
- Testes:
```bash
yarn test --watchAll=false
```
- Cobertura:
```bash
yarn coverage
```
- Typecheck da configuração de testes:
```bash
yarn typecheck:test
```

## Variáveis de Ambiente

Arquivos padrão:

- `.env.development`
- `.env.production`

Variáveis usadas:

- `VITE_API_URL`: URL base da API backend.
- `VITE_DEBUG_MODE`: habilita flags de debug no ambiente de desenvolvimento.

## API e i18n

- O frontend consome a API via `src/service/index.ts` com base URL de `getApiV1BaseUrl()` (`src/service/apiUrl.ts`).
- Textos visíveis devem vir de tradução via `useI18n()`.
- Não adicionar texto hardcoded em componentes.
- Novas mensagens devem ser adicionadas primeiro em `backend/translations` e consumidas por chave.

## Padrões de Implementação

Antes de alterar código frontend, siga:

- `docs/standards/frontend-standards.md`

Pontos obrigatórios:

- lógica e chamadas HTTP em hooks/providers, não em componentes de render;
- uso de alias `@/...`;
- testes com cobertura mínima global de 80%.
