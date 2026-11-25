# 26 — Robô aspirador

**Prioridade:** P2 · **Plataformas:** web · **Herda:** [00](00-padroes-transversais.md)

## 1. Objetivo

Comandar o robô aspirador da casa e ver o que ele está fazendo a partir do KuraNAS, pela rede local, sem depender do aplicativo do fabricante nem da nuvem dele.

## 2. Cenários de uso

- O dono está no PC e quer mandar o robô limpar a casa sem pegar o celular.
- O robô está limpando e alguém vai atender uma ligação: pausa, depois continua.
- Quer saber se o robô já terminou, quanto de bateria tem e se parou por alguma falha.
- Quer ver no mapa da casa por onde o robô já passou e onde ele está agora.
- Derrubou farelo na cozinha: manda limpar só a cozinha, ou só aquele ponto do chão.
- O robô sumiu embaixo de um móvel: manda ele apitar.
- Trocou o filtro ou a escova e zera o contador de desgaste.

## 3. Escopo

Um robô, modelo KaBuM! Smart 900 (plataforma Tuya, protocolo local 3.3): cadastro da conexão, busca do robô na rede, estado atual, comandos de limpeza, controle manual, ajustes de limpeza, base de esvaziamento, desgaste de peças, totais acumulados, mapa da casa com o caminho da limpeza, limpeza por cômodo, "limpar aqui" e nome dos cômodos.

## 4. Fora de escopo

- Editar o mapa: dividir ou juntar cômodos, criar, mover ou apagar zonas proibidas e paredes virtuais. As zonas que o robô já tem só são mostradas. *(ver D-06)*
- Limpeza por zona desenhada pelo dono (retângulo arbitrário). *(ver D-06)*
- Apagar o mapa do robô. O servidor nunca envia esse comando (ponto 129).
- Agendamento de limpezas. *(ver D-02)*
- Mais de um robô e outros modelos ou fabricantes.
- Obtenção automática do identificador e da chave local: o dono os informa.
- Busca fora das redes locais em que o servidor está, ou em redes maiores que 254 endereços.
- Pacotes de voz, redefinição do mapa e atualização de firmware.
- Qualquer comunicação com a nuvem da Tuya além de ler o mapa (RN-25).
- App Android e kiosk.

## 5. Conceitos

- **Chave local**: segredo de 16 caracteres que o robô exige para aceitar conexão pela rede local. O dono a obtém uma vez na plataforma do fabricante; ela muda se o robô for pareado de novo.
- **Estado**: o que o robô está fazendo — parado, limpando, pausado, voltando para a base, carregando, carga completa, esvaziando o reservatório, dormindo, em controle manual, em falha.
- **Peça de desgaste**: filtro, escova principal e escova lateral, cada uma com um contador de vida útil mantido pelo robô.
- **Mapa**: planta da casa que o robô monta com o laser — uma grade de células de 5 cm em que cada célula é parede, piso de um cômodo, obstáculo ou desconhecido. O robô só a envia para a nuvem da Tuya; é de lá que o servidor a lê.
- **Cômodo**: região do piso que o robô separou no mapa, identificada por um número (1 a 126). O robô não informa nome; o nome é dado pelo dono no KuraNAS.
- **Caminho**: sequência de posições (x, y) em milímetros por onde o robô passou na limpeza em andamento, no mesmo sistema de coordenadas do mapa. O robô o envia pela rede local enquanto limpa.
- **Credenciais da nuvem**: identificador de acesso, segredo de acesso e região de um projeto Cloud em iot.tuya.com, com o serviço "Sweeping Robot Open Service" autorizado e o robô vinculado.

## 6. Regras de negócio

### Conexão

- **RN-01** — O robô é identificado pelo identificador do dispositivo e pela chave local, informados pelo dono em Configurações. O endereço na rede é encontrado pelo servidor (RN-21) ou informado à mão.
- **RN-02** — A chave local é só de escrita: é gravada no servidor e nunca é devolvida por nenhuma resposta da API, nem escrita em log. A API informa apenas se há uma chave gravada.
- **RN-03** — O servidor fala com o robô somente pela rede local. Nenhuma função desta spec depende de internet, exceto ler o mapa (RN-25): sem internet, o bloco do mapa falha e todo o resto continua funcionando.
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

