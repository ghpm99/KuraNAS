---
name: react-error-boundaries
description: Error boundaries em React — capturar erros de renderização da árvore abaixo e mostrar uma UI de fallback, via componente de classe com static getDerivedStateFromError (renderiza fallback) e componentDidCatch (loga). Cobre "Catching rendering errors with an error boundary" e a referência de Component de react.dev. TRIGGER ao isolar falhas de UI, mostrar fallback quando um componente quebra no render, logar erros da árvore, usar a lib react-error-boundary, ou resetar após erro. SKIP para erros em event handlers/código async (use try/catch — não são capturados) e para estado de carregamento com Suspense (react-suspense-lazy).
metadata:
  type: skill
  source: react.dev
---

# React Error Boundaries

## Regra canônica
Um error boundary é um componente que **captura erros lançados durante a renderização** (render, métodos de ciclo de vida e construtores) de qualquer componente **na árvore abaixo dele** e renderiza uma UI de fallback no lugar da subárvore que quebrou. Só **componentes de classe** podem ser error boundaries — não existe Hook equivalente no core.

## O que captura (e o que NÃO captura)
Captura:
- Erros durante o render.
- Erros em métodos de ciclo de vida.
- Erros em construtores da árvore abaixo.

NÃO captura (precisam de `try/catch` comum):
- Event handlers (`onClick` etc.).
- Código assíncrono (`setTimeout`, `fetch`, promises).
- Renderização no servidor (SSR).
- Erros lançados no próprio error boundary.

## Implementação (classe)
```jsx
class ErrorBoundary extends React.Component {
  state = { hasError: false };

  // Renderiza o fallback no próximo render
  static getDerivedStateFromError(error) {
    return { hasError: true };
  }

  // Efeito colateral: logar
  componentDidCatch(error, info) {
    logError(error, info.componentStack);
  }

  render() {
    if (this.state.hasError) return this.props.fallback;
    return this.props.children;
  }
}
```
Uso:
```jsx
<ErrorBoundary fallback={<p>Algo deu errado.</p>}>
  <Profile />
</ErrorBoundary>
```

## Onde colocar
- **Granular**: em volta de widgets independentes, para que um quebre sem derrubar a página inteira.
- **No topo**: um boundary global ("algo deu errado") como rede de segurança final.
- **Com `<Suspense>`**: o boundary captura erros; o Suspense mostra o loading.

## Reset
Para tentar de novo depois de um erro, **remonte** o boundary trocando sua `key` (ou resetando `hasError`). Mudar a `key` recria a subárvore do zero.

## Lib recomendada (react.dev)
A doc oficial sugere a lib `react-error-boundary` em vez de escrever a classe à mão:
```jsx
import { ErrorBoundary } from 'react-error-boundary';

<ErrorBoundary
  FallbackComponent={Fallback}
  onReset={() => {/* limpa o estado que causou o erro */}}
  resetKeys={[userId]}
>
  <Profile />
</ErrorBoundary>
```

## Erros comuns
- ❌ Esperar que o boundary pegue erro de `onClick`/`fetch`. ✅ Use `try/catch` nesses casos; opcionalmente jogue para o state e re-lance no render.
- ❌ Usar function component como boundary. ✅ Boundary precisa ser classe (ou use `react-error-boundary`).
- ❌ Só um boundary no topo. ✅ Use boundaries granulares para falhas isoladas.
- ❌ Sem reset → usuário fica preso no fallback. ✅ Resete via `key`/`resetKeys`.

## Checklist
- [ ] É componente de classe (ou `react-error-boundary`).
- [ ] Implementa `getDerivedStateFromError` (fallback) e/ou `componentDidCatch` (log).
- [ ] O erro é de render/ciclo de vida — não de evento/async.
- [ ] Há boundary granular onde uma falha não deve derrubar a página toda.
- [ ] Existe caminho de reset (`key`/`resetKeys`).
