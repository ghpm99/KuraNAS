---
name: react-refs
description: >-
  Ensina refs em React: useRef para guardar valores que NÃO disparam re-render e
  persistem entre renders (ref.current mutável), ref vs state, quando usar (IDs de
  timeout, nós do DOM), não ler/escrever ref.current durante o render; e manipular
  o DOM (ref={myRef} → myRef.current é o nó, focus/scroll, ref como prop/forwardRef,
  flushSync, não mexer no DOM gerido pelo React). Cobre "Referencing Values with
  Refs" e "Manipulating the DOM with Refs" de react.dev/learn. Use quando o usuário
  disser "useRef", "ref no input", "focar um input", "guardar valor sem
  re-renderizar", "acessar elemento do DOM", "forwardRef", ou "ref.current".
when_to_use: >-
  Disparar ao usar useRef, acessar nós do DOM, ou guardar valor mutável que não
  afeta render. NÃO usar para valores exibidos na UI (react-state-usestate) nem para
  sincronizar com sistemas externos (react-effects).
---

# React: Refs (valores e DOM)

Fontes: react.dev/learn/referencing-values-with-refs,
react.dev/learn/manipulating-the-dom-with-refs

`useRef` deixa o componente "lembrar" de informação **sem disparar re-render**.

```jsx
const ref = useRef(0);     // { current: 0 }
ref.current = ref.current + 1; // muta direto, sem re-render
```

## Ref vs State

| | Ref | State |
|---|---|---|
| Hook | `useRef(v)` → `{current: v}` | `useState(v)` → `[v, setV]` |
| Re-render ao mudar | ❌ não | ✅ sim |
| Mutabilidade | mutável (`.current = ...`) | "imutável" (via setter) |
| Ler durante render | ❌ evite | ✅ ok |

**Regra:** se a informação é usada para **renderizar**, use state. Se só é usada por
handlers e não precisa re-render, use ref.

Bons usos: IDs de `setTimeout`/`setInterval` para limpar, nós do DOM, valores que
não entram na renderização.

## Manipular o DOM

```jsx
const inputRef = useRef(null);
function handleClick() { inputRef.current.focus(); }
return (<><input ref={inputRef} /><button onClick={handleClick}>Focus</button></>);
```

React põe o nó DOM em `ref.current`; ao remover o elemento, volta a `null`.

- **Ref para um filho**: passe `ref` como prop. Em React 19 funciona direto; em
  versões anteriores use `forwardRef((props, ref) => <input ref={ref} />)`.
- **flushSync** (`react-dom`): aplica o DOM de forma síncrona quando você precisa
  ler o DOM logo após um setState (raro).
- **Não altere o DOM gerido pelo React** (nós que ele renderiza) — só manipule DOM
  que o React não controla.

## Armadilhas

- Ler `ref.current` durante o render → comportamento imprevisível. Exceção: init
  única `if (!ref.current) ref.current = new X()`.
- Usar ref para valor exibido (`<button>{ref.current}</button>`) → não atualiza na tela.
- Abusar de refs em vez de state/props → sinal de design ruim (refs são escape hatch).
- Acessar `ref.current` antes de montar (no primeiro render é `null`).

## Verificação

- O valor afeta a UI? Então é state, não ref.
- Não lê/escreve `.current` no corpo do render? Acessa o DOM só em handlers/effects?

Relacionados: [[react-state-usestate]], [[react-effects]], [[react-dom-apis]].
