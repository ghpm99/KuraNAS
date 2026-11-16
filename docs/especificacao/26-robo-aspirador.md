# 26 — Robô aspirador

**Prioridade:** P2 · **Plataformas:** web · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Comandar o robô aspirador da casa e ver o que ele está fazendo a partir do KuraNAS, pela rede local, sem depender do aplicativo do fabricante nem da nuvem dele.

## 2. Cenários de uso

- O dono está no PC e quer mandar o robô limpar a casa sem pegar o celular.
- O robô está limpando e alguém vai atender uma ligação: pausa, depois continua.
- Quer saber se o robô já terminou, quanto de bateria tem e se parou por alguma falha.
- O robô sumiu embaixo de um móvel: manda ele apitar.
- Trocou o filtro ou a escova e zera o contador de desgaste.

## 3. Escopo

Um robô, modelo KaBuM! Smart 900 (plataforma Tuya, protocolo local 3.3): cadastro da conexão, busca do robô na rede, estado atual, comandos de limpeza, controle manual, ajustes de limpeza, base de esvaziamento, desgaste de peças e totais acumulados.

## 4. Fora de escopo

- Mapa da casa, limpeza por cômodo, limpeza por zona, "limpar aqui" e zonas proibidas. *(ver D-01)*
- Agendamento de limpezas. *(ver D-02)*
- Mais de um robô e outros modelos ou fabricantes.
- Obtenção automática do identificador e da chave local: o dono os informa.
- Busca fora das redes locais em que o servidor está, ou em redes maiores que 254 endereços.
- Pacotes de voz, redefinição do mapa e atualização de firmware.
- Qualquer comunicação com a nuvem da Tuya: tudo é feito pela rede local.
- App Android e kiosk.

## 5. Conceitos

- **Chave local**: segredo de 16 caracteres que o robô exige para aceitar conexão pela rede local. O dono a obtém uma vez na plataforma do fabricante; ela muda se o robô for pareado de novo.
- **Estado**: o que o robô está fazendo — parado, limpando, pausado, voltando para a base, carregando, carga completa, esvaziando o reservatório, dormindo, em controle manual, em falha.
- **Peça de desgaste**: filtro, escova principal e escova lateral, cada uma com um contador de vida útil mantido pelo robô.

## 6. Regras de negócio

### Conexão

- **RN-01** — O robô é identificado pelo identificador do dispositivo e pela chave local, informados pelo dono em Configurações. O endereço na rede é encontrado pelo servidor (RN-21) ou informado à mão.
- **RN-02** — A chave local é só de escrita: é gravada no servidor e nunca é devolvida por nenhuma resposta da API, nem escrita em log. A API informa apenas se há uma chave gravada.
- **RN-03** — O servidor fala com o robô somente pela rede local. Nenhuma função desta spec depende de internet.
- **RN-04** — Sem robô configurado, a tela do robô explica o que falta e leva às Configurações; nenhum comando é oferecido.
- **RN-05** — Robô configurado mas inalcançável (desligado, fora da rede, endereço errado) e chave local recusada são falhas diferentes, cada uma com mensagem própria.

### Estado

- **RN-06** — O estado mostrado é o que o robô informa, não o que o servidor supõe: depois de um comando, a tela só muda de estado quando o robô confirma.
- **RN-07** — Com a tela aberta, o estado é atualizado sozinho a cada 5 segundos, sem piscar e sem trocar o conteúdo por esqueleto ([00/RN-31](00-padroes-transversais.md)). Com a aba em segundo plano, a atualização para.
- **RN-08** — Falha em uma atualização automática mantém o último estado na tela, marcado como desatualizado com o horário da última leitura; não vira tela de erro.
- **RN-09** — Quando o robô informa falha, a tela mostra o código da falha em destaque. *(ver D-03)*

### Comandos

- **RN-10** — Comandos: iniciar limpeza da casa inteira, pausar, continuar, parar, voltar para a base e localizar (o robô apita).
- **RN-11** — Só são oferecidos os comandos que fazem sentido no estado atual: "Iniciar" em `idle`, `charging`, `charged` e `sleeping`; "Pausar" em `cleaning`; "Continuar" em `paused`; "Parar" em `cleaning`, `paused` e `returning`; "Voltar para a base" em qualquer estado, menos `charging`, `charged`, `returning` e `emptying`; "Localizar" sempre. Em `unknown` todos são oferecidos.
- **RN-12** — Todo comando segue [00 §6](00-padroes-transversais.md): o botão fica desabilitado até o robô responder, clique duplo não envia duas vezes, e a falha mostra a mensagem do servidor.
- **RN-13** — Controle manual: frente, ré, esquerda, direita e parar. O robô anda enquanto o botão está pressionado e para ao soltar; sair da tela ou perder o foco da aba também manda parar.

