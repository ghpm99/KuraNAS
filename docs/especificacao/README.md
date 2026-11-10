# Especificação do KuraNAS

Este diretório diz **o que o sistema deve fazer**. Cada feature tem um arquivo com objetivo, escopo, regras de negócio, requisitos por tela, contrato HTTP, estados, critérios de aceite e fora de escopo.

A especificação é **normativa**: quando o código e a spec divergem, o código está errado — ou a spec é corrigida primeiro, em commit próprio, e só então o código muda. Os boards `docs/melhorias/` e `docs/revisao-produto/` continuam sendo listas de trabalho; a partir daqui toda lacuna aponta para a regra ou o critério da spec que ela satisfaz.

## Para que o sistema existe

Central de arquivos e multimídia de uma casa, rodando em um PC usado como servidor, dentro da rede local.

- **Quem usa:** uma pessoa (o dono), em 3 computadores, 1 celular e 1 tablet.
- **Onde roda:** rede interna, sem acesso externo. Não há login; o controle de acesso é a whitelist de IPs.
- **O que precisa resolver:**

| # | Necessidade do dono | Specs que respondem |
|---|---|---|
| N1 | Tirar foto no celular, enviar para o servidor e ver no PC | [03](03-transferencia-entre-dispositivos.md), [04](04-imagens.md), [24](24-app-android.md) |
| N2 | Subir um arquivo do PC e baixar no celular (e o inverso) | [02](02-arquivos.md), [03](03-transferencia-entre-dispositivos.md) |
| N3 | Achar uma foto antiga | [04](04-imagens.md), [08](08-busca-global.md) |
| N4 | Organizar documentos e achar o documento certo quando precisar | [07](07-documentos.md), [08](08-busca-global.md) |
| N5 | Ouvir a própria coleção de músicas, organizada por gênero e em playlists que só têm o que o dono escolheu | [05](05-musica.md) |
| N6 | Assistir filmes e animes próprios, sem anúncio, continuando de onde parou | [06](06-videos.md) |
| N7 | Tudo centralizado em um servidor só, sem perder arquivo | [20](20-armazenamento.md), [21](21-indexacao.md), [10](10-lixeira.md) |

Tudo o que não serve a N1–N7 é secundário e não pode quebrar o que serve.

## Princípios de produto

1. **O básico funciona antes do avançado.** Listar, abrir, enviar, baixar, tocar e assistir vêm antes de IA, e-mail, analytics e automações. Uma feature secundária nunca justifica regressão em uma primária.
2. **Nada some.** Excluir vai para a lixeira; mover e renomear preservam o item; uma lista nunca fica vazia por causa de um erro — erro é mostrado como erro.
3. **O mesmo comportamento em toda tela.** Toda lista pagina, recarrega, ordena e falha do mesmo jeito ([00](00-padroes-transversais.md)).
4. **Rede interna, uma pessoa.** Sem login, sem multiusuário, sem compartilhamento público, sem dependência de serviço externo para o uso básico.
5. **Web e Android contam.** O celular e o tablet são clientes de primeira classe para N1, N2, N5 e N6.

## Índice

Prioridade: **P0** = necessidade direta do dono (N1–N7); **P1** = sustenta as P0; **P2** = secundária.

