package com.kuranas.android.feature.files.ui

import android.net.Uri
import app.cash.turbine.test
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FileContent
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.download.DownloadEntry
import com.kuranas.android.feature.files.data.download.DownloadExecutor
import com.kuranas.android.feature.files.data.download.DownloadFailure
import com.kuranas.android.feature.files.data.download.DownloadOutcome
import com.kuranas.android.feature.files.data.download.DownloadProgress
import com.kuranas.android.feature.files.data.download.DownloadPurpose
import com.kuranas.android.feature.files.data.download.DownloadQueue
import com.kuranas.android.feature.files.data.download.DownloadRequest
import com.kuranas.android.feature.files.data.download.DownloadState
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

private class RecordingDownloadQueue : DownloadQueue {
    private val mutableEntries = MutableStateFlow<List<DownloadEntry>>(emptyList())
    override val entries: StateFlow<List<DownloadEntry>> = mutableEntries
    val enqueuedRequests = mutableListOf<DownloadRequest>()

    override fun enqueue(requests: List<DownloadRequest>): List<Long> {
        enqueuedRequests += requests
        val created = requests.mapIndexed { index, request ->
            DownloadEntry((enqueuedRequests.size - requests.size + index + 1).toLong(), request, DownloadState.Queued)
        }
        mutableEntries.update { it + created }
        return created.map { it.id }
    }

    fun setState(entryId: Long, state: DownloadState) {
        mutableEntries.update { current -> current.map { if (it.id == entryId) it.copy(state = state) else it } }
    }
}

@OptIn(ExperimentalCoroutinesApi::class)
class FileViewerViewModelTest {

    private val repository = mockk<FilesRepository>()
    private val queue = RecordingDownloadQueue()
    private val executor = mockk<DownloadExecutor>()
    private val uri = mockk<Uri>()
    private val smallPdf = FileItemDto(rawId = 5, name = "Relatório São João.pdf", format = ".pdf", size = 2048)

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun viewModel() = FileViewerViewModel(repository, queue, executor)

    private fun givenInfo(file: FileItemDto) {
        coEvery { repository.getFileInfo(file.id) } returns AppResult.Success(file)
    }

    @Test
    fun `a file within the limit loads its content`() {
        givenInfo(smallPdf)
        val content = FileContent.Text("texto", truncated = false)
        coEvery { repository.getFileContent("5") } returns AppResult.Success(content)
        val viewModel = viewModel()

        viewModel.load("5")

        assertEquals(FileViewerState.Content(content, smallPdf), viewModel.state.value)
    }

    @Test
    fun `a file above 20 MB offers download and never fetches the content`() {
        val hugeVideo = smallPdf.copy(size = 20L * 1024 * 1024 + 1)
        givenInfo(hugeVideo)
        val viewModel = viewModel()

        viewModel.load("5")

        assertEquals(FileViewerState.TooLargeToPreview(hugeVideo), viewModel.state.value)
        coVerify(exactly = 0) { repository.getFileContent(any()) }
    }

    @Test
    fun `a metadata failure shows the server message`() {
        coEvery { repository.getFileInfo("5") } returns AppResult.Error("Arquivo não encontrado", 404)
        val viewModel = viewModel()

        viewModel.load("5")

        assertEquals(FileViewerState.Failed("Arquivo não encontrado"), viewModel.state.value)
    }

    @Test
    fun `download queues the listed name, type and size and ignores a second tap while running`() {
        givenInfo(smallPdf)
        coEvery { repository.getFileContent("5") } returns AppResult.Success(FileContent.Text("", false))
        val viewModel = viewModel()
        viewModel.load("5")

        viewModel.download()
        viewModel.download()

        assertEquals(
            listOf(DownloadRequest("5", "Relatório São João.pdf", 2048, DownloadPurpose.SAVE_TO_DEVICE)),
            queue.enqueuedRequests,
        )
    }

    @Test
    fun `download state follows the queue through progress, failure and completion`() {
        givenInfo(smallPdf)
        coEvery { repository.getFileContent("5") } returns AppResult.Success(FileContent.Text("", false))
        val viewModel = viewModel()
        viewModel.load("5")
        assertNull(viewModel.downloadState.value)

        viewModel.download()
        assertEquals(DownloadState.Queued, viewModel.downloadState.value)

        queue.setState(1, DownloadState.Running(DownloadProgress(512, 2048)))
        val firstProgress = (viewModel.downloadState.value as DownloadState.Running).progress.receivedBytes
        queue.setState(1, DownloadState.Running(DownloadProgress(1536, 2048)))
        val secondProgress = (viewModel.downloadState.value as DownloadState.Running).progress.receivedBytes
        assertTrue(secondProgress > firstProgress)

        queue.setState(1, DownloadState.Failed(DownloadFailure.ConnectionLost))
        assertEquals(DownloadState.Failed(DownloadFailure.ConnectionLost), viewModel.downloadState.value)

        viewModel.download()
        queue.setState(2, DownloadState.Completed(uri, "application/pdf"))
        assertEquals(DownloadState.Completed(uri, "application/pdf"), viewModel.downloadState.value)
    }

    @Test
    fun `share downloads to the cache and delivers the uri`() = runTest {
        givenInfo(smallPdf)
        coEvery { repository.getFileContent("5") } returns AppResult.Success(FileContent.Text("", false))
        coEvery { executor.execute(any(), any()) } returns DownloadOutcome.Completed(uri, "application/pdf")
        val viewModel = viewModel()
        viewModel.load("5")

        viewModel.deliveries.test {
            viewModel.share()

            assertEquals(
                ExternalDelivery.Ready(uri, "application/pdf", "Relatório São João.pdf", DownloadPurpose.SHARE),
                awaitItem(),
            )
        }
        assertNull(viewModel.externalPreparation.value)
    }

    @Test
    fun `open with reports a failed preparation`() = runTest {
        givenInfo(smallPdf)
        coEvery { repository.getFileContent("5") } returns AppResult.Success(FileContent.Text("", false))
        coEvery { executor.execute(any(), any()) } returns DownloadOutcome.Failed(DownloadFailure.ConnectionLost)
        val viewModel = viewModel()
        viewModel.load("5")

        viewModel.deliveries.test {
            viewModel.openWith()

            assertEquals(ExternalDelivery.Failed(DownloadFailure.ConnectionLost), awaitItem())
        }
    }
}
