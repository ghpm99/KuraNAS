---
name: react-effects
description: >-
  Ensina useEffect em React: Effects rodam DEPOIS do render para sincronizar com
  sistemas externos, o array de dependências ([] = só no mount, [dep] = quando dep
  muda, sem array = todo render), a função de cleanup (return () => ...) que roda
  antes do próximo effect e no unmount, por que rodam 2x em dev (StrictMode), e
  padrões (conexão, subscription, timer, fetch com flag ignore). Cobre
  "Synchronizing with Effects" e "Lifecycle of Reactive Effects" de
  react.dev/learn. Use quando o usuário disser "useEffect", "effect roda duas
  vezes", "cleanup", "array de dependências", "subscription/addEventListener",
  "fetch no effect", ou "sincronizar com sistema externo".
when_to_use: >-
  Disparar ao escrever useEffect, configurar deps/cleanup, ou debugar effect que
  roda demais. NÃO usar quando o effect é desnecessário (react-you-might-not-need-effect)
  nem para dependências reativas/eventos de effect (react-effect-events).
---

# React: Effects (useEffect)

Fontes: react.dev/learn/synchronizing-with-effects,
react.dev/learn/lifecycle-of-reactive-effects

**Effects** rodam **depois do render** para sincronizar o componente com sistemas
externos (DOM não-React, libs, rede, browser APIs). Diferente do render (puro) e de
handlers (resposta a interação).

```jsx
useEffect(() => {
  // roda depois do render
}, [dependencies]);
```

## Array de dependências

```jsx
useEffect(() => {...});        // sem array → todo render
useEffect(() => {...}, []);    // [] → só no mount
useEffect(() => {...}, [dep]); // [dep] → quando dep muda
```

As deps **devem bater** com tudo que o effect lê (props/state/funções reativas).

## Cleanup

Retorne uma função para desfazer o que o effect fez:

```jsx
useEffect(() => {
  const connection = createConnection();
  connection.connect();
  return () => connection.disconnect(); // cleanup
}, []);
```

Roda: **antes** do effect rodar de novo (deps mudaram), no **unmount**, e em dev
após o remount do StrictMode.

## Padrões

```jsx
// subscription
useEffect(() => {
  window.addEventListener('scroll', onScroll);
  return () => window.removeEventListener('scroll', onScroll);
}, []);

// fetch com flag para ignorar resposta obsoleta (race condition)
useEffect(() => {
  let ignore = false;
  fetchTodos(userId).then(json => { if (!ignore) setTodos(json); });
  return () => { ignore = true; };
}, [userId]);
```

## Roda 2x em dev (StrictMode)

Em dev, o React **remonta** o componente uma vez para verificar se o cleanup está
correto (Connecting → Disconnected → Connecting). **Não é bug**; em produção roda
uma vez. Se o "2x" causa problema, o conserto é **implementar cleanup**, não evitar.

## Armadilhas

- Usar `useRef` para impedir o effect de rodar 2x → esconde o bug; faça cleanup.
- Esquecer cleanup em subscriptions/timers/conexões → vazamentos e duplicação.
- Deps incompletas (linter `react-hooks/exhaustive-deps` avisa) → valores stale.
- Fazer fetch sem flag `ignore` → race conditions; prefira libs (TanStack Query/SWR)
  ou solução do framework.
- Usar effect para transformar dados de render ou tratar eventos → [[react-you-might-not-need-effect]].

## Verificação

- O effect realmente sincroniza com um sistema **externo**?
- Deps batem com o que ele lê? Tem cleanup para tudo que cria?

Relacionados: [[react-you-might-not-need-effect]], [[react-effect-events]], [[react-custom-hooks]].
