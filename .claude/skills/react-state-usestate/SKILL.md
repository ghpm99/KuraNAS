---
name: react-state-usestate
description: >-
  Ensina state com useState em React: por que variável local não basta (não
  persiste entre renders nem dispara re-render), a sintaxe const [x, setX] =
  useState(inicial), o que useState retorna, as Rules of Hooks (chamar no topo,
  nunca em condição/loop), state é privado por instância, e múltiplos states.
  Cobre "State: A Component's Memory" de react.dev/learn. Use quando o usuário
  disser "useState", "como guardo estado", "a tela não atualiza ao mudar a
  variável", "Rules of Hooks", "Invalid hook call", "useState dentro de if", ou
  "componente não re-renderiza".
when_to_use: >-
  Disparar ao adicionar estado, quando mudar uma variável local não atualiza a
  UI, ou ao ver erro de Rules of Hooks. NÃO usar para batching/updater functions
  em profundidade (react-state-updates) nem objetos/arrays (react-state-objects-arrays).
---

# React: State (useState)

Fonte oficial: https://react.dev/learn/state-a-components-memory

## Por que variável local não basta

```jsx
let index = 0;             // 🔴
function handleClick() { index = index + 1; }
```

Falha por dois motivos: (1) **variáveis locais não persistem entre renders** — o
React recria o componente do zero; (2) **mudá-las não dispara re-render**.

State resolve os dois: **retém** o dado entre renders **e dispara** novo render.

## useState

```jsx
import { useState } from 'react';

const [index, setIndex] = useState(0);
//      ^valor   ^setter        ^inicial
```

- `index`: valor atual do state.
- `setIndex`: setter que atualiza **e** dispara re-render.
- `0`: valor inicial (usado só no primeiro render).

Chamar `setIndex(1)` → React re-renderiza e agora `useState` devolve `[1, setIndex]`.

Múltiplos states são permitidos:
```jsx
const [index, setIndex] = useState(0);
const [showMore, setShowMore] = useState(false);
```

## Rules of Hooks

Hooks (funções `use*`) só podem ser chamados **no topo** do componente ou de um
custom Hook. **Nunca** dentro de condições, loops ou funções aninhadas:

```jsx
if (cond) { const [s, setS] = useState(0); } // 🔴 Invalid hook call
```

React depende da **ordem estável** das chamadas para casar cada state.

## State é local e privado

Cada instância de componente tem seu próprio state. Dois `<Gallery />` lado a lado
têm states independentes. O pai **não acessa** o state do filho (encapsulamento).

## Armadilhas

- Atualizar a UI mudando variável comum em vez de state → nada re-renderiza.
- `useState` dentro de `if`/loop/callback → "Invalid hook call" ou bugs de ordem.
- Esperar que o valor mude **na mesma execução** após `setX` — não muda; o novo
  valor só aparece no **próximo render** (state é snapshot → [[react-state-updates]]).
- Passar o resultado do setter como inicial, ou recomputar caro a cada render
  (use inicializador preguiçoso `useState(() => calcCaro())`).

## Verificação

- Todo dado que muda e afeta a UI está em state?
- Todos os hooks no topo, sem condição/loop? Mesma ordem todo render?

Relacionados: [[react-state-updates]], [[react-state-objects-arrays]], [[react-handling-events]].
