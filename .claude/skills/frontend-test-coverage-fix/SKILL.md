---
name: frontend-test-coverage-fix
description: >-
  Aumenta cobertura de testes frontend preferindo chamadas reais de endpoint
  (axios + msw/axios-mock-adapter) em vez de mocks pesados de componente. Analisa
  quais arquivos têm cobertura baixa, prioriza os mais críticos e escreve testes
  que exercitam o contrato de API real. TRIGGER quando: usuário disser "aumenta
  coverage", "cobertura está baixa", "adiciona testes no frontend", "testes não
  cobrem X". SKIP: projetos sem Jest/Vitest; quando o problema for memória/OOM
  (use jest-test-memory-debug primeiro).
---

# Frontend Test Coverage Fix

## Princípio: prefira endpoint-touching sobre mock de componente

O projeto prefere que chamadas de endpoint **aconteçam de verdade** nos testes —
interceptadas por `axios-mock-adapter` ou `msw`, mas o fluxo completo de
"componente → hook → axios → resposta → render" deve ser exercitado.

Mocks de componente (`jest.mock('../MyComponent')`) e mocks de hook
(`jest.mock('../useMyHook')`) testam implementação, não contrato. Evite-os.

## Passo 1 — Identifica arquivos com baixa cobertura

```bash
# Roda coverage e mostra os arquivos mais descobertos
npx jest --coverage --coverageReporters text 2>&1 | \
  grep -v "node_modules" | sort -t"|" -k4 -n | head -30

# Alternativa: abre relatório HTML
npx jest --coverage --coverageReporters html && open coverage/lcov-report/index.html
```

Prioriza arquivos de:
1. **Hooks customizados** (`useX.ts`) — geralmente zero cobertura, máximo impacto
2. **Serviços de API** (`*Service.ts`, `*Api.ts`) — calls de endpoint não testadas
3. **Componentes de página** (pages/*, screens/*) — fluxo completo não exercitado
4. **Utilitários** (`*utils.ts`, `*helpers.ts`) — funções puras fáceis de testar

## Passo 2 — Estrutura de teste preferida

### Para hooks com chamadas de endpoint

```tsx
import { renderHook, waitFor } from "@testing-library/react";
import MockAdapter from "axios-mock-adapter";
import axios from "axios";
import { useMyData } from "./useMyData";

const mock = new MockAdapter(axios);

afterEach(() => {
  mock.reset(); // limpa handlers, não remove o adapter
});

afterAll(() => {
  mock.restore(); // restaura axios original
});

describe("useMyData", () => {
  it("retorna dados quando endpoint responde 200", async () => {
    mock.onGet("/api/my-data/").reply(200, [{ id: 1, name: "Item" }]);

    const { result } = renderHook(() => useMyData());

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data).toHaveLength(1);
    expect(result.current.data[0].name).toBe("Item");
  });

  it("expõe erro quando endpoint retorna 500", async () => {
    mock.onGet("/api/my-data/").reply(500, { detail: "Erro interno" });

    const { result } = renderHook(() => useMyData());

    await waitFor(() => expect(result.current.error).toBeTruthy());
    expect(result.current.data).toEqual([]);
  });
});
```

### Para componentes com fetch de dados

```tsx
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import MockAdapter from "axios-mock-adapter";
import axios from "axios";
import { MyPage } from "./MyPage";

const mock = new MockAdapter(axios);

afterEach(() => mock.reset());
afterAll(() => mock.restore());

describe("MyPage", () => {
  it("exibe lista após carregar", async () => {
    mock.onGet("/api/items/").reply(200, [{ id: 1, name: "Primeiro" }]);

    render(<MyPage />);

    // Estado de loading
    expect(screen.getByText(/carregando/i)).toBeInTheDocument();

    // Dados carregados
    await waitFor(() =>
      expect(screen.getByText("Primeiro")).toBeInTheDocument()
    );
  });

  it("permite submeter formulário e faz POST", async () => {
    mock.onPost("/api/items/").reply(201, { id: 2, name: "Novo" });
    mock.onGet("/api/items/").reply(200, []);

    render(<MyPage />);

    await userEvent.type(screen.getByLabelText(/nome/i), "Novo");
    await userEvent.click(screen.getByRole("button", { name: /salvar/i }));

    await waitFor(() =>
      expect(mock.history.post).toHaveLength(1)
    );
    expect(JSON.parse(mock.history.post[0].data)).toMatchObject({ name: "Novo" });
  });
});
```

### Para serviços de API puros

```ts
import MockAdapter from "axios-mock-adapter";
import axios from "axios";
import { fetchItems, createItem } from "./itemService";

const mock = new MockAdapter(axios);
afterEach(() => mock.reset());
afterAll(() => mock.restore());

test("fetchItems chama GET /api/items/ e retorna array", async () => {
  mock.onGet("/api/items/").reply(200, [{ id: 1 }]);
  const result = await fetchItems();
  expect(result).toHaveLength(1);
});

test("createItem chama POST com payload correto", async () => {
  mock.onPost("/api/items/").reply(201, { id: 2, name: "Novo" });
  const result = await createItem({ name: "Novo" });
  expect(mock.history.post[0].data).toBe(JSON.stringify({ name: "Novo" }));
  expect(result.id).toBe(2);
});
```

## Passo 3 — Cenários obrigatórios por tipo de arquivo

Cada arquivo crítico precisa cobrir:

| Cenário | Cobertura que garante |
|---------|----------------------|
| Resposta 200 com dados | happy path |
| Resposta com lista vazia | estado vazio |
| Resposta 4xx (400/401/404) | tratamento de erro do usuário |
| Resposta 5xx | tratamento de erro de servidor |
| Loading state | UX de carregamento |
| Ação do usuário (click, submit) | interatividade |

## Passo 4 — Verifica que coverage aumentou

```bash
npx jest --coverage --coverageReporters text 2>&1 | grep -E "File|All files"
```

## Armadilhas comuns

- **`mock.reset()` vs `mock.restore()`**: `reset()` limpa os handlers mas mantém
  o adapter ativo (use no `afterEach`). `restore()` remove o adapter completamente
  (use no `afterAll`).
- **`waitFor` sem timeout suficiente**: se o componente faz múltiplos fetches em
  sequência, aumenta o timeout: `await waitFor(() => ..., { timeout: 3000 })`.
- **Act warning**: toda atualização de state em teste deve estar dentro de `act`.
  `waitFor` já wrapa automaticamente, mas renders manuais precisam de:
  ```tsx
  act(() => { render(<MyComponent />); });
  ```
- **Axios com baseURL**: se o projeto configura `axios.defaults.baseURL`, o mock
  precisa interceptar o path completo ou relativo conforme a configuração.

## Relação com outras skills

- `jest-test-memory-debug` — se a suite OOM antes de chegar na cobertura
- `convert-mock-tests` — converte testes que usam mocks pesados de módulo
