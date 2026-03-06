## Critérios de aceite — Refatoração do Sistema de Workers (KuraNAS)

### A. Arquitetura e Responsabilidades

1. **Separação entre Orquestração e Execução**

   - Existe um componente de **orquestração** (Job Orchestrator/Planner) responsável por criar Jobs e Steps (com dependências).
   - Workers executam **somente** Steps atômicos (uma responsabilidade por step) e **não** contêm lógica “de pipeline completa”.

2. **Workers especializados**

   - Cada worker/step tem uma responsabilidade clara (ex.: `scan`, `diff`, `metadata`, `checksum`, `persist`, `thumbnail`, `playlist_index`).
   - Não existe um worker “monolítico” que execute a pipeline inteira.

3. **Idempotência**

   - Reexecutar o mesmo Step para o mesmo arquivo/versão não pode corromper dados nem gerar duplicações.
   - Steps devem suportar “**skip**” quando o resultado já estiver atualizado.

4. **Centralização de checksum**

   - O checksum é calculado/atualizado por **um único Step/Worker** oficial (fonte única de verdade).
   - Nenhum outro fluxo (upload, watcher, startup) calcula checksum fora desse Step.

---

### B. Modelo de Jobs, Steps e Persistência de Estado

5. **Persistência de execução**

   - Jobs e Steps possuem estado persistido no banco: `queued`, `running`, `completed`, `failed`, `canceled`, `skipped` (para Step).
   - Reiniciar o backend não “apaga” o progresso/estado; o sistema retoma ou marca como interrompido de forma consistente.

6. **Rastreabilidade completa**

   - Todo processamento em background (startup scan, upload process, fs event) gera um `job_id` único.
   - O `job_id` permite consultar: status, steps, erros, timestamps.

7. **Erros por step**

   - Falhas são registradas por step com `last_error` e contagem de tentativas.
   - Um Step que falha não torna o sistema inconsistente: o Job marca `partial_fail` quando aplicável (ex.: thumbnail falha mas persistência ok).

---

### C. Progresso e Feedback ao Usuário

8. **Progresso baseado em estado (não em canal acoplado)**

   - O progresso é derivado do estado persistido do Job/Steps.
   - Não existe dependência obrigatória de “pipeline alimentando channel” para a UI funcionar.

9. **API de consulta de progresso**

   - Existem endpoints para:

     - Consultar Job por ID (status, progresso geral, timestamps).
     - Listar Jobs ativos/recentes.
     - Consultar Steps de um Job (status e erros).

   - A UI consegue exibir progresso com **polling** (WebSocket/SSE é opcional).

---

### D. Startup Scan (StartFileProcessingPipeline → Job)

10. **Startup scan ocorre como Job**

- O scan inicial do sistema (no boot) cria um Job do tipo `startup_scan`.
- O Job inclui steps mínimos: `scan_filesystem` + `diff_against_db` + processamento apenas de arquivos `new/modified`.

11. **Processamento incremental**

- Arquivos `unchanged` (sem mudança detectada) **não** são enviados para steps pesados (checksum, thumbnail, metadata).
- A detecção de mudança usa pelo menos `size` + `mtime` (ou equivalente) e só faz verificação mais cara quando necessário.

12. **Tratamento de deletados**

- Arquivos removidos do filesystem são refletidos no banco por Step específico (ex.: `mark_deleted`), sem exigir reprocessamento total.

---

### E. Watcher (startEntryPointWatcher)

13. **Watcher não executa pipeline**

- O watcher não dispara “pipeline completa”.
- Ele apenas converte eventos em Jobs (`fs_event`) com escopo apropriado.

14. **Debounce / agregação**

- Eventos em rajada (ex.: muitas alterações em segundos) são agregados por janela de tempo e/ou por pasta/arquivo.
- O sistema evita criar centenas de Jobs redundantes para a mesma alteração.

---

### F. Upload (request rápida + processamento em background)

15. **Upload não bloqueia processamento**

- A request de upload:

  - salva o arquivo localmente
  - cria um Job `upload_process`
  - retorna sucesso imediatamente com `job_id` e referência do arquivo (ex.: `file_id` ou path)

- O processamento pesado roda em background via Steps.

16. **Steps corretos no upload**

- O Job de upload inclui steps necessários (condicionais):

  - `metadata` (se aplicável)
  - `checksum` (sempre, conforme regra de negócio)
  - `persist` (quando necessário)
  - `thumbnail` (apenas para formatos suportados)
  - `playlist_index` (apenas para formatos suportados)

---

### G. Performance, Concorrência e Estabilidade

17. **Limite de concorrência por tipo de worker**

- Há configuração de concorrência separada por categoria (ex.: checksum, thumbnail, metadata).
- O sistema aplica backpressure (fila) sem criar goroutines ilimitadas.

18. **Prioridade**

- Jobs gerados por upload do usuário têm prioridade maior que jobs de startup scan.
- O sistema respeita prioridade na execução (ex.: `high > normal > low`).

19. **Skip inteligente**

- Thumbnail/metadata/checksum devem detectar “já processado e atualizado” e marcar Step como `skipped` sem custo elevado.

---

### H. Resiliência e Retentativas

20. **Retry com limites**

- Steps possuem retry limitado com backoff para erros transitórios.
- Erros permanentes não entram em loop infinito; Step finaliza como `failed` com erro registrado.

21. **Cancelamento (mínimo)**

- É possível cancelar Jobs longos (ao menos `startup_scan` e `reindex_folder`), marcando como `canceled`.
- Steps em execução respeitam `context cancellation` (ou equivalente) quando possível.

---

### I. Observabilidade e Qualidade

22. **Métricas e logs estruturados**

- Cada Job/Step registra logs com: `job_id`, `step_id`, `file_id/path`, duração.
- Métricas básicas disponíveis: tempo por step, throughput, jobs ativos.

23. **Cobertura de testes**

- Testes unitários cobrindo:

  - Orquestrador (criação de jobs/steps e dependências)
  - idempotência/skip
  - diff (changed vs unchanged)

- Testes de integração cobrindo:

  - fluxo de upload (retorno imediato + job processando)
  - startup scan (não reprocessa unchanged)
  - watcher (debounce + job creation)

24. **Compatibilidade com estrutura existente**

- Backend segue o padrão atual do projeto (handler → service → repository → queries `go:embed` → model/dto).
- Mudanças não quebram rotas existentes (ex.: comportamento esperado no frontend para acompanhar processamento).

---

### J. Critérios de “Done” do projeto

25. **Remoção/aposentadoria do worker monolítico**

- O `StartFileProcessingPipeline` não contém mais pipeline monolítica; é substituído por Job/Steps equivalentes.
- O `ScanDirWorker` descontinuado é removido ou substituído por Step reutilizável.

26. **Fluxos principais funcionando**

- 3 fluxos são suportados end-to-end via Job system:

  1. startup scan incremental
  2. upload assíncrono com job_id
  3. watcher criando jobs com debounce

27. **UI consegue acompanhar progresso**

- A UI (ou API consumer) consegue mostrar:

  - status do job
  - progresso por steps
  - erros

- Sem depender de implementação específica de “canal de progresso da pipeline”.