# 23 — Configurações e sistema

**Prioridade:** P2 · **Plataformas:** web · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Reunir em um lugar só tudo o que o dono ajusta no sistema — aparência, idioma, players, IA, atualização, desligamento — e as informações sobre o servidor, de modo que cada ajuste tenha efeito real e imediato.

## 2. Cenários de uso

- Troca o tema e o idioma.
- Programa o servidor para desligar de madrugada.
- Vê que há versão nova e atualiza.
- Instala o app Android no tablet e a extensão no navegador a partir do próprio servidor.
- Na primeira instalação, configura banco, pasta e portas pelo assistente de configuração.

## 3. Escopo

Tela de Configurações (estrutura e comportamento comum), aparência, idioma, players, desligamento automático, atualização do sistema, Sobre, distribuição dos aplicativos cliente, assistente de configuração do servidor.

Seções com spec própria — a tela de Configurações só as hospeda: Armazenamento ([20](20-armazenamento.md)), Indexação, Bibliotecas e Pastas monitoradas ([21](21-indexacao.md)), Acesso ([22](22-acesso-pela-rede.md)), IA ([18](18-assistente.md)), E-mail ([19](19-email.md)), Lixeira ([10](10-lixeira.md)), ferramenta de download ([11](11-download-por-url.md)), Capturas ([12](12-capturas.md)).

## 4. Fora de escopo

- Usuários, perfis e permissões.
- Configuração por dispositivo guardada no servidor (o que é por dispositivo fica no dispositivo).
- Editar configuração do servidor pelo Android.
- Reiniciar o servidor pela interface. *(ver D-02)*
- Agendamento de ligar o servidor (depende de BIOS/rede).
- Exportar e importar configurações.

## 5. Regras de negócio

### Comportamento comum

- **RN-01** — Toda configuração exibida tem efeito real. Opção que nenhum código lê não aparece na tela.
- **RN-02** — **Um só modelo de salvar em toda a tela**: cada seção salva ao alterar, com confirmação visível ("salvo") ou erro no campo. Não coexistem seções que salvam sozinhas e seções que dependem de um "Salvar" geral.
- **RN-03** — Valor inválido é apontado no campo, com a mensagem do servidor; o valor anterior é mantido.
- **RN-04** — Configuração que só vale após reiniciar diz isso ao lado do campo.
- **RN-05** — Falha ao carregar uma seção mostra erro com "Tentar de novo" naquela seção; as demais funcionam.
- **RN-06** — Cada seção tem endereço próprio (`/settings#secao`), usado por atalhos de outras telas e pela busca global.

### Aparência e idioma

- **RN-07** — Tema (claro, escuro, do sistema), cor de destaque e redução de movimento valem em todos os dispositivos e são aplicados na hora, na web e no Android.
- **RN-08** — O idioma do aplicativo (pt-BR, en-US) vale para os textos da interface. As mensagens vindas do servidor seguem o idioma do servidor; a tela informa quando os dois diferem.

### Players

- **RN-09** — Lembrar a fila de música; lembrar o progresso de vídeo; reproduzir o próximo vídeo automaticamente; tempo da apresentação de fotos. Valem como padrão para todos os dispositivos, na web e no Android.

### Desligamento automático

- **RN-10** — Desligado por padrão. Ligado, desliga **o computador servidor** no horário configurado.
- **RN-11** — Antes de desligar há um período de aviso (padrão 60 s), com notificação visível em todos os clientes abertos e a ação "Cancelar desta vez".
- **RN-12** — Não desliga enquanto há atividade que não pode ser interrompida: envio em andamento, backup, migração de tier, reprodução ativa em algum dispositivo. Nesse caso adia e avisa.
- **RN-13** — "Sugerir horário" propõe um horário com base nos desligamentos anteriores, quando há histórico suficiente.

### Atualização

