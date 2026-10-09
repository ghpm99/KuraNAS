---
name: react-suspense-lazy
description: >-
  Ensina <Suspense> e lazy em React: <Suspense fallback={...}> mostra um fallback
  enquanto os filhos "suspendem" (carregando dados ou código), lazy(() =>
  import('./X')) para code-splitting de componentes, fronteiras de Suspense
  aninhadas, revelar conteúdo junto, e integração com transições. Cobre as refs
  oficiais react.dev/reference/react/Suspense e /lazy. Use quando o usuário disser
  "Suspense", "fallback de loading", "React.lazy", "code splitting", "carregar
  componente sob demanda", "lazy import", ou "mostrar spinner enquanto carrega".
when_to_use: >-
  Disparar ao usar <Suspense>/lazy, code-splitting, ou fallback de loading
  declarativo. NÃO usar para loading manual com useState/useEffect, nem para
  loading.js do Next (nextjs-loading-streaming).
---

# React: Suspense e lazy

Fontes: react.dev/reference/react/Suspense, react.dev/reference/react/lazy

## <Suspense>

Mostra um **fallback** enquanto os filhos ainda estão carregando (dados via fonte
compatível com Suspense, ou código via `lazy`):

```jsx
<Suspense fallback={<Spinner />}>
  <Albums />
</Suspense>
```

- Quando algo dentro **suspende**, o React mostra o `fallback` mais próximo acima.
- Quando o conteúdo fica pronto, o React troca o fallback pelos filhos.
- Fronteiras **aninhadas** revelam a UI em etapas; conteúdo dentro de uma mesma
  fronteira aparece **junto** (tudo ou nada).

## lazy — code splitting

```jsx
import { lazy, Suspense } from 'react';
const MarkdownPreview = lazy(() => import('./MarkdownPreview.js'));

<Suspense fallback={<Loading />}>
  <MarkdownPreview />
</Suspense>
```

`lazy(load)` recebe uma função que retorna uma **Promise de módulo** com `default`
sendo um componente React. O bundle só é baixado quando o componente é renderizado.

## Integração com transições

Durante uma transição ([[react-concurrent-hooks]]), o React **evita esconder**
conteúdo já visível com o fallback — em vez disso espera ou mantém o anterior. Bom
para navegação sem "piscar" spinner.

## Armadilhas

- **Declare `lazy` fora do componente** (no top level do módulo). Dentro do render,
  recria o componente lazy a cada render e perde state/recarrega.
- `Suspense` só captura fontes de dados **compatíveis com Suspense** (frameworks,
  `use(promise)`, libs como Relay/TanStack). `fetch` cru no `useEffect` **não**
  aciona Suspense — continue com state de loading manual.
- O módulo de `lazy` precisa ter **`export default`** do componente.
- Sem fronteira `<Suspense>` acima de algo que suspende → erro/propaga até a raiz.

## Verificação

- `lazy()` no top level? Componente tem default export?
- Há `<Suspense fallback>` acima de tudo que pode suspender?

Relacionados: [[react-concurrent-hooks]], [[react-utility-hooks]], [[nextjs-loading-streaming]].
