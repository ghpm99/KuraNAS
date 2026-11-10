# 20 — Armazenamento

**Prioridade:** P1 · **Plataformas:** web · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Dizer ao sistema onde os arquivos ficam no servidor, aproveitar disco rápido e disco grande sem o dono ter de pensar nisso, e manter uma cópia de segurança recuperável.

## 2. Cenários de uso

- O dono compra um HD novo e o adiciona como mais uma área de armazenamento; em minutos o conteúdo dele aparece no sistema.
- O SSD está enchendo: arquivos que ele não toca há meses vão sozinhos para o HD grande, sem mudar de lugar nas pastas.
- Apagou e esvaziou a lixeira por engano, ou um arquivo corrompeu: recupera a versão anterior do backup.
- Quer saber, em uma olhada, se o backup está em dia.

## 3. Escopo

Raízes de armazenamento (cadastrar, editar, desabilitar, remover), tier quente/frio (regras, migração automática, uso por tier, promoção manual), backup com retenção de versões (configuração, execução, situação).

## 4. Fora de escopo

- Gestão de discos, volumes, RAID, SMART. Redundância é do sistema operacional.
- Quotas por raiz.
- Mais de dois tiers; cache de leitura.
- Backup para nuvem ou para fora de casa; deduplicação, compressão e criptografia do backup.
- Gerência do HD externo desconectável (segunda cópia fica com o sistema operacional).
- Restauração de backup pela interface. *(ver D-01)*
- Fixar arquivo ou pasta no tier quente. *(ver D-02)*

## 5. Conceitos

- **Raiz**: pasta do servidor que o KuraNAS gerencia. Tudo o que o sistema mostra está dentro de alguma raiz.
- **Tier quente / frio**: armazenamento rápido (SSD) e armazenamento grande (HDs). Um arquivo está em um dos dois; o caminho que o dono vê é o mesmo.
- **Área de backup**: pasta fora de todas as raízes que recebe as cópias.
- **Versão**: cópia anterior de um arquivo alterado ou excluído, guardada por um prazo.

## 6. Regras de negócio

### Raízes

- **RN-01** — O sistema tem pelo menos uma raiz. Cada raiz tem nome de exibição e caminho no servidor.
- **RN-02** — O caminho de uma raiz precisa existir, ser uma pasta e permitir leitura e escrita; é verificado ao cadastrar, com mensagem específica para cada falha.
- **RN-03** — Raízes não se sobrepõem: uma não pode ser ancestral nem descendente de outra, nem conter a área de backup ou o tier frio.
- **RN-04** — Cadastrar uma raiz inicia a indexação dela ([21](21-indexacao.md)) e ela aparece no nível mais alto de Arquivos.
- **RN-05** — Desabilitar uma raiz a tira da navegação, da busca e das bibliotecas, sem apagar nada do disco nem do índice; reabilitar traz tudo de volta.
- **RN-06** — Remover uma raiz remove seus itens do índice, **nunca do disco**. Pede confirmação dizendo exatamente isso. A última raiz não pode ser removida.
- **RN-07** — Raiz cujo volume está desconectado aparece como indisponível; seus itens não são marcados como excluídos por causa disso.
- **RN-08** — Nenhuma operação do sistema lê ou escreve fora das raízes, do tier frio e da área de backup configurados.

### Tiering

- **RN-09** — Desligado por padrão. Ligado, exige um caminho de tier frio válido, fora de todas as raízes.
- **RN-10** — Arquivo sem interação (abrir, tocar, baixar, modificar) há mais de N dias e acima de um tamanho mínimo vai para o frio na execução periódica. Arquivo frio que volta a ser usado retorna ao quente na execução seguinte.
- **RN-11** — A migração não muda o caminho lógico, a identidade, os metadados, a estrela nem a presença do arquivo em qualquer tela.
- **RN-12** — Toda operação funciona igual com arquivo frio; a única diferença visível é o indicador de tier (e possível lentidão).
- **RN-13** — A migração nunca perde arquivo: copia, confere, atualiza o registro e só então remove a origem. Interrompida no meio, o arquivo está íntegro em um dos dois lados.
- **RN-14** — O dono pode trazer um arquivo de volta para o quente na hora.

### Backup

- **RN-15** — Desligado por padrão. Ligado, exige destino válido fora de todas as raízes.
- **RN-16** — Roda no horário configurado e copia só o que é novo ou mudou desde a última execução. Cada cópia é conferida por checksum antes de ser considerada feita.
- **RN-17** — Quando um arquivo é alterado ou excluído na origem, a versão anterior fica guardada por N dias (retenção). Backup não é espelho: excluir na origem não exclui a cópia imediatamente.
- **RN-18** — A área de backup, o tier frio e a lixeira nunca aparecem em Arquivos, bibliotecas, busca ou analytics.
- **RN-19** — Conclusão com falhas e falha total geram notificação. Backup interrompido retoma na próxima execução sem corromper nada.
- **RN-20** — O backup cobre arquivos nos dois tiers.

## 7. Requisitos por tela

### Configurações → Armazenamento

