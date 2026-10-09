---
name: react-scroll-layout-debug
description: >-
  Diagnostica e corrige scroll automático indesejado em páginas React: scroll na
  carga da página, scroll para um card/elemento específico sem ação do usuário,
  página que pula para meio do conteúdo. TRIGGER quando: usuário disser "a página
  está fazendo scroll automático", "ao carregar a página rola para X", "scroll
  indesejado", "a tela está pulando para baixo". SKIP: scroll proposital
  (navigate com hash, smooth scroll em botão), problemas de layout sem scroll.
---

# React Scroll Layout Debug

## Causas mais comuns

### 1. `useEffect` com `scrollIntoView` ou `window.scrollTo`

```tsx
// PROBLEMA: scroll disparado na montagem do componente
useEffect(() => {
  ref.current?.scrollIntoView(); // rola para o elemento ao montar
}, []); // dependência vazia = roda na montagem

// FIX: adiciona guard para só fazer scroll quando for intencional
useEffect(() => {
  if (shouldScroll) {
    ref.current?.scrollIntoView({ behavior: "smooth" });
  }
}, [shouldScroll]);
```

### 2. `IntersectionObserver` disparando callback na montagem

```tsx
// PROBLEMA: observer chama callback imediatamente quando elemento entra na viewport
const observer = new IntersectionObserver((entries) => {
  entries.forEach((entry) => {
    if (entry.isIntersecting) {
      // lógica que chama scrollIntoView ou setState que causa scroll
    }
  });
});

// FIX: ignora a primeira entrada (montagem) com um ref de controle
const hasInteractedRef = useRef(false);
const observer = new IntersectionObserver((entries) => {
  entries.forEach((entry) => {
    if (entry.isIntersecting && hasInteractedRef.current) {
      // ...
    }
  });
});
```

### 3. `useLayoutEffect` com scroll

`useLayoutEffect` roda antes do paint mas depois do DOM estar pronto — perfeito
para triggers de scroll indesejados.

```tsx
// PROBLEMA
useLayoutEffect(() => {
  containerRef.current?.scrollTo(0, savedScrollPosition);
}, []); // restaura posição de scroll mas de forma indesejada

// FIX: move para useEffect com condição ou remove
```

### 4. Elemento com `autofocus` ou `tabIndex` que recebe foco

Elementos que recebem foco automaticamente causam scroll do browser para
colocá-los na viewport.

```tsx
// PROBLEMA
<input autoFocus />
<div tabIndex={0} ref={cardRef}>  {/* recebe foco via JS */}

// FIX: remove autoFocus ou previne o scroll do focus
element.focus({ preventScroll: true });
```

### 5. Hash na URL que o browser interpreta como âncora

```tsx
// PROBLEMA: URL /internal/financial#emergency-fund causa scroll para #emergency-fund
// ou o React Router está restaurando hash navigation

// Verifica se a URL tem hash
console.log(window.location.hash);

// FIX: se não quiser este comportamento, limpa o hash na montagem
useEffect(() => {
  if (window.location.hash) {
    history.replaceState(null, "", window.location.pathname);
  }
}, []);
```

### 6. Biblioteca de carousel/slider causando scroll

Componentes como Embla, Swiper, Splide às vezes chamam `scrollIntoView`
internamente.

## Diagnóstico passo a passo

### 1. Encontra o elemento que recebe o scroll

```js
// Cola no console do browser
const origScrollIntoView = Element.prototype.scrollIntoView;
Element.prototype.scrollIntoView = function(...args) {
  console.trace("scrollIntoView chamado em:", this);
  return origScrollIntoView.apply(this, args);
};

const origScrollTo = window.scrollTo;
window.scrollTo = function(...args) {
  console.trace("window.scrollTo chamado com:", args);
  return origScrollTo.apply(this, args);
};
```

Recarrega a página com o console aberto — o stack trace vai mostrar exatamente
qual código está chamando o scroll.

### 2. Grep por chamadas suspeitas no código

```bash
# Procura por chamadas de scroll no código
grep -rn "scrollIntoView\|scrollTo\|scrollTop\|scrollY" src/ \
  --include="*.tsx" --include="*.ts" | grep -v "node_modules\|.test."

# Procura por useLayoutEffect (mais provável de causar scroll no mount)
grep -rn "useLayoutEffect" src/ --include="*.tsx" --include="*.ts"

# Procura por autofocus
grep -rn "autoFocus\|autofocus" src/ --include="*.tsx"
```

### 3. Identifica o componente problemático

```bash
# Abre a página com React DevTools e profila o mount
# Ou usa o Chrome Performance tab para ver o call stack no primeiro scroll
```

### 4. Verifica se é hash na URL

```bash
# Verifica se alguma rota do React Router tem hash restoration
grep -rn "hash\|scrollRestoration" src/ --include="*.tsx" --include="*.ts"
```

## Correção definitiva

Depois de identificar a causa via console trace:

1. Localiza o arquivo e linha exata
2. Adiciona condição para só fazer scroll quando houver interação do usuário
3. Ou remove o scroll se for resquício de feature antiga

```tsx
// Padrão seguro: scroll só em resposta a ação do usuário
const handleNavigateToCard = () => {
  cardRef.current?.scrollIntoView({ behavior: "smooth", block: "center" });
};
// Não chama handleNavigateToCard em useEffect
```

## Verifica a correção

1. Abre a página no browser com DevTools aberto na aba Console
2. Confirma que o monkey-patch de `scrollIntoView` não dispara no load
3. Confirma que scroll funciona quando esperado (ex: botão "ir para X")
4. Testa em mobile viewport também (scroll behavior pode diferir)
