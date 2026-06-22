---
name: react-concurrent-hooks
description: >-
  Ensina os hooks concorrentes do React: useTransition() → [isPending,
  startTransition] para marcar updates de state como transições não-bloqueantes
  (mantém a UI responsiva e mostra estado pending); useDeferredValue(value) que
  devolve uma cópia "atrasada" do valor para updates não-urgentes; e a API
  startTransition standalone. Cobre as refs oficiais
  react.dev/reference/react/useTransition, /useDeferredValue e /startTransition.
  Use quando o usuário disser "useTransition", "startTransition", "isPending",
  "useDeferredValue", "UI trava ao digitar/filtrar lista grande", "update não
  urgente", ou "manter a interface responsiva".
when_to_use: >-
  Disparar ao tornar updates pesados não-bloqueantes, mostrar pending de navegação/
  filtro, ou deferir um valor caro de renderizar. NÃO usar para memoizar cálculo
  (react-performance-hooks) nem para debounce de rede (use debounce comum).
---

# React: Hooks concorrentes (useTransition, useDeferredValue)

Fontes: react.dev/reference/react/useTransition,
react.dev/reference/react/useDeferredValue, react.dev/reference/react/startTransition

Permitem marcar parte das atualizações como **não-urgentes (transições)**, para o
React manter a UI responsiva e interromper renders pesados.

## useTransition

```jsx
const [isPending, startTransition] = useTransition();

function selectTab(nextTab) {
  startTransition(() => {
    setTab(nextTab); // update marcado como transição (não bloqueia digitação/cliques)
  });
}
// isPending → mostre spinner/opacity enquanto a transição renderiza
```

- Updates dentro de `startTransition` são **não-bloqueantes** e podem ser
  interrompidos por updates urgentes (ex: digitação).
- `isPending` indica que a transição está em andamento.
- Suporta **Actions assíncronas** (await dentro do callback) em versões recentes.

## startTransition (standalone)

Mesma ideia sem o `isPending`, fora de um componente:
```jsx
import { startTransition } from 'react';
startTransition(() => setState(next));
```

## useDeferredValue

Devolve uma cópia do valor que **fica para trás** durante updates rápidos, deixando
a parte cara renderizar depois:

```jsx
const deferredQuery = useDeferredValue(query);
// <input> usa `query` (urgente); a lista pesada usa `deferredQuery` (atrasa)
const results = useMemo(() => search(deferredQuery), [deferredQuery]);
```

Combine com `<Suspense>` e `memo` para não bloquear a digitação.

## useTransition vs useDeferredValue

- **useTransition**: você controla **qual setState** é transição (envolve a chamada).
- **useDeferredValue**: você não controla o setState (ex: valor vem de prop) e quer
  **deferir o valor derivado**.

## Armadilhas

- O update dentro de `startTransition` deve ser **síncrono** ao agendar o state (em
  versões mais antigas; Actions assíncronas exigem React recente).
- Marcar input controlado como transição → input "engasga"; o texto digitado deve
  ser update **urgente** (fora da transição).
- `useDeferredValue` não substitui **debounce de requisição** — ele difere o render,
  não a chamada de rede.
- Sem `memo`/`useMemo` na parte cara, deferir pouco ajuda.

## Verificação

- O update urgente (digitação/clique) ficou fora da transição?
- A parte pesada usa o valor deferido/transição e está memoizada?

Relacionados: [[react-performance-hooks]], [[react-suspense-lazy]], [[react-form-actions]].