| # | Spec | Prioridade | Plataformas |
|---|---|---|---|
| 00 | [Padrões transversais](00-padroes-transversais.md) — listas, estados, mutações, erros, i18n | P0 | todas |
| 01 | [Shell e navegação](01-shell-e-navegacao.md) | P0 | web |
| 02 | [Arquivos](02-arquivos.md) | P0 | web, Android |
| 03 | [Transferência entre dispositivos](03-transferencia-entre-dispositivos.md) — upload e download | P0 | web, Android |
| 04 | [Imagens](04-imagens.md) | P0 | web, Android |
| 05 | [Música](05-musica.md) | P0 | web, Android |
| 06 | [Vídeos](06-videos.md) | P0 | web, Android |
| 07 | [Documentos](07-documentos.md) | P0 | web |
| 08 | [Busca global](08-busca-global.md) | P0 | web, Android |
| 09 | [Favoritos](09-favoritos.md) | P1 | web, Android |
| 10 | [Lixeira](10-lixeira.md) | P1 | web, Android |
| 11 | [Download por URL](11-download-por-url.md) | P2 | web, plugin |
| 12 | [Capturas](12-capturas.md) | P2 | web, plugin |
| 13 | [Takeout](13-takeout.md) | P2 | web |
| 14 | [Home](14-home.md) | P1 | web, Android |
| 15 | [Notificações e jobs](15-notificacoes-e-jobs.md) | P1 | web, Android |
| 16 | [Analytics](16-analytics.md) | P2 | web |
| 17 | [Diário de atividades](17-diario-de-atividades.md) | P2 | web, Android |
| 18 | [Assistente](18-assistente.md) | P2 | web |
| 19 | [E-mail](19-email.md) | P2 | web, kiosk |
| 20 | [Armazenamento](20-armazenamento.md) — raízes, tiering, backup | P1 | web |
| 21 | [Indexação](21-indexacao.md) — scan, watcher, pastas monitoradas, bibliotecas | P1 | web |
| 22 | [Acesso pela rede](22-acesso-pela-rede.md) — whitelist de IPs, WebDAV | P1 | web |
| 23 | [Configurações e sistema](23-configuracoes-e-sistema.md) — aparência, idioma, IA, atualização, desligamento, assistente de configuração | P2 | web |
| 24 | [App Android](24-app-android.md) — conexão, navegação, paridade | P0 | Android |
| 25 | [App legado (kiosk)](25-app-legado-kiosk.md) | P2 | mobile |

## Como ler uma spec

Toda spec segue a mesma estrutura:

1. **Objetivo** — uma frase: para que a feature existe.
2. **Cenários de uso** — as situações reais do dono que ela atende.
3. **Escopo** e **Fora de escopo** — o que entra e o que deliberadamente não entra.
4. **Regras de negócio (`RN-nn`)** — o que vale sempre, independente de tela.
5. **Requisitos por tela (`RF-nn`)** — o que cada tela mostra e faz.
6. **Contrato HTTP** — endpoints que a feature usa.
7. **Estados e erros** — além do padrão de [00](00-padroes-transversais.md).
8. **Critérios de aceite (`CA-nn`)** — verificações objetivas, feitas no sistema rodando.
9. **Decisões em aberto** — o que só o dono decide; cada uma traz o comportamento provisório que vale até a decisão.
10. **Rastreabilidade** — itens dos boards que tratam da feature.

Referência cruzada: `02/RN-04` é a regra 04 da spec 02.

## Como usar na implementação

**Antes de começar (pronto para implementar):**

- A mudança cita os `RN`/`RF`/`CA` que ela satisfaz. Se não existe regra para o que vai ser feito, a spec é atualizada primeiro.
- Os consumidores do contrato afetado estão identificados (web, Android, mobile, plugin).

**Para dar por concluído:**

- Cada `CA` tocado foi verificado **no sistema rodando**, não só em teste unitário. Teste verde com mock não é evidência: a causa das regressões que motivaram esta especificação foi exatamente telas validadas apenas contra mocks.
- Os `CA` de [00](00-padroes-transversais.md) valem para toda tela com lista, mesmo que a mudança seja em outra parte da tela.
- Nenhum `CA` já marcado em outra spec deixou de valer (regressão). Mudança que toca componente compartilhado reverifica as telas que o usam.
- `make ci` verde.

**Checkboxes de aceite:** todos começam desmarcados, inclusive os de features já implementadas — o sistema hoje tem regressões e nada foi verificado contra esta spec. Um `CA` só é marcado depois de verificado no sistema rodando, no commit que registra a verificação.

**Contrato HTTP:** só cresce. Campo, parâmetro e rota não são removidos nem renomeados, porque web, dois apps Android e o plugin consomem a mesma API.

## Por onde começar

O levantamento feito para escrever estas specs (leitura de código, 2026-11-10) encontrou os defeitos abaixo. Eles explicam "telas com scroll que ficam sem itens" e "coisas que funcionavam e não funcionam mais", e bloqueiam as necessidades N1 e N2. Ordem sugerida de correção:

