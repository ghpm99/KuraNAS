package com.kuranas.android.feature.files.ui

import android.net.Uri
import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.SelectedFile
import com.kuranas.android.feature.files.data.UploadDestination
import com.kuranas.android.feature.files.data.UploadDestinationStore
import com.kuranas.android.feature.files.data.UploadResponseDto
import com.kuranas.android.feature.files.data.UploadedFileDto
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder
import java.io.File

@OptIn(ExperimentalCoroutinesApi::class)
class UploadViewModelTest {

    @get:Rule
    val temporaryFolder = TemporaryFolder()

    private val repository = mockk<FilesRepository>()
    private val storeScope = CoroutineScope(Dispatchers.IO + Job())
    private lateinit var store: UploadDestinationStore

    private val sharedUris = listOf(mockk<Uri>(), mockk<Uri>(), mockk<Uri>())
    private val selectedFiles = sharedUris.mapIndexed { index, uri -> SelectedFile(uri, "photo$index.jpg", 1000L + index) }
    private val photosFolder = UploadDestination("42", "Fotos")

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
        val preferencesFile = File(temporaryFolder.newFolder(), "upload.preferences_pb")
        store = UploadDestinationStore(PreferenceDataStoreFactory.create(scope = storeScope) { preferencesFile })
        coEvery { repository.describeSelection(sharedUris) } returns selectedFiles
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun newViewModel() = UploadViewModel(repository, UploadBatchRunner(repository), store)

    private fun response(status: String, name: String, error: String = "") =
        AppResult.Success(UploadResponseDto(message = "ok", files = listOf(UploadedFileDto(name, "/data/$name", status, error))))

    private suspend fun UploadViewModel.awaitFinished(): UploadPhase.Finished =
        state.first { it.phase is UploadPhase.Finished }.phase as UploadPhase.Finished

    private suspend fun UploadViewModel.awaitChoosing(): UploadPhase.ChoosingDestination =
        state.first { it.phase is UploadPhase.ChoosingDestination }.phase as UploadPhase.ChoosingDestination

    @Test
    fun `sharing opens the destination choice with no preselection on the first use`() = runTest {
        val viewModel = newViewModel()

        viewModel.receiveShared(sharedUris)

        assertNull(viewModel.awaitChoosing().lastDestination)
    }

    @Test
    fun `sending all files reports N of N progress and a clean summary`() = runTest {
        val progressSeen = mutableListOf<UploadPhase.Sending>()
        val viewModel = newViewModel()
        coEvery { repository.uploadFile(any(), "42") } coAnswers {
            progressSeen += viewModel.state.value.phase as UploadPhase.Sending
            response("uploaded", firstArg<SelectedFile>().name)
        }
        viewModel.receiveShared(sharedUris)
        viewModel.awaitChoosing()

        viewModel.confirmDestination(photosFolder, shouldAlwaysSendHere = false)
        val finished = viewModel.awaitFinished()

        assertEquals(
            listOf(
                UploadPhase.Sending(1, 3, "photo0.jpg"),
                UploadPhase.Sending(2, 3, "photo1.jpg"),
                UploadPhase.Sending(3, 3, "photo2.jpg"),
            ),
            progressSeen,
        )
        assertEquals(UploadSummary(uploadedCount = 3), finished.summary)
    }

    @Test
    fun `one failed file does not stop the others and the summary names it with the server reason`() = runTest {
        val viewModel = newViewModel()
        coEvery { repository.uploadFile(selectedFiles[0], "42") } returns response("renamed", "photo0.jpg")
        coEvery { repository.uploadFile(selectedFiles[1], "42") } returns response("failed", "photo1.jpg", "Falha ao gravar o arquivo")
        coEvery { repository.uploadFile(selectedFiles[2], "42") } returns response("uploaded", "photo2.jpg")
        viewModel.receiveShared(sharedUris)
        viewModel.awaitChoosing()

        viewModel.confirmDestination(photosFolder, shouldAlwaysSendHere = false)
        val summary = viewModel.awaitFinished().summary

        assertEquals(1, summary.uploadedCount)
        assertEquals(1, summary.renamedCount)
        assertEquals(
            listOf(UploadFailure(selectedFiles[1], UploadFailureReason.ServerMessage("Falha ao gravar o arquivo"))),
            summary.failures,
        )
        coVerify(exactly = 3) { repository.uploadFile(any(), "42") }
    }

