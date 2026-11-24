# Pendências que dependem do dono

Itens que o agente não resolve sozinho por exigirem instalar algo na máquina, decisão de produto ou ambiente que só o dono tem. O trabalho segue sem esperar; cada item diz o que ficou no lugar.

| Data | Item | Por que depende do dono | O que ficou no lugar |
|---|---|---|---|
| 2026-11-24 | Miniatura da primeira página de PDF na grade de Documentos (`07/RF-01`, task 45 item 4) | Rasterizar PDF exige uma ferramenta no servidor (`pdftoppm`/poppler ou `mutool`), que não está instalada; instalar programas na máquina só com autorização. | A grade mostra o ícone do formato. Com a ferramenta disponível, `/files/thumbnail/:id` passa a gerar a miniatura de PDF sem mudar o contrato. |
| 2026-11-24 | Verificação em aparelho real das tasks 37, 38 e 39 (Android) | Sem aparelho nem emulador nesta máquina; instalar emulador só com autorização. | Build, testes unitários e chamadas à API; o que não foi exercitado está anotado em cada CA. |
| 2026-11-24 | Avanço por `currentTime` no player web depois de trocar a faixa de áudio (task 43.1) | Sem navegador na máquina para exercitar a reprodução real. | Backend verificado (o remux começa no quadro-chave informado); comportamento do player coberto por testes. |
| 2026-11-24 | Telas web de Documentos (task 45) num navegador real: abrir PDF de 300 páginas, ir para a página do trecho, Markdown/CSV | Sem navegador na máquina. | API verificada no backend local e telas cobertas por testes, alimentadas pela resposta real do handler. |
| 2026-11-24 | Publicar o próximo release com `SHA256SUMS` e atualizar a partir dele; tentar um pacote adulterado (task 32, `23/CA-11`) | Publicar release é ação do dono, e aplicar uma atualização troca o binário do servidor. | Recusa por checksum coberta por testes do `updater`. Atenção: servidor com este código recusa atualizar para release sem `SHA256SUMS`. |
| 2026-11-24 | Deixar o desligamento automático chegar ao fim sem cancelar (task 33, `23/CA-08`) | Executa `shutdown` na máquina. | Aviso e cancelamento verificados no backend local; o desligamento em si coberto por teste com execução simulada. |
| 2026-11-24 | Montar o WebDAV no Windows Explorer e com davfs2 (task 11) | Clientes não instalados nesta máquina; instalar só com autorização. | Verbos WebDAV verificados com `curl` no backend local depois da correção do roteamento. |
| 2026-11-24 | App legado no Galaxy Tab 2 (tasks 17 e 18) e watcher no `kuranas.exe` do Windows (task 06) | Exigem o tablet e o servidor Windows do dono. | Build e testes do app legado; payload do painel medido no backend local (< 4 KB); watcher verificado no Linux. |
