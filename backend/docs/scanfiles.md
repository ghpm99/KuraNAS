# Scan Files Job System

```mermaid
flowchart TD
    A[Worker Bootstrap] --> B[Create startup_scan job]
    B --> C[Job Scheduler Tick]
    C --> D[scan_filesystem step]
    D --> E[diff_against_db step]
    E --> F[mark_deleted step]

    G[Upload endpoint] --> H[Create upload_process job]
    H --> C
    C --> I[metadata]
    I --> J[checksum]
    J --> K[persist]
    K --> L[thumbnail]
    K --> M[playlist_index]

    N[Filesystem watcher] --> O[Debounce scopes]
    O --> P[Create fs_event/reindex_folder job]
    P --> C
```

## Notes
- `StartFileProcessingPipeline` monolith was retired.
- `ScanDirWorker` was retired.
- Legacy queue tasks (`ScanFiles` and `ScanDir`) are now adapters that enqueue jobs.