- **RN-14** — O sistema verifica se há versão nova e avisa por notificação. Atualizar é sempre ação explícita do dono.
- **RN-15** — O pacote baixado é conferido por integridade (checksum) antes de ser aplicado; pacote que não confere é descartado e a versão atual continua.
- **RN-16** — A tela avisa, antes de confirmar, que o servidor ficará indisponível por instantes e que as reproduções serão interrompidas. Depois de reiniciar, os clientes reconectam sozinhos ([01/RF-12](01-shell-e-navegacao.md)) e mostram a nova versão.
- **RN-17** — Atualização que falha deixa a versão anterior funcionando e informa o motivo.

### Assistente de configuração do servidor

- **RN-18** — Só pode ser usado a partir do próprio servidor (localhost). De outro dispositivo, a tela explica isso.
- **RN-19** — Segredos (senhas, chaves) nunca são exibidos; só é possível substituí-los.
- **RN-20** — Antes de gravar, a configuração anterior é guardada em cópia de segurança. A tela lista o que mudou e avisa que é preciso reiniciar o servidor.
- **RN-21** — Banco e caminhos podem ser testados antes de salvar.

### Distribuição dos aplicativos

- **RN-22** — O servidor oferece para download os aplicativos cliente (app Android, app legado, extensão do navegador) na versão compatível com ele, com tamanho e checksum.

## 6. Requisitos por tela

### Configurações (`/settings`)

- **RF-01** — Índice lateral das seções (em tela estreita, lista que leva a cada seção), agrupadas: **Geral** (Aparência, Idioma, Players) · **Arquivos** (Armazenamento, Bibliotecas, Indexação, Pastas monitoradas, Lixeira, Backup, Tiering) · **Rede** (Acesso, WebDAV) · **Mídia e IA** (IA, Ferramenta de download, Capturas, E-mail) · **Sistema** (Desligamento automático, Atualização, Sobre).
- **RF-02** — Busca dentro das configurações.

### Aparência, Idioma, Players

- **RF-03** — Tema com pré-visualização imediata; cor de destaque; reduzir movimento.
- **RF-04** — Idioma do aplicativo; aviso quando difere do idioma do servidor.
- **RF-05** — As quatro preferências de RN-09.

### Desligamento automático

- **RF-06** — Ligar/desligar; horário; período de aviso; "Sugerir horário"; próxima execução; texto dizendo claramente que o computador será desligado.
- **RF-07** — Durante o período de aviso, faixa em todas as telas com contagem regressiva e "Cancelar desta vez".

### Atualização e Sobre (`/about`)

- **RF-08** — Versão atual, data da versão, identificador da compilação (copiável), plataforma, tempo ligado, idioma do servidor, situação dos processos de fundo.
- **RF-09** — Atualização: "em dia" ou a versão disponível com link para as novidades; "Atualizar" com confirmação (RN-16), progresso e resultado.
- **RF-10** — Situação das dependências externas — conversão de mídia, ferramenta de download, IA local — com o que cada uma habilita e o que fica indisponível sem ela.
- **RF-11** — Atalhos: assistente de configuração, tarefas, logs (onde ficam no servidor).

### Aplicativos (`/downloads`)

- **RF-12** — Lista dos aplicativos disponíveis: nome, plataforma, versão, requisito mínimo, tamanho, "Baixar". Instruções de instalação de cada um (APK no Android; extensão não publicada no navegador).
- **RF-13** — Em um celular, o item do app Android vem em destaque.
- **RF-14** — Sem aplicativos publicados no servidor, estado vazio explicando.

### Assistente de configuração (`/internal/config-wizard`)

- **RF-15** — Passos: Geral (pasta principal, porta, idioma), Banco, Acesso, E-mail, IA e processos de fundo, Revisão. Cada passo valida antes de avançar.
- **RF-16** — Revisão: o que será alterado, o que exige atenção, confirmação; ao salvar, instrução de reinício.

### Android → Configurações

