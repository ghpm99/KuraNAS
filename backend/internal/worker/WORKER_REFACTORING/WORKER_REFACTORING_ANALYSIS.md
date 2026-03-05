# Análise e Planejamento do Sistema de Workers

## Análise do Sistema Atual

### Workers Existentes

#### 1. **StartFileProcessingPipeline** (worker principal)

- **Problema**: É uma pipeline inteira implementada dentro de um worker
- **Responsabilidades**: Escaneamento, metadados, checksum, persistência, thumbnails
- **Performance**: Processamento pesado que leva minutos com CPU máxima
- **Acoplamento**: Alto acoplamento entre todas as etapas

#### 2. **ScanDirWorker** (descontinuado)

- **Status**: Não utilizado
- **Função**: Escaneamento de diretórios específicos
- **Problema**: Versão antiga superada pela pipeline

#### 3. **UpdateCheckSumWorker**

- **Status**: Parcialmente utilizado
- **Função**: Atualização de checksum de arquivos/diretórios
- **Problema**: Não centralizado como planejado originalmente

#### 4. **CreateThumbnailWorker**

- **Status**: Utilizado na pipeline
- **Função**: Geração de thumbnails para imagens e vídeos
- **Funcionamento**: Adequado

#### 5. **GenerateVideoPlaylistsWorker**

- **Status**: Utilizado ao final da pipeline
- **Função**: Geração de playlists inteligentes de vídeo
- **Funcionamento**: Adequado

#### 6. **File System Watcher**

- **Função**: Monitora mudanças no filesystem
- **Problema**: Dispara a pipeline completa para qualquer mudança

### Problemas Identificados

#### 1. **Arquitetural**

