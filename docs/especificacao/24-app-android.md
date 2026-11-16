# 24 — App Android

**Prioridade:** P0 · **Plataformas:** Android (`android/`) · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Fazer do celular e do tablet clientes completos do servidor: achar o servidor sozinho na rede de casa, enviar e baixar arquivos, ver fotos, ouvir música e assistir vídeos.

## 2. Cenários de uso

- Instala o app, abre, e ele encontra o servidor sem digitar endereço.
- Tira fotos e as manda para o servidor ([03](03-transferencia-entre-dispositivos.md)).
- Baixa no celular um arquivo que subiu pelo PC.
- Ouve música com a tela bloqueada; assiste a um episódio no tablet continuando de onde parou no PC.
- Sai de casa: o app diz que o servidor não está acessível, sem travar; ao voltar, reconecta.

## 3. Escopo

Conexão com o servidor (descoberta, endereço manual, troca, reconexão), navegação principal, comportamento comum das telas, e a paridade de cada feature com a web. O comportamento de cada feature está na spec dela, na parte "Android".

## 4. Fora de escopo

- Uso fora da rede de casa (acesso remoto).
- Conteúdo offline (baixar músicas e vídeos para ouvir/assistir sem rede). *(ver [05/D-03](05-musica.md), [06/D-04](06-videos.md))*
- iOS.
- Publicação em loja; a instalação é por APK baixado do servidor ([23/RF-12](23-configuracoes-e-sistema.md)).
- Configuração do servidor pelo app (raízes, backup, acesso, IA).
- Assistente, Analytics, Takeout, Capturas e E-mail no app.
- Android TV / interface para controle remoto. *(ver D-02)*
- Vários servidores ao mesmo tempo.

## 5. Regras de negócio

### Conexão

- **RN-01** — Ao abrir sem servidor salvo, o app procura o servidor na rede local automaticamente, por todos os meios que o servidor anuncia, e lista os encontrados com nome e endereço.
- **RN-02** — A busca tem fim: em até 10 s, ou mostra os servidores encontrados, ou diz que nenhum foi encontrado e oferece "Procurar de novo" e o endereço manual.
- **RN-03** — Um endereço — descoberto ou digitado — só é aceito depois de confirmado como servidor KuraNAS acessível. Endereço errado mostra o motivo (sem resposta; não é KuraNAS; acesso não liberado para este dispositivo).
- **RN-04** — O servidor escolhido é lembrado. Nas aberturas seguintes o app conecta direto.
- **RN-05** — Se o servidor lembrado não responde, o app tenta de novo a descoberta (o IP pode ter mudado) antes de pedir ação do dono.
- **RN-06** — Dispositivo fora da whitelist vê a explicação e o próprio IP, para o dono cadastrar ([22/RF-06](22-acesso-pela-rede.md)).
- **RN-07** — "Trocar servidor" esquece o atual e volta à tela de conexão; funciona de qualquer ponto em que é oferecido.

### Comportamento comum

- **RN-08** — Toda lista segue [00](00-padroes-transversais.md): paginada com `page`/`page_size` explícitos, scroll infinito que anexa, puxar para recarregar, erro que não esvazia a lista, estados distintos de carregando/vazio/erro.
- **RN-09** — Toda falha de operação é mostrada com a mensagem do servidor; nenhuma falha é ignorada em silêncio. Atualização otimista volta atrás quando o servidor recusa.
- **RN-10** — Sem conexão com o servidor, o app mostra um aviso persistente e continua navegável; ao reconectar, as telas recarregam.
- **RN-11** — O app identifica o dispositivo para o servidor, para que o estado dos players seja por dispositivo.
- **RN-12** — Todo texto vem dos recursos de tradução (pt-BR e en-US). Mensagens do servidor são exibidas como chegam.
- **RN-13** — O app segue o tema escolhido no servidor ([23/RN-07](23-configuracoes-e-sistema.md)) e funciona em retrato e paisagem, em celular e tablet (em tablet, usa a largura: grades com mais colunas, painel lateral).
- **RN-14** — O botão "voltar" do sistema volta à tela anterior; na tela inicial, sai do app. Rotação e troca de app não perdem o que estava na tela (lista, rolagem, formulário).
- **RN-15** — Toda tela existente é alcançável pela navegação; não há tela sem entrada nem item de menu sem efeito.
- **RN-16** — O app funciona em Android 8 ou superior.

