---
name: react-client-apis
description: APIs de entrada do react-dom/client — createRoot(domNode) para montar uma app React só-cliente (root.render / root.unmount) e hydrateRoot(domNode, jsx) para hidratar HTML já renderizado no servidor. Cobre "createRoot" e "hydrateRoot" de react.dev. TRIGGER ao montar o React num <div id="root">, escrever o entry point (main.jsx/index.js), escolher createRoot vs hydrateRoot, configurar onUncaughtError/onCaughtError/onRecoverableError, ou depurar hydration mismatch. SKIP quando um framework (Next.js) já monta a raiz por você, e para <StrictMode> em si (react-strict-mode).
metadata:
  type: skill
  source: react.dev
---

# React Client APIs (createRoot / hydrateRoot)

## Regra canônica
`react-dom/client` expõe os dois pontos de entrada que ligam o React a um nó do DOM:
- **`createRoot`** — para apps **renderizadas só no cliente** (CSR).
- **`hydrateRoot`** — para anexar o React a HTML **já renderizado no servidor** (SSR/SSG).

Você chama uma dessas **uma vez**, no entry point. Em frameworks (Next.js) isso já é feito por você.

## createRoot
```jsx
import { createRoot } from 'react-dom/client';

const root = createRoot(document.getElementById('root'));
root.render(<App />);
// depois, opcionalmente:
root.unmount();
```
- `createRoot(domNode, options?)` retorna um objeto root com `render` e `unmount`.
- `render` monta/atualiza; `unmount` destrói e limpa.
- O `domNode` deve estar **vazio** — o React controla todo o conteúdo dele.

## hydrateRoot
```jsx
import { hydrateRoot } from 'react-dom/client';

hydrateRoot(document.getElementById('root'), <App />);
```
- `hydrateRoot(domNode, reactNode, options?)` reaproveita o HTML do servidor, anexando os event listeners — não recria o DOM.
- A árvore do cliente **precisa bater** com o HTML do servidor; divergência gera *hydration mismatch*.

## createRoot vs hydrateRoot
| | createRoot | hydrateRoot |
|---|---|---|
| Conteúdo inicial do DOM | vazio | HTML do servidor |
| Ação | renderiza do zero | hidrata o existente |
| Uso | app só-cliente | SSR/SSG |

## Opções de erro (ambas)
- `onUncaughtError` — erro não capturado por nenhum error boundary.
- `onCaughtError` — erro capturado por um error boundary.
- `onRecoverableError` — erro do qual o React se recuperou (ex.: hydration mismatch).
- `identifierPrefix` — prefixo dos IDs gerados (`useId`), útil com múltiplas roots na página.

## Erros comuns
- ❌ Usar `createRoot` em HTML vindo do servidor. ✅ Use `hydrateRoot` para SSR/SSG.
- ❌ Render diferente no servidor e no cliente (datas, `Math.random`, `window`). ✅ Mantenha o 1º render idêntico; difira só **após** montar.
- ❌ Chamar `createRoot` a cada atualização. ✅ Crie a root uma vez; use `root.render` para atualizar.
- ❌ Reescrever isso num projeto Next.js. ✅ O framework já monta a raiz (ver skills `nextjs-*`).

## Checklist
- [ ] É um app sem framework (senão o framework monta a raiz).
- [ ] CSR puro → `createRoot`; HTML do servidor → `hydrateRoot`.
- [ ] A root é criada **uma vez** no entry point.
- [ ] Sem hydration mismatch (render server == client no 1º paint).