- **Monolítico**: [StartFileProcessingPipeline](cci:1://file://wsl.localhost/Ubuntu-24.04/mnt/wsl/PHYSICALDRIVE4/projects/KuraNAS/backend/internal/worker/fileProcessingPipeline.go:39:0-128:1) faz tudo
- **Baixa coesão**: Workers com múltiplas responsabilidades
- **Alto acoplamento**: Dependências diretas entre etapas

#### 2. **Performance**

- **Processamento desnecessário**: Arquivos não alterados são reprocessados
- **Bloqueio**: Pipeline síncrona bloqueia outros processos
- **Recursos**: Uso intensivo de CPU mesmo para arquivos unchanged

#### 3. **Manutenibilidade**

- **Complexidade**: Difícil entender e modificar a pipeline
- **Testes**: Lógica complexa dificulta testes unitários
- **Debugging**: Difícil isolar problemas em etapas específicas

#### 4. **Escalabilidade**

- **Upload de arquivos**: Processamento síncrono no upload
- **Paralelismo**: Limitado pela estrutura da pipeline
- **Feedback**: Não há progresso granular para o usuário

## Nova Arquitetura Proposta

### Workers Especializados

#### 1. **FileSystemScannerWorker**

```go
// Responsabilidade: Apenas escanear filesystem
// Output: Lista de arquivos alterados
type FileSystemScanResult struct {
    NewFiles      []string
    ModifiedFiles []string
    DeletedFiles  []string
}
```

#### 2. **FileMetadataWorker**

```go
// Responsabilidade: Extrair metadados de arquivos
// Input: File paths
// Output: FileDto com metadados
```

#### 3. **FileChecksumWorker**

```go
// Responsabilidade: Calcular checksum
// Input: FileDto
// Output: FileDto com checksum
```

#### 4. **FilePersistenceWorker**

```go
// Responsabilidade: Persistir no banco
// Input: FileDto completo
// Output: FileDto com ID
```

#### 5. **ThumbnailGenerationWorker**

```go
// Responsabilidade: Gerar thumbnails
// Input: FileID
// Output: Status da geração
```

#### 6. **PlaylistGenerationWorker**

```go
// Responsabilidade: Gerar playlists
// Input: Trigger (completion/changes)
// Output: Status da geração
```

### Sistema de Coordenação

#### 1. **TaskManager**

```go
type TaskManager struct {
    tasks        chan Task
    progress     chan ProgressUpdate
    workers      map[TaskType][]Worker
}
```

#### 2. **ProgressMonitor**

```go
type ProgressUpdate struct {
    TaskID    string
    Stage     string
    Progress  int
    FileCount int
    Errors    []string
}
```

#### 3. **FileChangeDetector**

```go
type FileChangeDetector struct {
    lastSnapshot map[string]FileMetadata
    checksumCache map[string]string
}
```

## Fluxo de Upload Refatorado

### Upload Service

```go
func (s *UploadService) UploadFile(file multipart.File, header *multipart.FileHeader, targetPath string) error {
    // 1. Salvar arquivo localmente
    savedPath, err := s.saveFileLocally(file, header, targetPath)
    if err != nil {
        return err
    }

    // 2. Disparar workers em paralelo
    taskID := s.taskManager.CreateTask("file_upload")

    go func() {
        // Metadata extraction
        s.taskManager.EnqueueTask(Task{
            Type: TaskTypeExtractMetadata,
            Data: savedPath,
            TaskID: taskID,
        })

        // Checksum calculation
        s.taskManager.EnqueueTask(Task{
            Type: TaskTypeCalculateChecksum,
            Data: savedPath,
            TaskID: taskID,
        })

        // Thumbnail generation (se aplicável)
        if s.isMediaFile(header.Filename) {
            s.taskManager.EnqueueTask(Task{
                Type: TaskTypeGenerateThumbnail,
                Data: savedPath,
                TaskID: taskID,
            })
        }
    }()

    // 3. Retornar sucesso imediatamente
    return nil
}
```

## Otimizações Propostas

### 1. **Detecção de Alterações**

```go
type FileChangeDetector struct {
    modTimeCache map[string]time.Time
    sizeCache    map[string]int64
    checksumCache map[string]string
}

func (f *FileChangeDetector) HasChanged(path string, info os.FileInfo) bool {
    cached, exists := f.modTimeCache[path]
    if !exists {
        return true // Novo arquivo
    }

    // Verificar modTime e size primeiro (rápido)
    if !info.ModTime().Equal(cached) {
        // Verificar checksum apenas se modTime mudou
        return f.verifyChecksumChange(path, info)
    }

    return false
}
```

### 2. **Pipeline Paralela**

```go
type ParallelPipeline struct {
    metadataWorkers   int
    checksumWorkers   int
    persistenceWorkers int
    thumbnailWorkers  int
}

func (p *ParallelPipeline) ProcessFile(filePath string) {
    var wg sync.WaitGroup

    // Metadata e checksum podem rodar em paralelo
    wg.Add(2)
    go func() {
        defer wg.Done()
        p.processMetadata(filePath)
    }()

    go func() {
        defer wg.Done()
        p.processChecksum(filePath)
    }()

    wg.Wait()

    // Persistência depende dos dois
    p.persistFile(filePath)

    // Thumbnail pode rodar após persistência
    if p.isMediaFile(filePath) {
        p.generateThumbnail(filePath)
    }
}
```

### 3. **Cache Inteligente**

```go
type ProcessingCache struct {
    metadataCache  map[string]time.Time
    checksumCache  map[string]time.Time
    thumbnailCache map[string]time.Time
}

func (c *ProcessingCache) NeedsProcessing(filePath string, processingType ProcessingType) bool {
    lastProcessed, exists := c.getCache(filePath, processingType)
    if !exists {
        return true
    }

    info, err := os.Stat(filePath)
    if err != nil {
        return true
    }

    return info.ModTime().After(lastProcessed)
}
```

## Sistema de Monitoramento

### 1. **Progress Channel**

```go
type ProgressMonitor struct {
    subscribers map[string]chan ProgressUpdate
    activeTasks map[string]*TaskStatus
}

type TaskStatus struct {
    ID           string
    Type         string
    Stage        string
    Progress     int
    TotalFiles   int
    ProcessedFiles int
    Errors       []string
    StartTime    time.Time
    EstimatedEnd time.Time
}
```

### 2. **WebSocket para Feedback**

```go
func (h *Handler) WebSocketProgress(c *gin.Context) {
    conn := websocket.Upgrader{}
    ws, err := conn.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }

    // Inscrever para atualizações de progresso
    progressChan := h.progressMonitor.Subscribe("upload_" + userID)

    for update := range progressChan {
        ws.WriteJSON(update)
    }
}
```

## Pontos de Melhoria Identificados

### 1. **Centralização de Checksum**

- **Problema**: Múltiplos lugares calculam checksum
- **Solução**: Worker dedicado com cache inteligente
- **Benefício**: Evita recálculo desnecessário

### 2. **Processamento Incremental**

- **Problema**: Reprocessamento completo do filesystem
- **Solução**: Detecção granular de mudanças
- **Benefício**: Processamento apenas do necessário

### 3. **Feedback Granular**

- **Problema**: Usuário não vê progresso do upload
- **Solução**: Canal de progresso por WebSocket
- **Benefício**: Experiência do usuário melhor

### 4. **Resiliência**

- **Problema**: Falha em uma etapa afeta todo o processo
- **Solução**: Workers independentes com retry
- **Benefício**: Sistema mais robusto

### 5. **Escalabilidade Horizontal**

- **Problema**: Pipeline síncrona limita throughput
- **Solução**: Workers distribuídos com load balancing
- **Benefício**: Processamento paralelo real

## Implementação Sugerida

1. Extrair workers especializados da pipeline
2. Implementar TaskManager
3. Criar sistema de progresso
4. Implementar detecção de alterações
5. Adicionar cache inteligente
6. Paralelizar processamento
7. Refatorar upload service
8. Implementar WebSocket progress
9. Migrar sistema de scan inicial
10. Dashboard de progresso
11. Métricas de performance
12. Sistema de alertas

## Conclusão

A refatoração proposta transforma o sistema atual monolítico em uma arquitetura modular, paralela e monitorável. Os principais benefícios são:

- **Performance**: Processamento apenas de arquivos alterados
- **Escalabilidade**: Workers paralelos e distribuídos
- **Manutenibilidade**: Workers especializados e testáveis
- **Experiência do Usuário**: Feedback em tempo real
- **Resiliência**: Falha isolada por worker

A implementação deve ser gradual, começando pela extração dos workers e evoluindo para um sistema completamente paralelo e monitorado.