| # | Defeito | Spec |
|---|---|---|
| 1 | Arquivos, Favoritos, Recentes, busca na pasta e seletor de pasta nunca carregam a segunda página (o cliente lê `hasNext`, o servidor envia `has_next`) | [00 anexo A1–A3](00-padroes-transversais.md) |
| 2 | Em oito telas, a falha de uma página seguinte ou da revalidação automática troca a lista inteira pela tela de erro | [00 anexo A4](00-padroes-transversais.md) |
| 3 | O app Android não encontra o servidor (procura um nome de serviço diferente do anunciado) e a busca nunca termina | [24 §8](24-app-android.md) |
| 4 | O app Android mostra só 15 itens por pasta | [24 §8](24-app-android.md) |
| 5 | O app Android não envia arquivos para o servidor | [03 §9](03-transferencia-entre-dispositivos.md) |
| 6 | Excluir uma captura promovida apaga o vídeo da biblioteca | [12 §10](12-capturas.md) |
| 7 | Takeout sobrescreve arquivos de mesmo nome e descarta as datas das fotos | [13 §9](13-takeout.md) |
| 8 | Opções de configuração sem efeito; "Reindexar" e "Ver duplicados" que não fazem o que dizem | [21 §9](21-indexacao.md), [16 §8](16-analytics.md), [23 §8](23-configuracoes-e-sistema.md) |

Depois disso, o caminho é verificar cada spec P0 contra o sistema rodando, marcando os `CA` que passam e abrindo lacuna para os que falham.

## Backlog → spec

Cada item dos dois boards e a spec que o governa.

**`docs/revisao-produto/`**

| Item | Spec |
|---|---|
| 01 Arquivos | [02](02-arquivos.md), [03](03-transferencia-entre-dispositivos.md) |
| 02 Shell e layout | [01](01-shell-e-navegacao.md) |
| 03 Imagens | [04](04-imagens.md) |
| 04 Busca global | [08](08-busca-global.md), [07](07-documentos.md) |
| 05 Música | [05](05-musica.md) |
| 06 Vídeos | [06](06-videos.md) |
| 07 Favoritos | [09](09-favoritos.md) |
| 08 Lixeira | [10](10-lixeira.md) |
| 09 Downloads / ingestão por URL | [11](11-download-por-url.md) |
| 10 Capturas (plugin) | [12](12-capturas.md) |
| 11 Takeout | [13](13-takeout.md) |
| 12 Analytics | [16](16-analytics.md) |
| 13 Notificações e jobs | [15](15-notificacoes-e-jobs.md) |
| 14 Diário de atividades | [17](17-diario-de-atividades.md) |
| 15 Assistente | [18](18-assistente.md) |
| 16 Configurações | [20](20-armazenamento.md), [21](21-indexacao.md), [22](22-acesso-pela-rede.md), [23](23-configuracoes-e-sistema.md) |
| 17 E-mail | [19](19-email.md) |
| 18 Home | [14](14-home.md) |

**`docs/melhorias/`**

| Task | Spec |
|---|---|
| 01 Indexação de diretórios · 02 Filtro `deleted_at` · 03 Debounce do watcher · 06 Watcher por eventos · 07 Pipeline legado · 08 Decompor query de arquivos | [21](21-indexacao.md) |
| 04 Whitelist de IPs · 11 WebDAV | [22](22-acesso-pela-rede.md) |
| 05 Operações atualizam o banco na hora | [02/RN-04](02-arquivos.md) |
| 09 Lixeira | [10](10-lixeira.md) |
| 10 Múltiplas raízes · 12 Backup · 13 Tiering | [20](20-armazenamento.md) |
| 14 Contas de e-mail · 15 Sincronização · 16 Análise por IA | [19](19-email.md) |
| 17 Enxugar app legado · 18 Kiosk | [25](25-app-legado-kiosk.md) |
| 19 Captura inteligente de streaming | [12](12-capturas.md) |

## Decisões já tomadas (não reabrir)

- Sem autenticação; acesso por whitelist de IPs. TLS fora de escopo enquanto o uso for rede interna.
- Sistema de usuário único: favoritos, playlists e histórico são globais; apenas o estado dos players é por dispositivo.
- Redundância de disco é do sistema operacional; o KuraNAS enxerga volumes comuns.
- Backup tem retenção de versões, não é espelho. Tiering é transparente na árvore lógica.
- Ingestão de mídia: apenas fontes sem DRM; conteúdo com DRM só por captura da saída renderizada. Nenhum código de decriptação entra no repositório.
- E-mail: somente leitura, sem anexos, sem seguir URLs; o modelo de IA não recebe ferramentas.
- Mensagens vindas do backend são traduzidas no servidor e exibidas como chegam.

O texto completo dessas decisões está em `docs/melhorias/README.md` → "Decisões registradas".
