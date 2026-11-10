# 18 — Assistente

**Prioridade:** P2 · **Plataformas:** web · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Deixar o dono perguntar em linguagem natural sobre o que está guardado no servidor e receber a resposta com os arquivos encontrados, usando um modelo de IA que por padrão roda dentro de casa.

## 2. Cenários de uso

- "Onde está o contrato de aluguel?" — o assistente busca e responde com o arquivo, clicável.
- "Que músicas do Pink Floyd eu tenho?"
- Uma pergunta geral, sem relação com arquivos, respondida pelo modelo.

## 3. Escopo

Conversa com resposta em fluxo, histórico de conversas, consulta à biblioteca (somente leitura), escolha do provedor e do modelo de IA, gestão dos modelos locais.

## 4. Fora de escopo

- **Ações que alteram o servidor** (mover, excluir, renomear, enviar, baixar): o assistente só lê.
- Acesso do assistente à internet, a e-mails ou a qualquer coisa fora da biblioteca.
- Anexar arquivos à conversa; análise de imagem na conversa; voz.
- Memória entre conversas.
- Assistente no Android.

## 5. Regras de negócio

- **RN-01** — Com nenhum provedor de IA disponível, a tela explica isso e aponta para Configurações; não mostra erro genérico.
- **RN-02** — O provedor padrão é o local. Provedores de nuvem só são usados por escolha explícita do dono, com aviso de que as mensagens — e os nomes de arquivos encontrados — saem da rede.
- **RN-03** — O assistente tem uma única capacidade além de conversar: buscar na biblioteca (a mesma busca de [08](08-busca-global.md)), somente leitura. Itens da lixeira nunca aparecem.
- **RN-04** — Arquivos citados na resposta são links para o item no sistema.
- **RN-05** — A resposta aparece conforme é gerada e pode ser interrompida pelo dono; o que foi gerado até ali fica na conversa.
- **RN-06** — Mensagem cujo envio falha não fica órfã: a pergunta permanece visível, marcada como falha, com "Tentar de novo".
- **RN-07** — Erros dizem o que houve: IA indisponível, tempo esgotado, modelo não instalado, conversa inexistente. Nunca uma mensagem de outra feature.
- **RN-08** — O assistente responde no idioma do aplicativo.
- **RN-09** — Conversas são globais, ordenadas da mais recente para a mais antiga, com título tirado da primeira pergunta (editável). Excluir uma conversa pede confirmação.
- **RN-10** — Cada resposta registra qual provedor e modelo a gerou.

### Provedores e modelos

- **RN-11** — Os provedores têm prioridade; o primeiro habilitado e disponível é usado. Alterações valem na hora, sem reiniciar.
- **RN-12** — Valores de configuração são validados: modelo obrigatório para provedor habilitado, endereço válido, prioridade e limites numéricos dentro de faixa.
- **RN-13** — Chaves de API de provedores de nuvem não são digitadas nem exibidas na interface; a tela mostra apenas se a chave está configurada no servidor.
- **RN-14** — Baixar um modelo local é um job com progresso ([15](15-notificacoes-e-jobs.md)); remover um modelo pede confirmação.

## 6. Requisitos por tela

### Assistente (`/assistant`)

- **RF-01** — Coluna de conversas (paginada, com "Nova conversa", renomear, excluir) e área da conversa.
- **RF-02** — Área da conversa: mensagens do dono e do assistente, resposta em fluxo com indicador, "Parar" durante a geração, campo de texto (Enter envia, Shift+Enter quebra linha).
- **RF-03** — Resposta com formatação (listas, negrito, código) e arquivos como cartões clicáveis (ícone, nome, pasta).
- **RF-04** — Ações na resposta: copiar, gerar de novo.
- **RF-05** — Conversa vazia mostra exemplos de perguntas.
- **RF-06** — Indicação do provedor e modelo em uso, com atalho para Configurações.
- **RF-07** — Em tela estreita, a lista de conversas vira painel recolhível.