### Mídia

- **RN-17** — Música toca em segundo plano com notificação e controles de tela de bloqueio, fone e Bluetooth ([05/RF-24](05-musica.md)).
- **RN-18** — Vídeo toca em tela cheia com retomada, legendas e próximo episódio ([06/RF-17](06-videos.md)).
- **RN-19** — Formato que o aparelho não reproduz é pedido ao servidor já convertido.

## 6. Requisitos por tela

### Conexão

- **RF-01** — Tela inicial sem servidor: indicador de busca, lista dos servidores encontrados (toque conecta), "Procurar de novo", campo de endereço manual com "Conectar" e a explicação do que fazer se nada aparece (estar no mesmo Wi-Fi, servidor ligado).
- **RF-02** — Erro de conexão com motivo (RN-03).

### Navegação principal

- **RF-03** — Barra inferior com cinco destinos: Início, Arquivos, Fotos, Música, Vídeos. Menu "Mais" (ou gaveta lateral) com: Documentos, Favoritos, Busca, Transferências, Lixeira, Notificações, Tarefas, Baixar de URL, Diário, Configurações.
- **RF-04** — Busca acessível do topo de qualquer tela principal.
- **RF-05** — Player de música reduzido, persistente acima da barra inferior, quando há fila.
- **RF-06** — Indicador de transferências em andamento ([03/RF-14](03-transferencia-entre-dispositivos.md)) e contador de notificações não lidas.

### Telas de feature

| Tela | Spec |
|---|---|
| Início | [14/RF-09](14-home.md) |
| Arquivos | [02/RF-23](02-arquivos.md) |
| Enviar, baixar, transferências, envio automático | [03/RF-11 a RF-18](03-transferencia-entre-dispositivos.md) |
| Fotos | [04/RF-17](04-imagens.md) |
| Música | [05/RF-24](05-musica.md) |
| Vídeos | [06/RF-17](06-videos.md) |
| Documentos | [07/RF-12](07-documentos.md) |
| Busca | [08/RF-12](08-busca-global.md) |
| Favoritos | [09/RF-09](09-favoritos.md) |
| Lixeira | [10/RF-12](10-lixeira.md) |
| Baixar de URL | [11/RF-08](11-download-por-url.md) |
| Notificações e tarefas | [15/RF-11](15-notificacoes-e-jobs.md) |
| Diário | [17/RF-06](17-diario-de-atividades.md) |
| Configurações | [23/RF-17](23-configuracoes-e-sistema.md) |

## 7. Contrato

O app usa a mesma API da web (`/api/v1`). O servidor se anuncia de duas formas, e o app usa as duas ao mesmo tempo (RN-01):

- **mDNS**: serviço `_kuranas._tcp`, na porta HTTP do servidor (8000).
- **UDP**: o app envia o texto `KURANAS_DISCOVER` por broadcast (`255.255.255.255`) para a porta 19520; o servidor responde ao remetente com `{"service": "kuranas", "port": 8000, "api": "/api/v1"}`. O endereço do servidor é o IP de origem da resposta mais `port`. Respostas com `service` diferente de `kuranas` são ignoradas.

**Confirmação do servidor (RN-03)**: `GET /api/v1/health` deve responder `200` com `{"status": "ok", "service": "kuranas"}` em até 3 s. Sem resposta → "sem resposta"; resposta que não é esse JSON → "não é KuraNAS"; `403` → "acesso não liberado para este dispositivo". Endereço digitado sem esquema ganha `http://`; sem porta, ganha `:8000`.

Toda mudança de rota ou de formato no backend confere este consumidor (`android/app/src/main/java/com/kuranas/android/feature/*/data`).

## 8. Situação em 2026-11-10

Paridade com a web, por feature:

