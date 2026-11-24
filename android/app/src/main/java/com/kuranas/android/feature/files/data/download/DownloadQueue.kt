package com.kuranas.android.feature.files.data.download

import kotlinx.coroutines.flow.StateFlow

interface DownloadQueue {
    val entries: StateFlow<List<DownloadEntry>>

    /** Enfileira os pedidos na ordem recebida e devolve o id de cada entrada criada. */
    suspend fun enqueue(requests: List<DownloadRequest>): List<Long>
}