### Mapa

- **RN-25** — O mapa é lido da nuvem da Tuya com as credenciais da nuvem informadas pelo dono em Configurações. É a única leitura feita na nuvem; nenhum comando passa por ela.
- **RN-26** — O segredo de acesso da nuvem segue RN-02: só de escrita, nunca devolvido por resposta da API nem escrito em log; a API informa apenas se há um gravado.
- **RN-27** — Sem credenciais da nuvem, o bloco do mapa explica o que falta e leva às Configurações; os demais blocos funcionam normalmente. Falha da nuvem (sem internet, credenciais recusadas, serviço não autorizado no projeto) aparece só no bloco do mapa, cada uma com mensagem própria; "serviço não autorizado" diz para autorizar o "Sweeping Robot Open Service" no projeto.
- **RN-28** — O mapa mostra as paredes, o piso de cada cômodo com uma cor própria, os obstáculos, a base de recarga, as zonas proibidas que o robô já tem (só exibidas), o caminho da limpeza em andamento e a posição atual do robô (o último ponto do caminho). Norte do mapa é o eixo y crescente.
- **RN-29** — O mapa é pedido ao abrir a tela e, enquanto o robô está limpando, de novo a cada 30 segundos, sem piscar ([00/RN-31](00-padroes-transversais.md)). O servidor guarda o último mapa lido por 30 segundos e não consulta a nuvem mais do que isso.
- **RN-30** — O caminho vem do robô pela rede local. Com o robô configurado, o servidor mantém uma conexão aberta com o robô para receber o caminho, e a refaz sozinho se ela cair. O caminho exibido é o da limpeza atual: quando o robô começa uma limpeza nova, o caminho anterior some. Com a tela aberta e o robô limpando, o caminho é atualizado junto com o estado (RN-07).
- **RN-31** — Pontos do caminho que chegam fora de ordem ou repetidos são posicionados pelo índice que o robô informa; o caminho nunca tem ponto duplicado nem fica fora de ordem.

### Limpeza pelo mapa

- **RN-32** — Limpar cômodos: o dono seleciona um ou mais cômodos (tocando no mapa ou na lista de cômodos) e manda "Limpar selecionados". O robô limpa só esses cômodos. Oferecido nos mesmos estados que "Iniciar" (RN-11).
- **RN-33** — Limpar aqui: o dono toca num ponto do piso e confirma; o robô vai até lá e limpa um quadrado de 1,5 m × 1,5 m centrado no ponto. Ponto em parede, obstáculo ou área desconhecida não é aceito. Oferecido nos mesmos estados que "Iniciar".
- **RN-34** — Comandos pelo mapa citam o mapa em que o dono escolheu (identificador do mapa). Se o robô refez o mapa desde então, o comando é recusado e a tela carrega o mapa novo, sem mandar o robô limpar um lugar errado.
- **RN-35** — Comandos pelo mapa seguem RN-12 e RN-06: a tela só mostra "limpando" quando o robô confirma.

### Nome dos cômodos

- **RN-36** — O robô não informa o nome dos cômodos. Cômodo sem nome aparece como "Cômodo N", com N o número do cômodo.
- **RN-37** — O dono dá ou troca o nome de um cômodo pela tela do robô. O nome fica gravado no KuraNAS, ligado ao mapa atual; nome vazio volta a "Cômodo N". Nome tem de 1 a 40 caracteres.
- **RN-38** — Quando o robô refaz o mapa (o identificador do mapa muda), os nomes do mapa anterior não são aplicados ao novo.

### Desgaste e totais

