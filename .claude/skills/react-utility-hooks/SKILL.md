---
name: react-utility-hooks
description: >-
  Ensina os hooks utilitários do React: useId (IDs únicos estáveis p/
  acessibilidade, não para keys de lista), useImperativeHandle (expor métodos
  imperativos via ref a partir de um componente), useLayoutEffect (effect síncrono
  antes do paint, p/ medir layout), useDebugValue (label no DevTools p/ custom
  hooks), useSyncExternalStore (assinar store externa de forma segura p/
  concorrência/SSR), e use (ler Promise/Context, pode ser chamado condicionalmente).
  Cobre as refs oficiais em react.dev/reference/react. Use quando o usuário disser
  "useId", "useImperativeHandle", "useLayoutEffect", "useSyncExternalStore", "use()
  hook", "ler promise com use", "id de acessibilidade", ou "expor método via ref".
when_to_use: >-
  Disparar ao usar um desses hooks específicos. NÃO usar para useEffect comum
  (react-effects), useRef/DOM (react-refs), ou performance (react-performance-hooks).
---

# React: Hooks utilitários

Fonte: react.dev/reference/react (páginas de cada hook)

## useId
IDs únicos e **estáveis** entre servidor e cliente, para acessibilidade (ligar
`<label htmlFor>` a `<input id>`):
```jsx
const id = useId();
<label htmlFor={id}>Nome</label><input id={id} />
```
**Não** use para keys de lista (use IDs dos dados → [[react-rendering-lists]]).

## useImperativeHandle
Expõe um objeto imperativo customizado via `ref` (com `forwardRef`, ou `ref` como
prop no React 19):
```jsx
useImperativeHandle(ref, () => ({ focus() { inputRef.current.focus(); } }), []);
```
Use com parcimônia — prefira props declarativas.

## useLayoutEffect
Como `useEffect`, mas roda **sincronamente após o DOM mutar e antes do paint**.
Para **medir layout** (posição/tamanho) e ajustar antes do usuário ver:
```jsx
useLayoutEffect(() => { const h = ref.current.offsetHeight; setHeight(h); });
```
⚠️ Bloqueia o paint → pode prejudicar performance; só quando precisa medir/evitar flicker.

## useDebugValue
Adiciona um label a um custom Hook no React DevTools: `useDebugValue(isOnline ? 'Online' : 'Offline')`.

## useSyncExternalStore
Assina uma **store externa** (fora do React) de forma segura para concorrência e SSR:
```jsx
const value = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
```
Use em libs de state/stores do browser em vez de `useEffect` + `setState`.

## use
Lê o valor de uma **Promise** (com Suspense) ou de um **Context** — e, diferente
dos outros Hooks, **pode ser chamado em condições/loops**:
```jsx
const theme = use(ThemeContext);
const data = use(promise); // suspende até resolver (dentro de <Suspense>)
```

## Armadilhas

- `useId` para keys de lista → errado; é para IDs de acessibilidade.
- `useLayoutEffect` no servidor (SSR) avisa — só roda no cliente; use `useEffect` se não precisa medir.
- `useImperativeHandle` virando muleta imperativa → reavalie o design declarativo.
- `use(promise)` precisa de uma Promise estável (criada no server/cacheada), senão
  recria a cada render; e exige `<Suspense>` acima.
- `useSyncExternalStore`: `getSnapshot` deve retornar valor **cacheado/estável**
  (mesma referência se nada mudou), senão loop de render.

## Verificação

- Hook certo para o caso? `useLayoutEffect` só quando mede layout?
- `getSnapshot` estável? `use(promise)` com Promise estável e Suspense acima?

Relacionados: [[react-refs]], [[react-effects]], [[react-suspense-lazy]], [[react-custom-hooks]].
