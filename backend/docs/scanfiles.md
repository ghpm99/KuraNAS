graph TD
    subgraph "main.go"
        A[main()] --> B(StartFileProcessingPipeline(targetDirectory string))
    end

    subgraph "pipeline.go"
        B --> C{Canais e Sincronização};
        C --> D[fileWalkChannel chan FileWalk]
        C --> E[fileDtoChannel chan FileDto]
        C --> F[metadataProcessedChannel chan FileDto]
        C --> G[checksumCompletedChannel chan FileDto]
        C --> H[workerGroup *sync.WaitGroup]

        subgraph "1. Explorador de Arquivos"
            I[go StartDirectoryWalker(targetDirectory, fileWalkChannel, workerGroup)]
            I --> J[filepath.Walk(targetDirectory, walkCallback)]
            J --> K[walkCallback(filePath, fileInfo, err)]
            K --&gt; L[fileWalkChannel &lt;- FileWalk{filePath, fileInfo}]
        end

        subgraph "2. Conversor de DTO"
            M[go StartDtoConverterWorker(fileWalkChannel, fileDtoChannel, workerGroup)]
            M --&gt; N[for fileWalkItem := range fileWalkChannel]
            N --&gt; O[fileDto := convertToDto(fileWalkItem)]
            O --&gt; P[fileDtoChannel &lt;- fileDto]
        end

        subgraph "3. Processador de Metadados"
            Q[go StartMetadataWorker(fileDtoChannel, metadataProcessedChannel, workerGroup)]
            Q --&gt; R[for unprocessedFile := range fileDtoChannel]
            R --&gt; S{switch unprocessedFile.Format}
            S --&gt; T[case "image": extractImageMetadata(unprocessedFile)]
            S --&gt; U[case "video": extractVideoMetadata(unprocessedFile)]
            S --&gt; V[case "audio": extractAudioMetadata(unprocessedFile)]
            T --&gt; W[metadataProcessedChannel &lt;- unprocessedFile]
            U --&gt; W
            V --&gt; W
        end

        subgraph "4. Gerador de Checksum"
            X[go StartChecksumWorker(metadataProcessedChannel, checksumCompletedChannel, workerGroup)]
            X --&gt; Y[for fileToProcess := range metadataProcessedChannel]
            Y --&gt; Z[fileToProcess.CheckSum = generateSHA256(fileToProcess.Path)]
            Z --&gt; AA[checksumCompletedChannel &lt;- fileToProcess]
        end

        subgraph "5. Persistência de Dados (Goroutine Única)"
            AB[go StartDatabasePersistenceWorker(checksumCompletedChannel, workerGroup)]
            AB --&gt; AC[for finalizedFile := range checksumCompletedChannel]
            AC --&gt; AD[existingRecord := db.FindRecordByPath(finalizedFile.Path)]
            AD --&gt; AE{Registro já existe?}
            AE --&gt; |Sim| AF[UpdateFileRecord(finalizedFile, existingRecord)]
            AE --&gt; |Não| AG[CreateFileRecord(finalizedFile)]
        end

        B --&gt; I; B --&gt; M; B --&gt; Q; B --&gt; X; B --&gt; AB;

        L --&gt; N
        P --&gt; R
        W --&gt; Y
        AA --&gt; AC

        B --&gt; H[workerGroup.Wait()]
    end

    style C fill:#f9f,stroke:#333,stroke-width:2px
    style AB fill:#bbf,stroke:#333,stroke-width:2px
    style AE fill:#f9f,stroke:#333,stroke-width:2px