---
name: react-form-actions
description: >-
  Ensina Actions e os hooks de formulário do React: useActionState(action,
  initialState) → [state, formAction, isPending] usado em <form action={formAction}>
  com action (prevState, formData) => newState; useFormStatus() (de react-dom) que
  lê {pending, data, method, action} do <form> pai; e useOptimistic(state, updateFn)
  → [optimisticState, addOptimistic] para UI otimista durante actions async. Cobre
  as refs oficiais react.dev/reference/react/useActionState, /useOptimistic e
  react-dom/.../useFormStatus. Use quando o usuário disser "form action", "Server
  Action no form", "useActionState", "useFormStatus", "botão de submit pending",
  "useOptimistic", "UI otimista", ou "estado de envio do formulário".
when_to_use: >-
  Disparar ao usar <form action={...}>, gerenciar pending/erro de submit, ou UI
  otimista. NÃO usar para useTransition genérico (react-concurrent-hooks) nem para
  eventos onClick comuns (react-handling-events).
---

# React: Actions e hooks de formulário

Fontes: react.dev/reference/react/useActionState,
react.dev/reference/react/useOptimistic,
react.dev/reference/react-dom/hooks/useFormStatus

**Actions**: funções (sync ou async) passadas a `<form action={fn}>` ou a hooks,
que o React executa gerenciando pending, erros e otimismo. Combinam com Server
Functions no Next.js ([[nextjs-server-actions]]).

## useActionState

```jsx
const [state, formAction, isPending] = useActionState(
  async (previousState, formData) => {
    const error = await submit(formData);
    return error ?? { ok: true }; // vira o novo `state`
  },
  null // initialState
);

return (
  <form action={formAction}>
    <input name="email" />
    <button disabled={isPending}>Enviar</button>
    {state?.error && <p>{state.error}</p>}
  </form>
);
```

A action recebe `(previousState, formData)` e o retorno vira o próximo `state`.
`isPending` cobre o envio.

## useFormStatus (react-dom)

Lê o status do `<form>` **pai** — útil num botão de submit reutilizável:

```jsx
import { useFormStatus } from 'react-dom';
function SubmitButton() {
  const { pending, data, method, action } = useFormStatus();
  return <button disabled={pending}>{pending ? 'Enviando…' : 'Enviar'}</button>;
}
```

⚠️ **Deve estar num componente renderizado DENTRO do `<form>`** (não no mesmo
componente que renderiza o `<form>`).

## useOptimistic

Mostra um estado otimista enquanto a action async não termina:

```jsx
const [optimisticMsgs, addOptimistic] = useOptimistic(
  messages,
  (state, newMsg) => [...state, { text: newMsg, sending: true }]
);
async function send(formData) {
  addOptimistic(formData.get('text')); // UI atualiza já
  await sendMessage(formData);          // realidade chega depois
}
```

## Armadilhas

- `useFormStatus` no **mesmo** componente do `<form>` → retorna sempre `pending:false`.
  Extraia o botão para um filho.
- Action de form recebe `(prevState, formData)` em `useActionState` — não confunda
  com a assinatura `(formData)` de uma action passada direto a `<form action>`.
- O estado otimista é **temporário**: quando a action resolve e o state real
  atualiza, o otimista é descartado.
- Esquecer `name` nos inputs → `formData.get(...)` vem vazio.

## Verificação

- A action retorna o próximo state? O botão de pending está dentro do `<form>`?
- Inputs têm `name`? O estado otimista some quando o real chega?

Relacionados: [[react-concurrent-hooks]], [[react-handling-events]], [[nextjs-server-actions]].