    @Test
    fun `without a server the batch stops, every file is kept as failed and retry resends them`() = runTest {
        val viewModel = newViewModel()
        coEvery { repository.uploadFile(any(), "42") } returns AppResult.Error("Sem conexão com o servidor")
        viewModel.receiveShared(sharedUris)
        viewModel.awaitChoosing()

        viewModel.confirmDestination(photosFolder, shouldAlwaysSendHere = false)
        val failed = viewModel.awaitFinished()

        assertEquals(selectedFiles, failed.summary.failures.map { it.file })
        assertTrue(failed.summary.failures.all { it.reason == UploadFailureReason.NoServer })
        coVerify(exactly = 1) { repository.uploadFile(any(), any()) }

        coEvery { repository.uploadFile(any(), "42") } coAnswers { response("uploaded", firstArg<SelectedFile>().name) }
        viewModel.retryFailed()
        val retried = viewModel.state.first { (it.phase as? UploadPhase.Finished)?.summary?.hasFailures == false }.phase as UploadPhase.Finished

        assertEquals(3, retried.summary.uploadedCount)
        coVerify { repository.uploadFile(selectedFiles[2], "42") }
    }

    @Test
    fun `a forbidden response shows the server message as it came and keeps the files for retry`() = runTest {
        val viewModel = newViewModel()
        coEvery { repository.uploadFile(any(), "42") } returns AppResult.Error("IP não autorizado", code = 403)
        viewModel.receiveShared(sharedUris)
        viewModel.awaitChoosing()

        viewModel.confirmDestination(photosFolder, shouldAlwaysSendHere = false)
        val finished = viewModel.awaitFinished()

        assertTrue(finished.summary.failures.all { it.reason == UploadFailureReason.ServerMessage("IP não autorizado") })
        assertEquals(3, finished.summary.failures.size)
        coVerify(exactly = 1) { repository.uploadFile(any(), any()) }
    }

    @Test
    fun `always send here is stored and the next share goes straight to that folder`() = runTest {
        coEvery { repository.uploadFile(any(), "42") } coAnswers { response("uploaded", firstArg<SelectedFile>().name) }
        val firstShare = newViewModel()
        firstShare.receiveShared(sharedUris)
        firstShare.awaitChoosing()
        firstShare.confirmDestination(photosFolder, shouldAlwaysSendHere = true)
        firstShare.awaitFinished()

        assertEquals(photosFolder, store.alwaysSendDestination.first())

        val nextShare = newViewModel()
        nextShare.receiveShared(sharedUris)
        val finished = nextShare.awaitFinished()

        assertEquals(photosFolder, finished.destination)
        assertEquals(3, finished.summary.uploadedCount)
        coVerify(exactly = 6) { repository.uploadFile(any(), "42") }
    }

    @Test
    fun `clearing the direct destination makes the next share ask again, preselecting the last folder`() = runTest {
        coEvery { repository.uploadFile(any(), "42") } coAnswers { response("uploaded", firstArg<SelectedFile>().name) }
        val firstShare = newViewModel()
        firstShare.receiveShared(sharedUris)
        firstShare.awaitChoosing()
        firstShare.confirmDestination(photosFolder, shouldAlwaysSendHere = true)
        firstShare.awaitFinished()

        store.clearAlwaysSendDestination()
        val nextShare = newViewModel()
        nextShare.receiveShared(sharedUris)

        assertEquals(photosFolder, nextShare.awaitChoosing().lastDestination)
    }

    @Test
    fun `sending to the root uploads with a null folder id`() = runTest {
        coEvery { repository.uploadFile(any(), null) } coAnswers { response("uploaded", firstArg<SelectedFile>().name) }
        val viewModel = newViewModel()

        viewModel.sendPicked(sharedUris, UploadDestination.Root)
        val finished = viewModel.awaitFinished()

        assertEquals(3, finished.summary.uploadedCount)
        assertEquals(UploadDestination.Root, store.lastDestination.first())
    }

    @Test
    fun `dismissing the result returns to idle`() = runTest {
        coEvery { repository.uploadFile(any(), null) } coAnswers { response("uploaded", firstArg<SelectedFile>().name) }
        val viewModel = newViewModel()
        viewModel.sendPicked(sharedUris, UploadDestination.Root)
        viewModel.awaitFinished()

        viewModel.dismissResult()

        assertEquals(UploadPhase.Idle, viewModel.state.value.phase)
    }
}
