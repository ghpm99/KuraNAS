# 39 — Android: navegação com barra inferior e menu "Mais"

**Tipo:** bug + reorganização · **Prioridade:** P1 · **Spec:** [24](../especificacao/24-app-android.md) `RN-06`, `RN-07`, `RN-10`, `RN-15`, `RF-03` a `RF-06`, `CA-06`, `CA-07`, `CA-10`, `CA-12`, `D-01`

## Contexto

Situação descrita na spec 24 §8, levantada por leitura de código em 2026-11-10 — **reconferir cada ponto no código antes de começar** e corrigir esta lista no primeiro commit se algo divergir:

- A navegação são onze páginas deslizáveis lado a lado.
- A tela de Tarefas existe e não é alcançável.
- "Esquecer servidor" na gaveta lateral não faz nada; o da tela de Configurações funciona.
- Dispositivo fora da whitelist não vê explicação nem o próprio IP.
- Sem conexão, as telas falham cada uma do seu jeito; não há aviso persistente.

## Objetivo

Toda tela do app é alcançável em no máximo dois toques, nenhum item de menu fica sem efeito, e o app diz claramente quando está sem servidor ou sem acesso.

## O que fazer

1. **Barra inferior** com cinco destinos: Início, Arquivos, Fotos, Música, Vídeos (`RF-03`), no lugar das páginas deslizáveis. Cada destino guarda sua própria pilha e sua posição de rolagem.
2. **Menu "Mais"**: Documentos, Favoritos, Busca, Transferências, Lixeira, Notificações, Tarefas, Baixar de URL, Diário, Configurações. Item cuja tela ainda não existe (Documentos — task 46; Baixar de URL — item 09 do board de revisão) **não aparece** até a tela existir. Transferências já existe (task 37).
3. **Tarefas alcançável** pelo menu (`RN-15`).
4. **"Trocar servidor"** funciona em todos os lugares em que aparece (`RN-07`): uma única ação compartilhada, que esquece o servidor e volta à tela de conexão.
5. **Busca** no topo das telas principais (`RF-04`) e **contador de não lidas** no acesso às notificações (`RF-06`).
6. **Aviso de sem conexão** (`RN-10`): faixa persistente quando o servidor lembrado não responde; o app continua navegável; ao reconectar, a tela visível recarrega.
7. **Fora da whitelist** (`RN-06`): resposta `403` do servidor leva a uma tela que explica e mostra o IP do aparelho para o dono cadastrar.
8. **Voltar do sistema** (`RN-14`): volta à tela anterior; no destino inicial, sai do app.

## Reconferência no código (2026-11-24)

Os cinco pontos do Contexto se confirmam. Além deles: o voltar do sistema em qualquer página do pager fecha o app; dentro de uma subpasta de Arquivos o voltar sai do destino inteiro; o `403` só é tratado na reconexão inicial e o IP do corpo (`{"error","ip"}`) é descartado; a contagem de não lidas tem API no app (`notifications/unread-count`) sem nenhum chamador.

## Decisões de desenho (2026-11-24)

- **Barra inferior** com grafo aninhado por destino e `saveState`/`restoreState`; o pager e a gaveta saem. "Mais" é um sexto item da barra que abre uma tela de lista (não gaveta), com indicador de transferências ativas e de notificações não lidas (`RF-06`).
- **Busca** vira rota própria, aberta pelo ícone no topo das cinco telas da barra (`RF-04`) e também listada no "Mais".
- **Trocar servidor**: uma ação única (`ServerStore.clear()` atrás de um caso de uso) com o mesmo diálogo de confirmação, usada em Configurações e onde mais aparecer.
- **Sem conexão** (`RN-10`): monitor na raiz que combina o callback de rede do sistema com `/api/v1/health` do servidor lembrado; faixa persistente acima da barra; ao voltar, um evento de reconexão faz a tela visível recarregar sem perder a rolagem.
- **Fora da whitelist** (`RN-06`): interceptor que, num `403` com `ip` no corpo, publica o evento; a raiz mostra a tela que explica e exibe o IP, com "Tentar de novo" e "Trocar servidor". A reconexão inicial passa a carregar o IP no `AccessDenied`.

## Critérios de aceite

Verificados em aparelho real, em celular e em tablet.

- [ ] `24/CA-12` — todas as telas do menu abrem; nenhum item fica sem efeito.
- [ ] `24/CA-07` — "Trocar servidor" volta à tela de conexão a partir de todos os lugares em que aparece.
- [ ] `24/CA-06` — em aparelho fora da whitelist, o app mostra que o acesso não está liberado e o IP do aparelho.
- [ ] `24/CA-10` — desligar o Wi-Fi com o app aberto mostra o aviso sem fechar o app; religar recarrega a tela.
- [ ] Trocar de destino na barra e voltar mantém a lista e a rolagem do destino anterior.
- [ ] `24/CA-13` — em tablet de 10" e em celular de 5", nada cortado.
- [ ] Textos em `strings.xml` (pt-BR e en-US).
- [ ] `assembleDebug` e `testDebugUnitTest` verdes.

## Fora de escopo

- Aplicar tema e preferências de players vindos do servidor (`24/RN-13`; item 16 do board de revisão).
- Interface para TV (`24/D-02`).
- Início com falha por bloco (`14/RN-01`; item 18 do board de revisão).

## Depende de

Task 22.