- **RF-01** — Lista das raízes: nome, caminho, situação (ativa, desabilitada, indisponível), espaço usado e livre do volume, quantidade de itens indexados.
- **RF-02** — Adicionar raiz: nome e caminho, com "Testar caminho" antes de salvar. Editar nome; habilitar/desabilitar; remover (RN-06).
- **RF-03** — Depois de adicionar, a linha da raiz mostra o andamento da indexação.

### Configurações → Tiering

- **RF-04** — Ligar/desligar; caminho do tier frio (com teste); dias sem uso; tamanho mínimo; horário da execução.
- **RF-05** — Uso por tier (espaço e quantidade de arquivos) e situação da última execução (quando, quantos migrados para cada lado, falhas).
- **RF-06** — Nos detalhes de um arquivo frio: "Trazer para o quente agora" ([02/RF-13](02-arquivos.md)).

### Configurações → Backup

- **RF-07** — Ligar/desligar; destino (com teste); horário; dias de retenção de versões.
- **RF-08** — Situação: última execução (quando, duração, resultado), arquivos pendentes, falhas com motivo, próxima execução. "Executar agora".
- **RF-09** — Explicação de como recuperar um arquivo do backup.

Toda alteração segue [00 §6](00-padroes-transversais.md): confirmação de sucesso, erro do servidor no campo, efeito sem reiniciar.

## 8. Contrato HTTP

| Método | Rota | Uso |
|---|---|---|
| GET, POST | `/storage-roots` | listar, cadastrar |
| PUT, DELETE | `/storage-roots/:id` | editar, remover |
| POST | `/configuration/env/test-path` | testar caminho |
| GET, PUT | `/tiering/settings` | regras de tiering |
| GET | `/tiering/status` · `/tiering/usage` | última execução, uso por tier |
| POST | `/tiering/promote/:file_id` | trazer para o quente |
| GET, PUT | `/backup/settings` | configuração do backup |
| GET | `/backup/status` · `/backup/pending` | situação, pendentes |

**A criar** (não existe hoje): "Executar agora" do backup.

## 9. Estados e erros

| Situação | O que o usuário vê |
|---|---|
| Caminho inexistente / sem permissão / não é pasta | motivo específico ao testar e ao salvar |
| Raiz sobreposta a outra | recusa dizendo com qual conflita |
| Volume desconectado | raiz indisponível, sem perda de índice |
| Tier frio ou destino de backup indisponível na execução | execução falha, notificação, nada é removido da origem |
| Disco de destino cheio | falha com motivo; arquivos já copiados continuam válidos |

## 10. Critérios de aceite

- [ ] **CA-01** — Adicionar uma segunda raiz a faz aparecer em Arquivos e seu conteúdo é indexado sem reiniciar.
- [ ] **CA-02** — Cadastrar um caminho dentro de uma raiz existente é recusado com explicação.
- [ ] **CA-03** — Desabilitar uma raiz tira seus itens de Arquivos, da busca e das bibliotecas; reabilitar os devolve com estrelas e playlists intactas.
- [ ] **CA-04** — Remover uma raiz não apaga nenhum arquivo do disco.
- [ ] **CA-05** — Desconectar o volume de uma raiz e reconectar: nenhum item foi para a lixeira nem perdeu metadados.
- [ ] **CA-06** — Enviar, mover, renomear e excluir funcionam em qualquer raiz; mover entre raízes preserva o arquivo e sua estrela.
- [ ] **CA-07** — Com tiering ligado, um arquivo além do limite de dias vai para o frio na execução e continua no mesmo lugar em Arquivos, abrindo normalmente, com indicador de tier.
- [ ] **CA-08** — Usar um arquivo frio o traz de volta ao quente na execução seguinte; "Trazer para o quente agora" faz isso na hora.
- [ ] **CA-09** — Interromper o servidor durante a migração não perde nem duplica arquivos.
- [ ] **CA-10** — O uso por tier mostrado confere com o disco.
- [ ] **CA-11** — Com backup ligado, um arquivo novo é copiado para o destino na execução seguinte e a situação mostra "em dia".
- [ ] **CA-12** — Alterar um arquivo e rodar o backup mantém a versão anterior recuperável; passado o prazo de retenção, ela é removida.
- [ ] **CA-13** — Excluir um arquivo e esvaziar a lixeira: a cópia do backup continua recuperável dentro da retenção.
- [ ] **CA-14** — Backup com destino indisponível gera notificação de falha e não altera a origem.
- [ ] **CA-15** — A área de backup e o tier frio não aparecem em nenhuma tela de conteúdo.

## 11. Decisões em aberto

- **D-01** — Restaurar do backup pela interface (procurar arquivo, escolher versão, restaurar). Hoje é manual, pelo sistema de arquivos. *Provisório: fora de escopo; RF-09 explica o procedimento manual.*
- **D-02** — Fixar arquivo ou pasta no quente ("nunca migrar"). *Provisório: fora de escopo.*

## 12. Rastreabilidade

`docs/melhorias/` tasks 10 (raízes), 12 (backup), 13 (tiering); `docs/revisao-produto` item 16 (Configurações, pendente).