### Ajustes

- **RN-14** — Ajustes de limpeza: potência de sucção (silenciosa, automática, forte, máxima), nível de água do pano (baixo, médio, alto), reforço de sucção em tapete, retomar a limpeza depois de recarregar.
- **RN-15** — Ajustes do aparelho: volume da voz (0 a 10) e não perturbe.
- **RN-16** — Base de esvaziamento: ligar ou desligar o esvaziamento automático e escolher a frequência.
- **RN-17** — Um ajuste só é mostrado se o robô o informa; ajuste que o robô não informa não aparece como desligado nem com valor inventado.
- **RN-18** — Alterar um ajuste envia somente aquele ajuste ao robô e o valor exibido passa a ser o que o robô confirmar.

### Busca na rede

- **RN-21** — "Procurar robô" faz o servidor percorrer os endereços das redes locais em que ele próprio está, procurando aparelhos que atendem na porta do protocolo, e reconhece o robô como o aparelho que aceita o identificador e a chave gravados. O endereço encontrado é gravado.
- **RN-22** — A busca só percorre redes privadas das interfaces do servidor, até 254 endereços por rede, e termina em até 5 segundos. É a exceção a [00/RN-39](00-padroes-transversais.md): a tela espera a resposta, com indicador no botão.
- **RN-23** — Quando o robô deixa de responder no endereço gravado, o servidor procura de novo sozinho, no máximo uma vez por minuto, e passa a usar o endereço novo. O dono não precisa fixar o endereço do robô no roteador.
- **RN-24** — Busca sem resultado diz que nenhum aparelho aceitou o identificador e a chave, e lembra as três causas: robô desligado, robô em outra rede, chave trocada por novo pareamento.

### Desgaste e totais

- **RN-19** — Para cada peça de desgaste que o robô informa, a tela mostra a vida útil e permite zerar o contador depois da troca. Zerar pede confirmação ([00/RN-37](00-padroes-transversais.md)), dizendo que não é reversível.
- **RN-20** — Totais acumulados: tempo total de limpeza, área total e número de limpezas.

## 7. Requisitos por tela

### Robô aspirador (`/vacuum`)

- **RF-01** — Item "Robô aspirador" no grupo Ferramentas do menu ([01/RF-01](01-shell-e-navegacao.md)).
- **RF-02** — Bloco de estado: nome do estado, bateria em porcentagem, e, durante a limpeza, tempo decorrido e área limpa. Em falha, o código da falha.
- **RF-03** — Bloco de comandos conforme RN-10 e RN-11.
- **RF-04** — Bloco de controle manual conforme RN-13, utilizável por toque e por teclado (setas).
- **RF-05** — Bloco de ajustes conforme RN-14 a RN-18.
- **RF-06** — Bloco de desgaste conforme RN-19 e bloco de totais conforme RN-20.
- **RF-07** — Cada bloco tem seu próprio estado de carregamento e erro ([00/RN-32](00-padroes-transversais.md)): a falha dos totais não esconde os comandos.
- **RF-08** — Em tela estreita os blocos empilham nesta ordem: estado, comandos, controle manual, ajustes, desgaste, totais. Alvos de toque de no mínimo 44 × 44 px.

### Configurações → Robô aspirador

- **RF-09** — Campos: identificador do dispositivo, chave local e endereço na rede (opcional). O campo da chave nunca mostra a chave gravada; indica apenas que existe uma, e deixá-lo vazio ao salvar mantém a atual.
- **RF-10** — "Procurar robô" (RN-21) preenche o endereço com o que foi encontrado. "Testar conexão" usa os valores gravados e responde com sucesso ou com a falha específica (RN-05).
- **RF-11** — Texto explicando onde o dono encontra o identificador e a chave local, e que a chave muda quando o robô é pareado de novo.

## 8. Contrato HTTP

Todas as rotas são a criar. Tempos em segundos, áreas em m².

