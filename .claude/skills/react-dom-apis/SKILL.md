---
name: react-dom-apis
description: >-
  Ensina as APIs de react-dom: createPortal(children, domNode) para renderizar
  filhos em outro nó do DOM (modais, tooltips) mantendo a árvore React; createRoot
  e hydrateRoot (entrypoint do client, montar/hidratar a app); flushSync para
  forçar atualização síncrona do DOM; e as APIs de pré-carregamento de recursos
  (preload, preinit, prefetchDNS, preconnect). Cobre react.dev/reference/react-dom.
  Use quando o usuário disser "createPortal", "modal/tooltip que escapa do
  container", "createRoot", "hydrateRoot", "flushSync", "preload de recurso", ou
  "renderizar fora da hierarquia do DOM".
when_to_use: >-
  Disparar ao usar portal, montar/hidratar a app, forçar flush síncrono, ou
  pré-carregar recursos. NÃO usar para refs/DOM em geral (react-refs) nem
  componentes <Form>/<Image> do Next (nextjs-built-in-components).
---

# React: APIs de react-dom

Fonte: react.dev/reference/react-dom

## createPortal — renderizar em outro nó do DOM

```jsx
import { createPortal } from 'react-dom';

return (
  <div>
    <p>Fica aqui</p>
    {createPortal(<ModalContent />, document.body)}
  </div>
);
```

Os filhos vão para `domNode` (ex: `document.body`), mas **continuam na árvore
React** — context, eventos (bubbling sobe pela árvore React, não pelo DOM) e state
funcionam normalmente. Ideal para **modais, tooltips, menus** que precisam escapar
de `overflow:hidden`/`z-index` do container.

## createRoot / hydrateRoot — entrypoint do client

```jsx
import { createRoot } from 'react-dom/client';
const root = createRoot(document.getElementById('root'));
root.render(<App />);
```

- `createRoot(domNode).render(<App/>)`: monta a app numa página client-rendered.
- `hydrateRoot(domNode, <App/>)`: **hidrata** HTML já renderizado no servidor (SSR),
  anexando os event listeners ao markup existente. (No Next.js isso é gerenciado
  pelo framework — você raramente chama direto.)

## flushSync — atualização síncrona

```jsx
import { flushSync } from 'react-dom';
flushSync(() => { setCount(c => c + 1); });
// DOM já atualizado aqui (ex: para então medir/scrollar)
```

Raro — força o React a aplicar o update e o DOM **sincronamente**, fora do batching.
Prejudica performance; use só quando precisa do DOM atualizado imediatamente.

## Pré-carregamento de recursos

Dão dicas ao browser para baixar recursos cedo:
- `preload(href, options)` — baixa um recurso (fonte, css, imagem) que será usado.
- `preinit(href, options)` — baixa **e executa/insere** (script/stylesheet).
- `prefetchDNS(href)` — resolve DNS de um domínio que você vai usar.
- `preconnect(href)` — abre conexão antecipada com um domínio.

## Armadilhas

- Portal muda o **DOM parent**, não o **React parent**: o bubbling de eventos sobe
  pela árvore React (o pai do JSX continua "ouvindo"), o que costuma surpreender.
- Esquecer de tratar foco/acessibilidade/`Esc` em modais via portal.
- `hydrateRoot` exige que o HTML do cliente **bata** com o do servidor → mismatch
  causa warning e re-render (cuidado com `Date.now()`, `window`, valores aleatórios no render).
- `flushSync` em excesso → mata a performance (perde batching).

## Verificação

- Modal/tooltip usa `createPortal`? Foco/Esc/aria tratados?
- Hidratação sem mismatch (nada de valores client-only no primeiro render)?

Relacionados: [[react-refs]], [[react-suspense-lazy]], [[nextjs-built-in-components]].
