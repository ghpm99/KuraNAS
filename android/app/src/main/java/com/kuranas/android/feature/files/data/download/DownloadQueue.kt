package com.kuranas.android.feature.files.data.download

import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.util.concurrent.atomic.AtomicLong

/** Fila de downloads executados um por vez. A task 37 troca a implementação por uma que sobrevive ao app fechado. */
interface DownloadQueue {
    val entries: StateFlow<List<DownloadEntry>>

    /** Enfileira os pedidos na ordem recebida e devolve o id de cada entrada criada. */
    fun enqueue(requests: List<DownloadRequest>): List<Long>
}

class InMemoryDownloadQueue(
    scope: CoroutineScope,
    private val executor: DownloadExecutor,
    private val notifier: DownloadNotifier,
) : DownloadQueue {

    private val mutableEntries = MutableStateFlow<List<DownloadEntry>>(emptyList())
    override val entries: StateFlow<List<DownloadEntry>> = mutableEntries.asStateFlow()

    private val waitingEntryIds = Channel<Long>(Channel.UNLIMITED)
    private val lastEntryId = AtomicLong(0)

    init {
        scope.launch {
            for (entryId in waitingEntryIds) run(entryId)
        }
    }

    override fun enqueue(requests: List<DownloadRequest>): List<Long> {
        val newEntries = requests.map { DownloadEntry(lastEntryId.incrementAndGet(), it, DownloadState.Queued) }
        mutableEntries.update { it + newEntries }
        newEntries.forEach { waitingEntryIds.trySend(it.id) }
        return newEntries.map { it.id }
    }

    private suspend fun run(entryId: Long) {
        val request = mutableEntries.value.first { it.id == entryId }.request
        val initialProgress = DownloadProgress(receivedBytes = 0, totalBytes = request.sizeBytes)
        publish(entryId, DownloadState.Running(initialProgress))
        notifier.showProgress(entryId.toInt(), request, initialProgress)
        val outcome = executeSafely(request) { progress ->
            publish(entryId, DownloadState.Running(progress))
            notifier.showProgress(entryId.toInt(), request, progress)
        }
        when (outcome) {
            is DownloadOutcome.Completed -> {
                publish(entryId, DownloadState.Completed(outcome.uri, outcome.mimeType))
                notifier.showCompleted(entryId.toInt(), request, outcome.uri, outcome.mimeType)
            }
            is DownloadOutcome.Failed -> {
                publish(entryId, DownloadState.Failed(outcome.failure))
                notifier.showFailed(entryId.toInt(), request, outcome.failure)
            }
        }
    }

    private suspend fun executeSafely(
        request: DownloadRequest,
        onProgress: (DownloadProgress) -> Unit,
    ): DownloadOutcome = try {
        executor.execute(request, onProgress)
    } catch (cancellation: CancellationException) {
        throw cancellation
    } catch (unexpected: Exception) {
        DownloadOutcome.Failed(DownloadFailure.StorageUnavailable)
    }

    private fun publish(entryId: Long, state: DownloadState) {
        mutableEntries.update { current ->
            current.map { if (it.id == entryId) it.copy(state = state) else it }
        }
    }
}