### Configurações → IA

- **RF-08** — Lista dos provedores: habilitado, modelo, endereço, prioridade, situação (disponível / indisponível e por quê / chave não configurada). Editar com validação e erro do servidor no campo.
- **RF-09** — Modelos locais: situação do serviço local, modelos instalados (nome, tamanho), instalar (com progresso), remover.
- **RF-10** — "Testar" um provedor: envia uma pergunta curta e mostra se respondeu.
- **RF-11** — Quais funções do sistema usam IA e se estão ligadas: marcação de imagens, análise de e-mail, busca com IA, assistente.

## 7. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| POST | `/assistant/chat/stream` | pergunta com resposta em fluxo |
| POST | `/assistant/chat` | pergunta com resposta inteira |
| GET | `/assistant/conversations` | conversas |
| GET | `/assistant/conversations/:id/messages` | mensagens |
| DELETE | `/assistant/conversations/:id` | excluir conversa |
| GET | `/ai/providers` · PUT `/ai/providers/:name` | provedores |
| GET | `/ai/ollama/status` · `/models` | serviço e modelos locais |
| POST | `/ai/ollama/models/pull` · DELETE `/ai/ollama/models/:name` | instalar, remover modelo |

**A criar** (não existe hoje): paginação de conversas e de mensagens; renomear conversa; testar provedor; validação dos campos do provedor.

## 8. Situação em 2026-11-10

- Todo erro do assistente responde com a mensagem "falha ao carregar configuração" e a tela mostra um erro genérico (contraria RN-07).
- A pergunta é gravada antes da geração; se a IA falha, fica sem resposta e sem como repetir (contraria RN-06).
- Lista de conversas sem paginação; excluir sem confirmação; sem "Parar".
- A decisão de buscar na biblioteca é por palavras-chave em português e o texto de instrução do modelo é fixo em português (contraria RN-08).
- Provedores aceitam qualquer valor (modelo vazio, prioridade negativa).

## 9. Critérios de aceite

- [ ] **CA-01** — Com o modelo local disponível, uma pergunta recebe resposta que aparece aos poucos.
- [ ] **CA-02** — Perguntar por um arquivo que existe devolve o arquivo como link que abre o item.
- [ ] **CA-03** — Com a IA indisponível, a tela diz que a IA está indisponível e aponta para Configurações.
- [ ] **CA-04** — Uma pergunta cuja resposta falha fica marcada e "Tentar de novo" a reenvia.
- [ ] **CA-05** — "Parar" interrompe a geração e mantém o texto parcial.
- [ ] **CA-06** — Reabrir uma conversa antiga mostra todas as mensagens; a lista de conversas com mais de 50 itens rola até o fim.
- [ ] **CA-07** — Excluir uma conversa pede confirmação e a remove.
- [ ] **CA-08** — Com o aplicativo em inglês, o assistente responde em inglês e encontra arquivos por pergunta em inglês.
- [ ] **CA-09** — Um arquivo na lixeira não é citado pelo assistente.
- [ ] **CA-10** — Pedir ao assistente para excluir ou mover um arquivo não altera nada no servidor.
- [ ] **CA-11** — Desabilitar o provedor local e habilitar outro vale na pergunta seguinte, sem reiniciar.
- [ ] **CA-12** — Salvar um provedor habilitado sem modelo é recusado com mensagem no campo.
- [ ] **CA-13** — Instalar um modelo local mostra o progresso até concluir; o modelo aparece na lista.

## 10. Decisões em aberto

- **D-01** — Esta feature não atende diretamente a nenhuma das necessidades N1–N7 (a busca global já resolve "achar"). Manter, congelar ou remover. *Provisório: manter como está, corrigindo apenas RN-06 e RN-07; nenhuma evolução antes das features P0.*

## 11. Rastreabilidade

`docs/revisao-produto` item 15 (pendente).
