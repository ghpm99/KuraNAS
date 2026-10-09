---
name: react-strict-mode
description: <StrictMode> do React — ativa checagens extras só em desenvolvimento numa subárvore para revelar bugs: re-executa componentes, inicializadores de state e reducers, remonta Effects (setup→cleanup→setup) e avisa sobre APIs depreciadas. Cobre "StrictMode" de react.dev. TRIGGER ao envolver a app em <StrictMode>, investigar por que um componente/Effect roda duas vezes em dev, ou caçar render impuro e cleanup faltando. SKIP para as APIs que montam a raiz (react-client-apis) e para o conceito de pureza em si (react-pure-components).
metadata:
  type: skill
  source: react.dev
---

# React StrictMode (<StrictMode>)

## Regra canônica
`<StrictMode>` é um componente (de `react`) que liga **checagens extras só em desenvolvimento** para a árvore dentro dele. **Não renderiza UI e não tem efeito em produção** — serve para *expor* bugs cedo.

```jsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <App />
  </StrictMode>
);
```

## O que ele faz (em dev)
1. **Double render**: re-executa o corpo do componente, os inicializadores de `useState` e as funções reducer **duas vezes** — para revelar render impuro (se o resultado muda, há efeito colateral no render).
2. **Double Effect**: remonta cada Effect logo após montar (**setup → cleanup → setup**) — para revelar **cleanup faltando**.
3. **Avisos** sobre APIs depreciadas.

Tudo isso roda **apenas em desenvolvimento**; em produção nada é duplicado.

## "Por que meu código roda duas vezes?"
É o StrictMode, **de propósito**. Se duplicar quebra algo, o bug está no seu código (impureza ou cleanup ausente), não no StrictMode. Corrija a causa — não remova o StrictMode.

## Escopo
Pode envolver a app inteira (comum) ou só uma parte:
```jsx
<>
  <Header />
  <StrictMode>
    <main><Profile /></main>
  </StrictMode>
</>
```

## Erros comuns
- ❌ Remover `<StrictMode>` porque "roda duas vezes". ✅ Trate a duplicação como diagnóstico: torne o render puro e adicione cleanup.
- ❌ Achar que afeta produção/performance. ✅ É dev-only, sem custo em prod.
- ❌ "Consertar" a duplicação com flag/ref para rodar só uma vez. ✅ Conserte a impureza/cleanup real.

## Checklist
- [ ] `<StrictMode>` envolve a app (ou a subárvore que você quer checar).
- [ ] Componente/inicializadores/reducers são puros (sobrevivem ao double render).
- [ ] Todo Effect tem cleanup que desfaz o setup (sobrevive ao double mount).
- [ ] Não há gambiarra para suprimir a re-execução em dev.