| Feature | Situação no app |
|---|---|
| Descoberta do servidor | **Não funciona.** O app procura o serviço `_nas._tcp`; o servidor anuncia `_kuranas._tcp`. A busca nunca termina (indicador para sempre). Não usa a descoberta por UDP. Só o endereço manual funciona, e ele não é validado. |
| Arquivos | **Cada pasta mostra só 15 itens**: não envia paginação e não carrega mais. Operações de renomear/mover/nova pasta não verificadas. |
| **Enviar para o servidor** | **Não existe.** A chamada de envio está declarada e nunca é usada; não há seletor de arquivos, "Compartilhar para o KuraNAS" nem envio automático da câmera. |
| Baixar para o aparelho | Existe, básico: carrega o arquivo inteiro na memória, sem progresso nem retomada — inviável para arquivo grande. |
| Fotos, Música, Vídeos, Favoritos, Lixeira | Paginados (carregar mais funciona). |
| Notificações, Diário | Só a primeira página; falhas de escrita ignoradas. |
| Tarefas | Tela existe e não é alcançável. |
| Início | Uma falha derruba a tela inteira; sem retomada de vídeo/música nem favoritos. |
| Navegação | Onze páginas deslizáveis; "Esquecer servidor" na gaveta não faz nada (o da tela de Configurações funciona). |
| Preferências do servidor (tema, players) | Lidas e não aplicadas. |
| Documentos, Baixar de URL, Transferências | Não existem. |

Duas das necessidades centrais do dono — enviar foto do celular (N1) e baixar no celular o que subiu pelo PC (N2) — dependem diretamente das linhas em negrito.

## 9. Critérios de aceite

- [ ] **CA-01** — Com o servidor ligado e o aparelho no mesmo Wi-Fi, o app recém-instalado lista o servidor em até 10 s, sem digitar endereço.
- [ ] **CA-02** — Com o servidor desligado, a busca termina em até 10 s dizendo que nenhum servidor foi encontrado.
- [ ] **CA-03** — Digitar um endereço que não é um servidor KuraNAS é recusado com o motivo.
- [ ] **CA-04** — Fechar e reabrir o app conecta direto ao servidor lembrado.
- [ ] **CA-05** — Trocar o IP do servidor: o app o reencontra sem o dono digitar nada.
- [ ] **CA-06** — Em aparelho fora da whitelist, o app mostra que o acesso não está liberado e o IP do aparelho.
- [ ] **CA-07** — "Trocar servidor" volta à tela de conexão a partir de todos os lugares em que aparece.
- [ ] **CA-08** — Uma pasta com 300 itens rola até o fim em Arquivos.
- [ ] **CA-09** — Os critérios de [03](03-transferencia-entre-dispositivos.md) CA-13 a CA-20 (enviar e baixar pelo Android) passam.
- [ ] **CA-10** — Desligar o Wi-Fi com o app aberto mostra o aviso de sem conexão sem fechar o app; religar recarrega a tela.
- [ ] **CA-11** — Girar o aparelho em qualquer lista mantém os itens carregados e a posição.
- [ ] **CA-12** — Todas as telas do menu abrem; nenhum item do menu fica sem efeito.
- [ ] **CA-13** — Em tablet de 10", as grades usam a largura e não há texto ou controle cortado; em celular de 5", idem.
- [ ] **CA-14** — Com o servidor recusando uma operação, o app mostra a mensagem do servidor.
- [ ] **CA-15** — Os critérios "No Android" de cada spec de feature passam.

## 10. Decisões em aberto

- **D-01** — Navegação: barra inferior com cinco destinos mais "Mais" (RF-03) no lugar das onze páginas deslizáveis de hoje. *Provisório: RF-03.*
- **D-02** — Interface para TV (assistir os vídeos na sala pelo controle remoto). Atende diretamente N6. *Provisório: fora de escopo; na TV, usa-se o navegador ou um aparelho Android com toque.*

## 11. Rastreabilidade

Lacunas de Android em `docs/revisao-produto/` 03 (05), 05 (21.x), 06 (64–68.x), 07 (21–23.x), 08 (29–30), 09 (23). Sem item no backlog: descoberta do servidor (RN-01 a RN-05), envio pelo Android, paginação do explorer, navegação.
