package com.kuranas.android.feature.files.data.download

import android.net.Uri
import io.mockk.mockk
import io.mockk.verify
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.delay
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class InMemoryDownloadQueueTest {

    private val notifier = mockk<DownloadNotifier>(relaxed = true)
    private val uri = mockk<Uri>()

    private fun request(fileId: String) = DownloadRequest(fileId, "arquivo $fileId.bin", 1000)

    @Test
    fun `downloads run one at a time in the order they were queued`() = runTest {
        val startedFileIds = mutableListOf<String>()
        var runningNow = 0
        var mostRunningAtOnce = 0
        val executor = object : DownloadExecutor {
            override suspend fun execute(request: DownloadRequest, onProgress: (DownloadProgress) -> Unit): DownloadOutcome {
                startedFileIds += request.fileId
                runningNow += 1
                mostRunningAtOnce = maxOf(mostRunningAtOnce, runningNow)
                delay(50)
                runningNow -= 1
                return DownloadOutcome.Completed(uri, "application/pdf")
            }
        }
        val queue = InMemoryDownloadQueue(backgroundScope, executor, notifier)

        queue.enqueue(listOf(request("1"), request("2"), request("3")))
        advanceTimeBy(1_000)
        runCurrent()

        assertEquals(listOf("1", "2", "3"), startedFileIds)
        assertEquals(1, mostRunningAtOnce)
        assertEquals(List(3) { DownloadState.Completed(uri, "application/pdf") }, queue.entries.value.map { it.state })
    }

    @Test
    fun `an entry reports growing progress and then completes`() = runTest {
        val finish = CompletableDeferred<Unit>()
        val executor = object : DownloadExecutor {
            override suspend fun execute(request: DownloadRequest, onProgress: (DownloadProgress) -> Unit): DownloadOutcome {
                onProgress(DownloadProgress(250, 1000))
                finish.await()
                return DownloadOutcome.Completed(uri, "application/pdf")
            }
        }
        val queue = InMemoryDownloadQueue(backgroundScope, executor, notifier)

        queue.enqueue(listOf(request("1"), request("2")))
        runCurrent()

        assertEquals(DownloadState.Running(DownloadProgress(250, 1000)), queue.entries.value[0].state)
        assertEquals(DownloadState.Queued, queue.entries.value[1].state)
        verify { notifier.showProgress(any(), request("1"), DownloadProgress(250, 1000)) }

        finish.complete(Unit)
        runCurrent()

        assertEquals(DownloadState.Completed(uri, "application/pdf"), queue.entries.value[0].state)
        verify { notifier.showCompleted(any(), request("1"), uri, "application/pdf") }
    }

    @Test
    fun `a failed download is recorded and the next one still runs`() = runTest {
        val executor = object : DownloadExecutor {
            override suspend fun execute(request: DownloadRequest, onProgress: (DownloadProgress) -> Unit) =
                if (request.fileId == "1") DownloadOutcome.Failed(DownloadFailure.Server("Arquivo não encontrado", 404))
                else DownloadOutcome.Completed(uri, "application/pdf")
        }
        val queue = InMemoryDownloadQueue(backgroundScope, executor, notifier)

        queue.enqueue(listOf(request("1"), request("2")))
        runCurrent()

        assertEquals(
            listOf(
                DownloadState.Failed(DownloadFailure.Server("Arquivo não encontrado", 404)),
                DownloadState.Completed(uri, "application/pdf"),
            ),
            queue.entries.value.map { it.state },
        )
        verify { notifier.showFailed(any(), request("1"), DownloadFailure.Server("Arquivo não encontrado", 404)) }
    }

    @Test
    fun `an unexpected error in the executor does not stop the queue`() = runTest {
        val executor = object : DownloadExecutor {
            override suspend fun execute(request: DownloadRequest, onProgress: (DownloadProgress) -> Unit): DownloadOutcome {
                if (request.fileId == "1") error("boom")
                return DownloadOutcome.Completed(uri, "application/pdf")
            }
        }
        val queue = InMemoryDownloadQueue(backgroundScope, executor, notifier)

        queue.enqueue(listOf(request("1"), request("2")))
        runCurrent()

        assertEquals(DownloadState.Failed(DownloadFailure.StorageUnavailable), queue.entries.value[0].state)
        assertEquals(DownloadState.Completed(uri, "application/pdf"), queue.entries.value[1].state)
    }
}