| Método | Rota | Uso |
|---|---|---|
| GET | `/vacuum/settings` | `{"address", "device_id", "has_local_key"}` |
| PUT | `/vacuum/settings` | corpo `{"address", "device_id", "local_key"}`; `local_key` ausente ou vazio mantém a gravada; `address` pode ser vazio. `200` com o mesmo formato do GET |
| POST | `/vacuum/connection-test` | sem corpo. `204` quando o robô respondeu |
| POST | `/vacuum/discovery` | sem corpo. `200` com `{"address"}`, já gravado; `404` quando nenhum aparelho aceitou o identificador e a chave |
| GET | `/vacuum/status` | `{"state", "battery_percent", "cleaning_seconds", "cleaning_area", "fault_codes", "read_at"}` |
| POST | `/vacuum/commands` | corpo `{"command"}` com `start`, `pause`, `resume`, `stop`, `dock` ou `locate`. `204` |
| POST | `/vacuum/remote-control` | corpo `{"direction"}` com `forward`, `backward`, `left`, `right` ou `stop`. `204` |
| GET | `/vacuum/preferences` | `{"suction", "water_level", "volume", "is_do_not_disturb", "is_carpet_boost", "is_resume_after_charge", "is_auto_empty", "auto_empty_frequency"}`; campo que o robô não informa vem `null` |
| PATCH | `/vacuum/preferences` | corpo com um ou mais campos do GET. `200` com o formato do GET |
| GET | `/vacuum/consumables` | `{"items": [{"part", "seconds", "max_seconds"}]}` com `part` em `filter`, `main_brush`, `side_brush`; só as peças que o robô informa |
| POST | `/vacuum/consumables/:part/reset` | `204` |
| GET | `/vacuum/totals` | `{"cleaning_seconds", "cleaning_area", "cleaning_count"}` |

`state` é um de: `idle`, `cleaning`, `paused`, `returning`, `charging`, `charged`, `emptying`, `sleeping`, `remote_control`, `fault`. Estado que o robô informa e não está na correspondência do anexo vem como `unknown` e a tela o mostra como "estado desconhecido", sem supor que o robô está parado.

## 9. Estados e erros

| Situação | Status | O que o usuário vê |
|---|---|---|
| Robô não configurado (falta identificador ou chave) | `409` | o que falta e link para Configurações (RN-04) |
| Busca sem resultado | `404` | mensagem de RN-24 |
| Robô inalcançável ou sem resposta em 5 s, mesmo depois da nova busca (RN-23) | `502` | "não foi possível falar com o robô", com "Tentar de novo" |
| Chave local recusada pelo robô | `502` | mensagem dizendo para conferir a chave, lembrando que ela muda ao parear de novo |
| Comando, direção, peça ou valor de ajuste desconhecido | `400` | mensagem do servidor |
| Endereço, identificador ou chave em formato inválido ao salvar | `400` | mensagem no campo |

## 10. Critérios de aceite

- [ ] **CA-01** — Sem robô configurado, `/vacuum` explica o que falta e leva às Configurações.
- [ ] **CA-02** — Salvar endereço, identificador e chave e usar "Testar conexão" responde sucesso; com a chave errada e com o endereço errado, cada caso mostra sua mensagem.
- [ ] **CA-03** — Nenhuma resposta da API e nenhum arquivo de log contém a chave local.
- [ ] **CA-04** — Com o robô na base, a tela mostra o estado e a bateria que o próprio robô informa.
- [ ] **CA-05** — "Iniciar" faz o robô sair para limpar e a tela passa a mostrar "limpando", com tempo e área avançando, sem recarregar a página.
- [ ] **CA-06** — "Pausar" para o robô; "Continuar" retoma; "Voltar para a base" o leva à base e o estado acompanha até "carregando".
- [ ] **CA-07** — "Localizar" faz o robô apitar.
- [ ] **CA-08** — No controle manual o robô anda enquanto o botão está pressionado e para ao soltar.
- [ ] **CA-09** — Mudar a potência de sucção durante a limpeza muda o ruído do robô e o valor persiste ao recarregar a tela.
- [ ] **CA-10** — Desligar o robô da tomada com a tela aberta: o último estado fica visível, marcado como desatualizado; ao religar, a tela volta a atualizar sozinha.
- [ ] **CA-11** — Zerar o contador de uma peça pede confirmação e o valor mostrado muda para o que o robô informa depois.
- [ ] **CA-12** — Com o cabo de internet do roteador desconectado, estado e comandos continuam funcionando.
- [ ] **CA-13** — A tela é utilizável no navegador do celular, sem rolagem horizontal.
- [ ] **CA-14** — Com só identificador e chave gravados, "Procurar robô" encontra o endereço e "Testar conexão" passa a responder sucesso.
- [ ] **CA-15** — Trocar o endereço do robô no roteador com a tela aberta: em até dois minutos a tela volta a atualizar sozinha, sem ninguém mexer nas Configurações.
- [ ] **CA-16** — Com a chave errada, "Procurar robô" não grava endereço nenhum e mostra a mensagem de busca sem resultado.