- **RF-17** — Servidor conectado e situação; trocar/esquecer servidor; versão do app e do servidor; aviso de atualização disponível; envio automático da câmera ([03/RF-15](03-transferencia-entre-dispositivos.md)); tema seguindo a preferência do servidor.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET, PUT | `/configuration/settings` | aparência, idioma, players, indexação |
| GET | `/configuration/about` | informações do servidor |
| GET | `/configuration/translation` | textos |
| GET, PUT | `/auto-shutdown/settings` · GET `/auto-shutdown/suggested-time` | desligamento |
| GET | `/update/status` · POST `/update/apply` | atualização |
| GET | `/downloads` · `/downloads/:id` | aplicativos |
| GET, PUT | `/configuration/env` · POST `/env/test-db` · `/env/test-path` | assistente de configuração |

**A criar** (não existe hoje): cancelar o desligamento em curso; situação das dependências externas.

## 8. Situação em 2026-11-10

- Três opções de indexação não têm efeito ([21 §9](21-indexacao.md)) (contraria RN-01).
- Nove seções salvam ao alterar e seis dependem de um "Salvar" geral no rodapé (contraria RN-02).
- A atualização confere só o **tamanho** do pacote, não a integridade (contraria RN-15).
- O desligamento não pode ser cancelado pela interface e não considera atividade em andamento (contraria RN-11 e RN-12).
- O Android lê as preferências e não aplica nenhuma (contraria RN-07 e RN-09).
- O campo de data de início do servidor é serializado como `statup_time` (erro de grafia que o contrato obriga a manter; um campo correto pode ser acrescentado).

## 9. Critérios de aceite

- [ ] **CA-01** — Cada opção da tela, ao ser alterada, produz o efeito que descreve.
- [ ] **CA-02** — Alterar qualquer opção salva na hora, com confirmação; recarregar a página mantém o valor.
- [ ] **CA-03** — Um valor inválido é recusado com mensagem no campo e o valor anterior permanece.
- [ ] **CA-04** — Com o endpoint de uma seção falhando, as outras seções carregam e funcionam.
- [ ] **CA-05** — Trocar o tema aplica na hora e aparece no outro dispositivo após recarregar; no Android também.
- [ ] **CA-06** — Trocar o idioma muda todos os textos da interface; nenhuma chave crua aparece.
- [ ] **CA-07** — Desligar "reproduzir o próximo automaticamente" faz o vídeo seguinte não começar sozinho em um dispositivo que nunca escolheu essa preferência.
- [ ] **CA-08** — Com o desligamento programado, no horário aparece o aviso com contagem; "Cancelar desta vez" impede o desligamento; sem cancelar, o computador desliga.
- [ ] **CA-09** — Com um envio em andamento no horário do desligamento, ele é adiado e avisado.
- [ ] **CA-10** — Havendo versão nova, a tela a mostra; atualizar reinicia o servidor e os clientes reconectam mostrando a nova versão.
- [ ] **CA-11** — Um pacote de atualização adulterado é recusado e a versão atual continua funcionando.
- [ ] **CA-12** — O app Android baixado em "Aplicativos" instala e conecta ao servidor.
- [ ] **CA-13** — De outro dispositivo, o assistente de configuração explica que só funciona no servidor; no servidor, salvar cria a cópia de segurança e nunca exibe segredos.
- [ ] **CA-14** — "Sobre" mostra a situação real das dependências externas.

## 10. Decisões em aberto

- **D-01** — Idioma: hoje há o idioma do servidor (definido na instalação) e o da interface. Unificar em uma preferência só. *Provisório: RN-08 (dois, com aviso quando diferem), conforme a decisão registrada "mix simples".*
- **D-02** — Reiniciar o servidor pela interface (após o assistente de configuração). *Provisório: fora de escopo; reinício manual.*
- **D-03** — Verificação de integridade da atualização por checksum publicado junto da versão. *Provisório: em escopo (RN-15).*

## 11. Rastreabilidade

`docs/revisao-produto` item 16 (pendente); `docs/config-wizard.md`; `02-shell-layout.md` lacuna 08.
