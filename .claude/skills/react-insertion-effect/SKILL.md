---
name: react-insertion-effect
description: useInsertionEffect — variante de Effect para bibliotecas de CSS-in-JS injetarem tags <style> ANTES de qualquer layout effect ler o DOM; dispara antes de useLayoutEffect e useEffect, e não pode ler refs/layout nem agendar updates. Cobre "useInsertionEffect" de react.dev. TRIGGER apenas ao escrever uma lib de CSS-in-JS que injeta estilos dinâmicos e precisa evitar recálculo de layout. SKIP para efeitos de sincronização normais e para useLayoutEffect/medição de DOM (react-effects) — código de aplicação quase nunca usa esta API.
metadata:
  type: skill
  source: react.dev
---

# React useInsertionEffect

## Regra canônica
`useInsertionEffect(setup, deps?)` é uma variante de Effect destinada **exclusivamente a bibliotecas de CSS-in-JS**. Ela dispara **antes** de qualquer `useLayoutEffect` ou `useEffect`, permitindo **injetar tags `<style>` dinâmicas** no momento certo, antes do React fazer mutações que dependem de layout. **Código de aplicação não deve usá-la.**

```jsx
import { useInsertionEffect } from 'react';

// dentro de uma lib de CSS-in-JS
function useCSS(rule) {
  useInsertionEffect(() => {
    if (!isInserted(rule)) insertStyleTag(rule); // injeta <style>
  });
  return rule;
}
```

## Ordem dos efeitos (mount)
`useInsertionEffect` → `useLayoutEffect` → `useEffect`.
Injetar estilos no `useInsertionEffect` evita que um `useLayoutEffect` posterior meça o layout com o CSS ainda faltando (evita recálculo de layout).

## Limitações
- **Não pode ler refs** nem medir layout (o DOM ainda não está atualizado).
- **Não pode agendar updates de state**.
- Roda só no **cliente**, não durante SSR.

## Por que (quase) nunca usar
Para aplicar estilos, prefira:
- `<link>`/CSS estático ou CSS Modules quando possível.
- `useLayoutEffect` se precisar medir e ler o DOM.

`useInsertionEffect` só ganha quando uma lib injeta `<style>` em runtime e quer fazê-lo **antes** dos layout effects.

## Erros comuns
- ❌ Usar em componente de aplicação "por ser mais cedo". ✅ Use `useEffect`/`useLayoutEffect`; reserve `useInsertionEffect` para libs de estilo.
- ❌ Ler `ref.current`/medir layout aqui. ✅ Faça isso em `useLayoutEffect`.
- ❌ Chamar `setState` dentro dele. ✅ Não é permitido; mova para outro Effect.

## Checklist
- [ ] Você está escrevendo uma lib de CSS-in-JS (não código de app).
- [ ] O objetivo é injetar `<style>` antes dos layout effects.
- [ ] Não lê refs/layout nem agenda state updates dentro dele.
- [ ] Considerou CSS estático/`useLayoutEffect` antes de optar por esta API.
