---
name: react-performance-hooks
description: >-
  Ensina os hooks de performance do React useMemo e useCallback: useMemo(calc,
  deps) memoiza o RESULTADO de um cálculo caro; useCallback(fn, deps) memoiza a
  própria FUNÇÃO (referência estável para filhos memoizados/deps de effect);
  quando realmente valem a pena (com memo, deps de outro hook) e quando NÃO usar
  (otimização prematura). Cobre as refs oficiais react.dev/reference/react/useMemo
  e /useCallback. Use quando o usuário disser "useMemo", "useCallback", "memoizar",
  "cálculo caro re-roda", "componente re-renderiza demais", "função nova a cada
  render quebra o memo", ou "estabilizar referência".
when_to_use: >-
  Disparar ao memoizar cálculo/função, estabilizar referência para React.memo ou
  deps de useEffect, ou debater otimização de render. NÃO usar para evitar effect
  desnecessário em geral (react-you-might-not-need-effect) nem React.memo de
  componente (react-api-utilities).
---

# React: useMemo e useCallback

Fontes: react.dev/reference/react/useMemo, react.dev/reference/react/useCallback

Ambos **cacheiam entre renders** com base num array de dependências. Só recalculam
quando alguma dep muda (comparação por `Object.is`).

## useMemo — cacheia o RESULTADO

```jsx
const visibleTodos = useMemo(
  () => getFilteredTodos(todos, filter), // recalcula só se todos/filter mudarem
  [todos, filter]
);
```

Usos válidos:
- **Cálculo caro** que não deve rodar todo render.
- Manter **referência estável** de um objeto/array que é **dep de outro hook**
  (`useEffect`/`useMemo`) ou prop de um componente memoizado.
- Memoizar um **nó JSX** passado a um filho memoizado.

## useCallback — cacheia a FUNÇÃO

```jsx
const handleSubmit = useCallback((data) => {
  post('/api', data);
}, []); // mesma referência entre renders
```

`useCallback(fn, deps)` ≡ `useMemo(() => fn, deps)`. Use para passar callbacks
estáveis a filhos envoltos em `React.memo`, ou como dep de `useEffect` (sem
reconectar a cada render).

## Quando NÃO usar

- A maioria dos componentes **não precisa**. Memoizar tem custo (memória + comparação).
- Só ajuda se: (a) o cálculo é realmente caro, (b) o valor é passado a um componente
  envolto em `memo`, ou (c) é dep de outro hook.
- Prefira primeiro **reduzir trabalho** (state melhor estruturado, calcular no render,
  key para reset) antes de memoizar.

## Armadilhas

- Memoizar tudo "por garantia" → código mais complexo sem ganho.
- Deps faltando → valor stale; deps demais → memo nunca acerta. Liste exatamente o
  que o cálculo/função lê.
- `useCallback` numa função que não vira dep nem prop de `memo` → inútil.
- Esquecer que o cálculo dentro de `useMemo` deve ser **puro**.

## Verificação

- Mediu/observou que vale a pena (cálculo caro, `memo`, ou dep de hook)?
- Deps batem com o que é lido? A função/cálculo é pura?

Relacionados: [[react-you-might-not-need-effect]], [[react-concurrent-hooks]], [[react-api-utilities]].
