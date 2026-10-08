# 02 — Shell e layout

Escopo: `frontend/src/components/layout/` (AppShell, Sidebar, Header, BottomNav, PageContainer, PageHeader, DomainPageLayout), `app/App.tsx`, `theme/`, providers globais, mini-player, notificações no header e o encaixe de cada página no sistema de layout.

## Diagnóstico (2026-10-08)

- **O shell não monta**: desde `717b3f4` o `App` importa o `AppShell` cru, sem `UIProvider`; `Sidebar` chama `useUI()` e lança. Em `/files` a sidebar ainda renderiza `FolderTree`, que exige `FileProvider` (que vive dentro da página). Os testes mockavam o shell inteiro, por isso nada pegou.
- No celular o mini-player (`position: fixed; bottom: 0; z-index 1300`) cobre o bottom nav e as toasts.
- Sidebar de 14 itens sem rolagem corta em telas baixas.
- Faltam básicos: 404 (rota `*` redireciona em silêncio), fallback de carregamento, boundary de erro por rota, aviso de servidor fora do ar, restauração de scroll, tema claro, sidebar recolhível, título da aba por rota, skip link/foco na troca de rota.
- Notificações, Assistente e Diário de atividades fora do `PageContainer`/`PageHeader`.

Decisões: preferência de tema vai para o grupo `appearance` das settings (contrato só cresce, igual ao `accent_color`); sidebar recolhida é preferência por dispositivo (`localStorage`). Sem biblioteca de virtualização nova — custo de render de listas longas tratado com `content-visibility`. Limites da busca global ficam para a funcionalidade 04.

## Lacunas

- [x] 01. Shell quebrado: `UIProvider` ausente e `FolderTree` da sidebar fora do `FileProvider`; teste de integração do `App` sem mockar o shell
- [x] 02. Mini-player cobre bottom nav/toasts no celular e a sidebar no desktop; safe-area (`viewport-fit=cover`)
- [x] 03. Sidebar sem rolagem em viewport baixo
- [x] 04. Página 404
- [x] 05. Fallback de carregamento por rota e boundary de erro dentro do shell (reseta ao navegar, mensagem i18n)
- [x] 06. Servidor fora do ar: banner global com tentar de novo, `refetchOnReconnect`, fallback de traduções
- [x] 07. Restauração/reset de scroll da área de conteúdo na navegação + voltar ao topo
- [x] 08. Tema claro/escuro/sistema
- [x] 09. Sidebar recolhível (ícones), persistida por dispositivo
- [x] 10. `document.title` por rota e `<html lang>` dinâmico
- [x] 11. Acessibilidade de navegação: skip link, foco no conteúdo ao trocar de rota, rótulos dos `<nav>`, bottom nav com links e `aria-current`, busca global como combobox
- [x] 12. Sino de notificações: largura em tela pequena, cores por token, tempo relativo traduzido, teclado, clique leva ao alvo, lista atualizada ao abrir
- [x] 13. Tela de Notificações no sistema de layout, scroll infinito, estado de erro
- [x] 14. Assistente e Diário de atividades no sistema de layout
- [ ] 15. Painéis com rolagem aninhada (preview de arquivos/favoritos, imagens, fila, playlist)
- [x] 16. Mini-player no celular: progresso e controles completos em painel expandido, alvos de 44px
- [ ] 17. `prefers-reduced-motion` do sistema operacional
- [ ] 18. Breakpoints como tokens únicos (CSS + tema MUI) e container queries
- [ ] 19. Largura "wide": grades de mídia usam a tela toda
- [x] 20. Atalhos globais (`?` ajuda, `g`+tecla para navegar)
- [x] 21. Alvos de toque < 44px (tabs de domínio, botões pequenos)
- [ ] 22. Atualização de dados: refetch ao voltar o foco para a aba
- [ ] 23. Custo de render de listas longas (`content-visibility`) nas grades/listas com scroll infinito
- [ ] 24. Código morto do shell (`useFullscreen`, `App.css`, `pages/files/files.css`, `.menuButton`, `showClock`/`useHeader`)
