# Pendências que dependem do dono

Itens que o agente não resolve sozinho por exigirem instalar algo na máquina, decisão de produto ou ambiente que só o dono tem. O trabalho segue sem esperar; cada item diz o que ficou no lugar.

| Data | Item | Por que depende do dono | O que ficou no lugar |
|---|---|---|---|
| 2026-11-24 | Miniatura da primeira página de PDF na grade de Documentos (`07/RF-01`, task 45 item 4) | Rasterizar PDF exige uma ferramenta no servidor (`pdftoppm`/poppler ou `mutool`), que não está instalada; instalar programas na máquina só com autorização. | A grade mostra o ícone do formato. Com a ferramenta disponível, `/files/thumbnail/:id` passa a gerar a miniatura de PDF sem mudar o contrato. |
| 2026-11-24 | Verificação em aparelho real das tasks 37, 38 e 39 (Android) | Sem aparelho nem emulador nesta máquina; instalar emulador só com autorização. | Build, testes unitários e chamadas à API; o que não foi exercitado está anotado em cada CA. |
| 2026-11-24 | Avanço por `currentTime` no player web depois de trocar a faixa de áudio (task 43.1) | Sem navegador na máquina para exercitar a reprodução real. | Backend verificado (o remux começa no quadro-chave informado); comportamento do player coberto por testes. |
