---
name: react-reducer-context
description: >-
  Ensina a combinar useReducer + Context em React para escalar state global sem
  prop drilling: criar dois contexts (state e dispatch), prover ambos no topo a
  partir do useReducer, consumir com useContext nos filhos fundos, e encapsular
  tudo num Provider + hooks custom. Cobre "Scaling Up with Reducer and Context" de
  react.dev/learn. Use quando o usuário disser "reducer com context", "state
  global sem prop drilling", "dispatch disponível em qualquer componente", "como
  escalo o state da app", ou "Provider que junta reducer e context".
when_to_use: >-
  Disparar ao expor state + dispatch de um useReducer para componentes distantes,
  ou ao montar um Provider de feature/app. NÃO usar para reducer isolado
  (react-reducer) nem context isolado (react-context).
---

# React: Escalando com Reducer + Context

Fonte: react.dev/learn/scaling-up-with-reducer-and-context

Combine [[react-reducer]] (centraliza a lógica) com [[react-context]] (entrega
sem prop drilling) para state que muitos componentes distantes leem **e** alteram.

## Padrão: dois contexts (state e dispatch)

```jsx
// TasksContext.js
import { createContext, useContext, useReducer } from 'react';

const TasksContext = createContext(null);
const TasksDispatchContext = createContext(null);

export function TasksProvider({ children }) {
  const [tasks, dispatch] = useReducer(tasksReducer, initialTasks);
  return (
    <TasksContext value={tasks}>
      <TasksDispatchContext value={dispatch}>
        {children}
      </TasksDispatchContext>
    </TasksContext>
  );
}

// Hooks custom para consumir (esconde o useContext):
export function useTasks() { return useContext(TasksContext); }
export function useTasksDispatch() { return useContext(TasksDispatchContext); }
```

Uso nos filhos, em qualquer profundidade:

```jsx
const tasks = useTasks();
const dispatch = useTasksDispatch();
dispatch({ type: 'added', id, text });
```

E no topo da app/feature: `<TasksProvider> ... </TasksProvider>`.

## Por que dois contexts

Separar `state` de `dispatch` deixa componentes que só **disparam** actions não
dependerem do valor do state. O reducer e os dois providers ficam num único
arquivo, mantendo os componentes limpos.

## Armadilhas

- Colocar `tasks` e `dispatch` no **mesmo** objeto de value → recria o objeto a cada
  render e re-renderiza consumidores à toa. Use dois contexts (ou memoize).
- Esquecer de envolver a árvore no `<TasksProvider>` → `useTasks()` devolve o default (null).
- Lógica de side effect no reducer (continua proibido — ele é puro).
- Acessar dispatch fora do Provider → null → erro ao chamar.

## Verificação

- Provider envolve todos os consumidores? Hooks custom escondem o `useContext`?
- state e dispatch em contexts separados? Reducer puro?

Relacionados: [[react-reducer]], [[react-context]], [[react-sharing-state]].