- **RN-19** — Para cada peça de desgaste que o robô informa, a tela mostra o tempo de uso desde a última troca e a vida útil prevista, em horas, e permite zerar o contador depois da troca. Zerar pede confirmação ([00/RN-37](00-padroes-transversais.md)), dizendo que não é reversível.
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
- **RF-08** — Em tela estreita os blocos empilham nesta ordem: estado, comandos, mapa, controle manual, ajustes, desgaste, totais. Alvos de toque de no mínimo 44 × 44 px.
- **RF-12** — Bloco do mapa conforme RN-27 a RN-35: o mapa ocupa a largura do bloco mantendo a proporção, sem rolagem horizontal em nenhuma largura de tela. Tocar num cômodo o seleciona ou desseleciona (destaque visível); tocar e escolher "Limpar aqui" envia RN-33. Botão "Limpar selecionados" (com a quantidade) só aparece com cômodo selecionado.
- **RF-13** — Lista de cômodos ao lado do mapa (abaixo dele em tela estreita): nome (RN-36), área em m², seleção (a mesma do mapa) e ação de renomear (RN-37). Utilizável por teclado: a lista é a alternativa ao toque no mapa.
- **RF-14** — Legenda com as cores de parede, obstáculo, zona proibida, base, caminho e robô.

### Configurações → Robô aspirador

- **RF-09** — Campos: identificador do dispositivo, chave local e endereço na rede (opcional). O campo da chave nunca mostra a chave gravada; indica apenas que existe uma, e deixá-lo vazio ao salvar mantém a atual.
- **RF-10** — "Procurar robô" (RN-21) preenche o endereço com o que foi encontrado. "Testar conexão" usa os valores gravados e responde com sucesso ou com a falha específica (RN-05).
- **RF-11** — Texto explicando onde o dono encontra o identificador e a chave local, e que a chave muda quando o robô é pareado de novo.
- **RF-15** — Seção "Mapa pela nuvem": identificador de acesso, segredo de acesso e região (`us`, `eu`, `cn`, `in`). O segredo segue a mesma regra da chave (RF-09). "Testar nuvem" usa os valores gravados e responde sucesso ou a falha específica (RN-27). Texto explicando que as credenciais vêm de um projeto Cloud em iot.tuya.com, com o "Sweeping Robot Open Service" autorizado e o robô vinculado, e que a região é o Data Center do projeto.

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
| GET | `/vacuum/cloud-settings` | `{"access_id", "region", "has_access_secret"}` |
| PUT | `/vacuum/cloud-settings` | corpo `{"access_id", "access_secret", "region"}`; `access_secret` ausente ou vazio mantém o gravado. `200` com o formato do GET |
| POST | `/vacuum/cloud-connection-test` | sem corpo. `204` quando a nuvem devolveu o mapa |
| GET | `/vacuum/map` | `{"map_id", "width", "height", "resolution", "origin_x", "origin_y", "cells", "dock", "forbidden_zones", "read_at"}` — ver abaixo |
| GET | `/vacuum/rooms` | `{"map_id", "items": [{"id", "name", "area"}]}`; `name` é `null` quando o dono não deu nome; `area` em m² |
| PUT | `/vacuum/rooms/:id` | corpo `{"map_id", "name"}`; `name` vazio apaga o nome. `200` com o item |
| GET | `/vacuum/path` | `{"path_id", "points": [[x, y], …], "position"}`; `position` é `{"x", "y"}` do último ponto ou `null`; `path_id` `null` e `points` vazio quando não há caminho |
| POST | `/vacuum/room-cleaning` | corpo `{"map_id", "room_ids": [n, …]}`, ao menos um cômodo existente no mapa. `204` |
| POST | `/vacuum/spot-cleaning` | corpo `{"map_id", "x", "y"}` em mm, sobre piso de cômodo. `204` |

Mapa (`GET /vacuum/map`): coordenadas em milímetros, no mesmo sistema do caminho. `resolution` é o lado da célula em mm (50). `origin_x`/`origin_y` são as coordenadas do canto da célula (0, 0). `cells` é a grade em base64, `width × height` bytes, linha a linha começando pela linha de menor y; a célula (coluna `c`, linha `r`) cobre x de `origin_x + c × resolution` a `+ resolution`, e y igual com `r`. Valor da célula: `0` parede, `1`–`126` piso do cômodo com esse número, `127` desconhecido, `255` obstáculo. `dock` é `{"x", "y"}` ou `null`. `forbidden_zones` é `[{"vertices": [[x, y], …]}]`.

