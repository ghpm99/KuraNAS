---
name: react-handling-events
description: >-
  Ensina event handlers em React: adicionar onClick={handleClick} (passar a
  função, NÃO chamá-la com ()), handlers inline, ler props no handler, passar
  handlers como props (convenção on*/handle*), propagação/bubbling de eventos,
  e.stopPropagation() e e.preventDefault(). Cobre "Responding to Events" de
  react.dev/learn. Use quando o usuário disser "onClick não funciona",
  "handler dispara no render", "como passo função pro onClick",
  "stopPropagation", "preventDefault", "evento dispara no pai também", ou "onSubmit".
when_to_use: >-
  Disparar ao escrever onClick/onSubmit/onChange, quando um handler executa no
  render (chamado com ()), ou ao lidar com bubbling/default do navegador. NÃO usar
  para useState em si (react-state-usestate).
---

# React: Respondendo a eventos

Fonte oficial: https://react.dev/learn/responding-to-events

```jsx
export default function Button() {
  function handleClick() {
    alert('You clicked me!');
  }
  return <button onClick={handleClick}>Click me</button>;
}
```

Convenção: handlers começam com `handle` + evento. Podem ser inline:
`<button onClick={() => alert('...')}>`.

## ⚠️ Passar a função, NÃO chamar

| ✅ Correto | 🔴 Errado (executa no render) |
|---|---|
| `onClick={handleClick}` | `onClick={handleClick()}` |
| `onClick={() => alert('x')}` | `onClick={alert('x')}` |

`handleClick()` com parênteses **executa imediatamente durante o render**, não no clique.

## Ler props e passar handlers como props

```jsx
function AlertButton({ message, children }) {
  return <button onClick={() => alert(message)}>{children}</button>;
}

// Pai passa o handler:
function Button({ onClick, children }) {
  return <button onClick={onClick}>{children}</button>;
}
```

Convenção para props de handler custom: `on` + Maiúscula (`onSmash`, `onPlay`).

## Propagação (bubbling)

Eventos sobem pela árvore: clicar no filho dispara também handlers dos pais.

- `e.stopPropagation()` — impede o evento de subir para os pais.
- `e.preventDefault()` — impede a ação padrão do navegador (ex: `<form onSubmit>` recarregar).

```jsx
<button onClick={e => { e.stopPropagation(); onClick(); }}>...</button>

<form onSubmit={e => { e.preventDefault(); /* enviar */ }}>...</form>
```

`onScroll` é o **único** evento que não propaga.

## Armadilhas

- **Handler dispara no render / loop infinito** → você chamou a função (`onClick={fn()}`)
  em vez de passá-la (`onClick={fn}`). Para passar args, use arrow: `onClick={() => fn(arg)}`.
- Confundir `stopPropagation` (bubbling) com `preventDefault` (ação do navegador).
- `<form>` recarrega a página ao submeter → faltou `e.preventDefault()`.
- Side effects são bem-vindos em handlers (diferente do render) — é o lugar certo
  para `setState`.

## Verificação

- Handlers passados sem `()`? Args via arrow function?
- `preventDefault` em submit de form? `stopPropagation` só quando realmente precisa?

Relacionados: [[react-state-usestate]], [[react-pure-components]], [[react-form-actions]].