## 11. Decisões em aberto

- **D-01** — Mapa, cômodos e zonas. O robô troca esses dados em formato binário proprietário e o mapa trafega pela nuvem do fabricante. *Provisório: fora de escopo.*
- **D-02** — Agendar limpezas pelo KuraNAS. *Provisório: fora de escopo.*
- **D-03** — Descrição das falhas. O robô informa a falha como um número de 1 a 30, sem texto. *Provisório: a tela mostra "Falha N"; a descrição entra quando os códigos forem conferidos com o manual.*
- **D-04** — Sentido do contador de desgaste (tempo já usado ou tempo restante). Na leitura de 2026-11-16 a escova principal informava 803013 de 1080000 segundos; só observar o valor antes e depois de uma limpeza decide. *Provisório: a tela mostra o valor e o máximo em horas, sem dizer "restante" nem "usado", até a conferência.*
- **D-05** — Peças que o robô não informa. Na mesma leitura só a escova principal veio; filtro e escova lateral não. *Provisório: RN-17 — a peça não aparece.*

## 12. Anexo — pontos de dados do robô

Lidos do próprio aparelho em 2026-11-16 (protocolo Tuya 3.3, porta 6668). Leitura conferida no sistema rodando; **nenhuma escrita foi conferida ainda**.

| Ponto | Significado | Usado em |
|---|---|---|
| 101 | limpeza da casa inteira (verdadeiro inicia, falso para) | `start`, `stop` |
| 102 | pausa (verdadeiro pausa, falso continua) | `pause`, `resume` |
| 103 | voltar para a base | `dock` |
| 105 | estado: `idle`, `pointing`, `areaing`, `totaling`, `sweep`, `mop`, `fault`, `pause`, `chargring`, `tocharge`, `fullcharge`, `remotectl`, `dormant`, `curpointing`, `selectroom`, `dustcenterworking` | `state` |
| 106 | bateria (%) | `battery_percent` |
| 107, 108 | tempo (s) e área (m²) da limpeza atual | `cleaning_seconds`, `cleaning_area` |
| 109 | sucção: `quiet`, `auto`, `strong`, `max` (e `mop`, só pano) | `suction` |
| 110 | água: `low`, `mid`, `high` | `water_level` |
| 111 | controle manual: `forward`, `backward`, `left`, `right`, `stop` | `/vacuum/remote-control` |
| 112 | localizar | `locate` |
| 113, 114 | não perturbe, volume (0–10) | `is_do_not_disturb`, `volume` |
| 115 | zerar peça: `resetsidebrush`, `resetmainbrush`, `resetfilter` | `/vacuum/consumables/:part/reset` |
| 116, 117, 118 | totais: tempo (s), área (m²), número de limpezas | `/vacuum/totals` |
| 119, 120, 121 | escova lateral, escova principal, filtro (s) | `/vacuum/consumables` |
| 122 | falhas (mapa de bits, 1 a 30) | `fault_codes` |
| 135, 136 | esvaziamento automático e frequência (`0` a `3`) | `is_auto_empty`, `auto_empty_frequency` |
| 137 | reforço em tapete | `is_carpet_boost` |
| 151 | retomar depois de recarregar | `is_resume_after_charge` |

Correspondência de estado: `idle` → `idle`; `sweep`, `mop`, `totaling`, `pointing`, `areaing`, `curpointing`, `selectroom` → `cleaning`; `pause` → `paused`; `tocharge` → `returning`; `chargring` → `charging`; `fullcharge` → `charged`; `dustcenterworking` → `emptying`; `dormant` → `sleeping`; `remotectl` → `remote_control`; `fault` → `fault`.

## 13. Rastreabilidade

Feature nova, pedida pelo dono em 2026-11-16; sem item nos boards.
