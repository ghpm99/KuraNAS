---
name: jest-test-memory-debug
description: >-
  Diagnostica e corrige falhas de memória / heap / stack em suites Jest: "Maximum
  call stack size exceeded", "Jest worker ran out of memory", "Reached heap limit
  allocation failed". TRIGGER quando: a suite Jest trava ou mata o processo; testes
  que passavam individualmente falham ao rodar juntos; logs mostram OOM, stack
  overflow ou "RangeError: Maximum call stack size exceeded". SKIP: erros de
  assertion (expect falhou), erros de TypeScript, falhas de importação de módulo.
---

# Jest Test Memory Debug

## Causas mais comuns (por frequência)

### 1. Recursão circular em mocks ou módulos
`Maximum call stack size exceeded` quase sempre é mock circular ou módulo que
importa a si mesmo através de barrel exports.

Sinais:
- O stack trace aponta para `node_modules/jest-circus/build/jestAdapterInit.js`
- O erro ocorre antes mesmo de qualquer `it()` rodar
- `--runInBand` (serial) não resolve

Diagnóstico:
```bash
# Identifica imports circulares
npx madge --circular src/

# Roda um arquivo isolado para confirmar que o problema é de setup
npx jest path/to/broken.test.ts --runInBand --detectOpenHandles
```

Correção:
- Mover a inicialização do mock para dentro do `beforeEach` em vez do topo do módulo
- Quebrar o barrel export que cria o ciclo

### 2. Worker OOM por acúmulo de testes pesados
`Jest worker ran out of memory` / `Reached heap limit allocation failed`

Sinais:
- Erro aparece após vários arquivos rodarem
- Roda bem com `--runInBand` (mas lento)
- Testes usam `axios-mock-adapter`, `msw` ou objetos grandes no escopo do módulo

Diagnóstico:
```bash
# Aumenta heap do worker como teste de isolamento
NODE_OPTIONS=--max-old-space-size=4096 npx jest

# Identifica arquivos que consomem mais memória
npx jest --logHeapUsage 2>&1 | grep -E "heap|MB"

# Roda cada arquivo em processo separado (confirma que o problema é acúmulo)
npx jest --maxWorkers=1 --forceExit
```

Correção padrão em `jest.config.js` / `jest.config.ts`:
```js
module.exports = {
  // ...
  workerIdleMemoryLimit: "512MB",   // descarta worker quando passa do limite
  maxWorkers: "50%",                 // menos workers em paralelo
  testTimeout: 10000,                // mata testes lentos antes de OOM
};
```

### 3. Mock de axios / fetch que não é restaurado
`axios-mock-adapter` ou `jest.spyOn(axios, 'get')` sem `afterEach` → adapter
acumula handlers → leak de memória entre arquivos.

```ts
// ERRADO — handler não é limpo
const mock = new MockAdapter(axios);
test("...", () => { mock.onGet("/foo").reply(200, {}); });

// CERTO
let mock: MockAdapter;
beforeEach(() => { mock = new MockAdapter(axios); });
afterEach(() => { mock.restore(); });
```

### 4. Timers falsos não restaurados
`jest.useFakeTimers()` sem `jest.useRealTimers()` no teardown → pode criar
timers infinitos se o código usa `setInterval`.

```ts
beforeEach(() => jest.useFakeTimers());
afterEach(() => jest.useRealTimers());
```

### 5. Componentes com estado global vazando entre testes (React)
Store Zustand/Redux inicializado no topo do módulo não é resetado entre testes.

```ts
// Em cada describe que usa a store
beforeEach(() => {
  useMyStore.setState(initialState);
});
```

## Passo a passo de diagnóstico

1. **Isola o arquivo** — roda só o arquivo que falha:
   ```bash
   npx jest path/to/file.test.ts --runInBand
   ```

2. **Adiciona `--detectOpenHandles`** — revela timers e conexões abertas:
   ```bash
   npx jest --detectOpenHandles --forceExit
   ```

3. **Habilita `--logHeapUsage`** — identifica qual arquivo drena memória:
   ```bash
   npx jest --logHeapUsage --runInBand 2>&1 | grep -E "PASS|FAIL|heap"
   ```

4. **Verifica imports circulares**:
   ```bash
   npx madge --circular src/ || npx dpdm --no-warning src/index.ts
   ```

5. **Revisa todos os mocks no arquivo suspeito** procurando:
   - `MockAdapter` sem `.restore()` no `afterEach`
   - `jest.spyOn` sem `jest.restoreAllMocks()` ou `.mockRestore()`
   - `jest.useFakeTimers` sem `jest.useRealTimers()`
   - Objetos grandes criados no escopo do módulo (fora de `beforeEach`)

6. **Verifica configuração Jest** — adiciona limites de memória e timeout.

## Configuração defensiva recomendada

```js
// jest.config.js
module.exports = {
  workerIdleMemoryLimit: "512MB",
  maxWorkers: "50%",
  testTimeout: 15000,
  clearMocks: true,       // limpa mock.calls entre testes
  restoreMocks: true,     // restaura spyOn entre testes
  resetMocks: false,      // não reseta implementação (use com cuidado)
};
```

## Após corrigir

Roda a suite completa e confirma que o erro não volta:
```bash
npm test -- --coverage --bail
```

Se coverage caiu, a skill `frontend-test-coverage-fix` cobre como recuperá-la.