`map_id` divergente do mapa atual do robô em `PUT /vacuum/rooms/:id`, `/vacuum/room-cleaning` ou `/vacuum/spot-cleaning` responde `409` (RN-34).

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
| Credenciais da nuvem não configuradas (mapa, cômodos, comandos pelo mapa) | `409` | o que falta e link para Configurações (RN-27) |
| Nuvem inalcançável ou credenciais recusadas | `502` | mensagem própria de cada caso |
| Serviço "Sweeping Robot Open Service" não autorizado no projeto | `502` | mensagem dizendo para autorizar o serviço (RN-27) |
| Mapa mudou desde que a tela o carregou | `409` | mensagem dizendo que o mapa foi atualizado; a tela recarrega o mapa (RN-34) |
| Cômodo inexistente no mapa, ponto fora do piso, nome com mais de 40 caracteres | `400` | mensagem do servidor |

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
- [x] **CA-17** — Com as credenciais da nuvem gravadas, `/vacuum` mostra o mapa da casa com as paredes, cada cômodo numa cor e a base, na mesma forma da planta real.
- [ ] **CA-18** — Sem credenciais da nuvem, o bloco do mapa explica o que falta e leva às Configurações; estado, comandos e os outros blocos funcionam.
- [x] **CA-19** — Nenhuma resposta da API e nenhum arquivo de log contém o segredo de acesso da nuvem.
- [ ] **CA-20** — Com o serviço do robô não autorizado no projeto da nuvem, o bloco do mapa mostra a mensagem que manda autorizá-lo.
- [x] **CA-21** — Iniciar uma limpeza com a tela aberta: a posição do robô e o caminho avançam sobre o mapa sem recarregar a página, sobre o piso e não atravessando paredes.
- [x] **CA-22** — Selecionar um cômodo e "Limpar selecionados": o robô vai até aquele cômodo e limpa lá; o estado passa a "limpando".
- [x] **CA-23** — Tocar num ponto do piso e "Limpar aqui": o robô vai até o ponto e limpa em volta dele.
- [x] **CA-24** — Dar nome a um cômodo: o nome aparece no mapa e na lista e continua depois de recarregar a página.
- [ ] **CA-25** — No navegador do celular, o mapa cabe na largura da tela sem rolagem horizontal e o toque seleciona o cômodo certo.
- [ ] **CA-26** — Sem internet: o bloco do mapa mostra a falha da nuvem; estado, comandos e caminho continuam funcionando.

## 11. Decisões em aberto

