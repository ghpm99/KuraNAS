---
name: react-reducer
description: >-
  Ensina useReducer em React: consolidar a lógica de atualização de state numa
  função pura reducer(state, action) => newState, dispatch de actions com type,
  switch por action.type, migração de useState→useReducer em 3 passos, useState vs
  useReducer, e useImmerReducer. Cobre "Extracting State Logic into a Reducer" de
  react.dev/learn. Use quando o usuário disser "useReducer", "muitos setState
  espalhados", "dispatch", "action type", "lógica de state complexa", "reducer", ou
  "como organizo atualizações de state relacionadas".
when_to_use: >-
  Disparar ao ter muitos handlers atualizando o mesmo state, ao introduzir
  dispatch/actions, ou ao migrar de useState para reducer. NÃO usar para Context
  (react-context) nem combinação reducer+context (react-reducer-context).
---

# React: useReducer

Fonte: react.dev/learn/extracting-state-logic-into-a-reducer

Um **reducer** é uma função pura que concentra a lógica de atualização fora do
componente: `(state, action) => próximoState`. Os handlers só **descrevem o que
aconteceu** (dispatch de actions); o reducer decide **como** o state muda.

## Migração useState → useReducer (3 passos)

```jsx
// 1. Handlers despacham actions (objetos com type):
function handleAddTask(text) {
  dispatch({ type: 'added', id: nextId++, text });
}

// 2. Reducer com switch por action.type (puro, imutável):
function tasksReducer(tasks, action) {
  switch (action.type) {
    case 'added':
      return [...tasks, { id: action.id, text: action.text, done: false }];
    case 'changed':
      return tasks.map(t => t.id === action.task.id ? action.task : t);
    case 'deleted':
      return tasks.filter(t => t.id !== action.id);
    default:
      throw Error('Unknown action: ' + action.type);
  }
}

// 3. useReducer no componente:
const [tasks, dispatch] = useReducer(tasksReducer, initialTasks);
```

## Regras de um bom reducer

- **Puro**: mesmos inputs → mesmo output; sem side effects (fetch, timeout, mutação).
- **Imutável**: copie com spread / métodos de array ([[react-state-objects-arrays]]).
- **Uma action por interação do usuário**: prefira `{type:'reset_form'}` a cinco
  `set_field` — gera um histórico claro pra debug.

## useState vs useReducer

| | useState | useReducer |
|---|---|---|
| Código | menos inicial | mais inicial, escala melhor |
| Complexidade | bom p/ simples | separa "o que" de "como" |
| Debug | difícil achar a mudança errada | log no reducer mostra cada update |
| Teste | no contexto do componente | função pura, testável isolada |

Pode misturar `useState` e `useReducer` no mesmo componente. Para syntax enxuta com
mutação aparente, use `useImmerReducer` (`draft.push(...)`).

## Armadilhas

- Side effect dentro do reducer (fetch, `Date.now()`, mutar `tasks`) → impuro, bugs.
- Esquecer o `default` do switch → actions desconhecidas passam silenciosas.
- Mutar o state no case (`tasks.push`) em vez de retornar novo array.
- Action sem `type` ou com dados faltando para o case.

## Verificação

- Reducer é puro e imutável? Tem `default`? Cada action mapeia uma interação?
- Handlers só fazem `dispatch`, sem lógica de state?

Relacionados: [[react-state-usestate]], [[react-context]], [[react-reducer-context]].
