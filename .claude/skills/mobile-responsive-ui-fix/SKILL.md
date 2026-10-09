---
name: mobile-responsive-ui-fix
description: "Corrige problemas de responsividade em telas mobile e tablet: overflow horizontal, componentes cortados, texto ilegível, botões inacessíveis, tabelas que não cabem na tela, e layouts quebrados em viewports pequenos. Use quando o usuário disser 'tela quebrada no celular', 'não aparece no mobile', 'tá cortando no tablet', 'responsividade ruim', 'layout quebrado em tela pequena', ou após implementar uma tela nova sem testar em mobile."
---

# /mobile-responsive-ui-fix

Diagnostica e corrige problemas de responsividade. Foca nos padrões mais comuns encontrados em telas React Native / React + Tailwind / Flutter: overflow, texto fixo, tabelas horizontais e touch targets pequenos.

## Quando invocar

- "Tela quebrada no celular / tablet"
- "Componente não aparece em mobile"
- "Está cortando na tela pequena"
- Após implementar tela nova sem testar responsividade

## Procedimento

### 1. Identificar a plataforma e viewport afetados

Pergunte (se não informado):
- É React Native, React Web (Tailwind/CSS), ou Flutter?
- Em quais viewports quebra? (iPhone SE 375px? iPad 768px? apenas landscape?)
- Qual componente ou rota específica?

### 2. Inspecionar a tela com foco em responsividade

Leia o arquivo do componente e procure:

**React Web (Tailwind):**
```
grep -n "w-\[" {arquivo}   # larguras fixas em px
grep -n "overflow-hidden\|overflow-x" {arquivo}
grep -n "flex\b" {arquivo} # flex sem flex-wrap
grep -n "min-w\|max-w" {arquivo}
```

**React Native:**
```
grep -n "width:" {arquivo}   # larguras fixas
grep -n "height:" {arquivo}  # alturas fixas
grep -n "fontSize:" {arquivo} # tamanhos de fonte fixos
```

### 3. Checklist de fixes comuns

**Overflow horizontal:**
- Substitua larguras fixas (`w-[400px]`) por relativas (`w-full`, `max-w-xl`)
- Adicione `overflow-x-auto` em wrappers de tabela
- Troque `flex` sem wrap por `flex flex-wrap` ou `grid`

**Tabelas em mobile:**
- Envolva em `<div className="overflow-x-auto">` ou use layout de cards em `sm:`
- Esconda colunas menos importantes: `hidden sm:table-cell`

**Texto ilegível:**
- Remova `text-[10px]` ou `fontSize: 10` fixo — use escala relativa
- Confirme que não há `numberOfLines={1}` truncando conteúdo importante (RN)

**Touch targets pequenos:**
- Botões devem ter mínimo 44×44px (`min-h-[44px] min-w-[44px]`)
- Adicione `py-3 px-4` se o botão for muito comprimido

**Layout quebrado:**
- Troque `absolute` com coordenadas fixas por `relative` + flexbox quando possível
- Em RN, use `Dimensions.get('window')` com cuidado — prefira `flex: 1`

### 4. Verificar em múltiplos breakpoints

Após o fix, confirme visualmente ou por inspeção de código que funciona em:
- 375px (iPhone SE / small Android)
- 414px (iPhone Plus)
- 768px (tablet portrait)
- 1024px (tablet landscape / desktop mínimo)

## Armadilhas

- **`overflow-hidden` no pai** quebra scroll interno — verifique a hierarquia de overflow antes de adicionar `overflow-x-auto`.
- **`position: absolute` com top/left fixos** nunca funciona em mobile — investigar se não há um `relative` pai faltando.
- Em React Native, **`Dimensions.get('window').width`** não reage a rotação sem `useWindowDimensions()`.
- Tailwind classes condicionais (`sm:`, `md:`) não funcionam quando geradas dinamicamente via template string — devem estar literais no código.
- Não "force" responsividade escondendo o componente com `hidden` em mobile sem oferecer alternativa acessível.

## Verificação

- Nenhum scroll horizontal indesejado em 375px
- Todos os botões com área de toque ≥ 44px
- Texto legível sem zoom (≥ 14px / `text-sm`)
- Nenhum conteúdo cortado ou sobrepostos