- **D-01** — *Decidida em 2026-11-25.* Mapa, cômodos e "limpar aqui". Conferido no robô: pela rede local ele envia só o caminho (ponto 123); a grade com paredes e cômodos vai só para a nuvem da Tuya (`GET /v1.0/users/sweepers/file/{device_id}/realtime-map`, item `map_type` 0, JSON com a grade comprimida em LZ4). A limpeza por cômodo e por ponto é aceita pela rede local (ponto 127). O dono escolheu o modo híbrido: mapa lido da nuvem, caminho e comandos pela rede local (RN-25 a RN-38).
- **D-02** — Agendar limpezas pelo KuraNAS. *Provisório: fora de escopo.*
- **D-03** — Descrição das falhas. O robô informa a falha como um número de 1 a 30, sem texto. *Provisório: a tela mostra "Falha N"; a descrição entra quando os códigos forem conferidos com o manual.*
- **D-04** — *Decidida em 2026-11-16.* O contador de desgaste é tempo de uso, não tempo restante: depois de uma limpeza de 1792 segundos a escova principal foi de 803013 para 804805, o mesmo avanço do tempo total de limpeza. `seconds` é o uso e `max_seconds` a vida útil prevista (RN-19).
- **D-06** — Editar o mapa (dividir e juntar cômodos, zonas proibidas, paredes virtuais) e limpar zona desenhada. Nenhum desses comandos foi visto na rede local, e a documentação aberta não os descreve. *Provisório: fora de escopo; as zonas existentes só são mostradas. Próximo passo: capturar o que o app do fabricante envia ao editar.*
- **D-07** — Ponto 128. Muda durante a limpeza (vistos 0, 6, 22, 24, 25) sem falha nem travamento — provavelmente aviso de voz ou evento. *Provisório: ignorado.*
- **D-08** — Fim da limpeza por cômodo. No teste do cômodo 2 o robô voltou à base sozinho depois de cerca de 2 minutos, sem ter travado. *Provisório: a tela mostra o que o robô informa (RN-06); conferir com uma limpeza completa de cômodo.*
- **D-09** — Mapa próprio, por privacidade. Conferido em 2026-11-25: o robô envia a planta da casa e os percursos para a nuvem da Tuya, que os guarda (armazenamento nos EUA). O dono quer poder bloquear a internet do robô sem perder o mapa. Pela rede local só chega a posição do robô (ponto 123), o que permite saber por onde ele passou, mas não onde estão as paredes. *Provisório: próxima melhoria, nesta ordem:* (1) guardar no KuraNAS a última planta lida da nuvem e servir o mapa a partir dela quando a nuvem não responder; (2) acumular o caminho de todas as limpezas num mapa de cobertura próprio, mostrado sobre a planta guardada, e aprender o contorno de cada cômodo pelas limpezas por cômodo mandadas pelo KuraNAS; (3) com isso, o dono pode bloquear a internet do robô no roteador, e a nuvem passa a ser lida só quando ele pedir para atualizar a planta. Pendente: como perceber, sem a nuvem, que o robô refez o mapa (o sistema de coordenadas muda). O contrato de `/vacuum/map`, `/vacuum/rooms` e `/vacuum/path` não muda; muda a origem dos dados.
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
| 104 | modo atual (`smart`, `selectroom`, `pose`, `backcharge`, …) | não usado |
| 123 | caminho (base64 de JSON, `infoType` 21011: `pathID`, `startPos`, `pointCounts`, `totalPoints`, `posArray` em mm). Enviado só pelo robô, a cada ~2 s durante a limpeza | `/vacuum/path` |
| 127 | comandos pelo mapa (base64 de JSON, ver abaixo); o robô ecoa o comando com `"message": "ok"` | `/vacuum/room-cleaning`, `/vacuum/spot-cleaning` |
| 128 | aviso (D-07) | não usado |
| 129 | apagar o mapa | **nunca enviado** |

Comando pelo mapa (ponto 127, conferido no robô em 2026-11-25): `{"infoType": 30000, "dInfo": {"ts": "<ms>", "userId": "0"}, "data": {"cmds": [{"infoType": 21023, "data": {"cleanId": [-3], "mapId": <map_id>, "segmentId": [<cômodos>], "extraAreas": []}}, {"infoType": 21005, "data": {"mode": "reAppointClean"}}], "mainCmds": [21005]}}`. Para "limpar aqui", sem `segmentId` e com `extraAreas: [{"id": 100, "mode": "point", "active": "depth", "name": "", "tag": "room", "vertexs": [4 cantos do quadrado de 1500 mm, em mm]}]`. O robô entra em `selectroom` (cômodo) ou `pointing` (ponto).

Mapa na nuvem (conferido em 2026-11-25): o arquivo do item `map_type` 0 é JSON com `data.width`, `data.height`, `data.resolution` (0,05 m), `data.x_min`/`data.y_min` (m), `data.map` (base64 de bloco LZ4 sem cabeçalho, `width × height` bytes descomprimidos, linha 0 = menor y), `data.mapId`, `data.chargeHandlePos` (mm) e `data.area` (zonas: `active: "forbid"`, `vertexs` em mm). A assinatura das chamadas segue o método HMAC-SHA256 da Tuya (token em `/v1.0/token?grant_type=1`).

Correspondência de estado: `idle` → `idle`; `sweep`, `mop`, `totaling`, `pointing`, `areaing`, `curpointing`, `selectroom` → `cleaning`; `pause` → `paused`; `tocharge` → `returning`; `chargring` → `charging`; `fullcharge` → `charged`; `dustcenterworking` → `emptying`; `dormant` → `sleeping`; `remotectl` → `remote_control`; `fault` → `fault`.

## 13. Rastreabilidade

Feature nova, pedida pelo dono em 2026-11-16; sem item nos boards. Mapa, cômodos e "limpar aqui" pedidos pelo dono em 2026-11-25 ("o maior problema do robô hoje").
