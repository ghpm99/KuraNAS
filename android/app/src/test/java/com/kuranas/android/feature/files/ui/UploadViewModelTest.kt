package com.kuranas.android.feature.files.ui

import android.net.Uri
import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import com.kuranas.android.feature.files.data.UploadDestination
import com.kuranas.android.feature.files.data.UploadDestinationStore
import com.kuranas.android.feature.transfers.data.upload.TransferRepository
import com.kuranas.android.feature.transfers.data.upload.UploadRequestItem
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
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
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder
import java.io.File
import java.io.IOException

@OptIn(ExperimentalCoroutinesApi::class)
class UploadViewModelTest {

    @get:Rule
    val temporaryFolder = TemporaryFolder()

    private val transferRepository = mockk<TransferRepository>()
    private val storeScope = CoroutineScope(Dispatchers.IO + Job())
    private lateinit var store: UploadDestinationStore

    private val uris = listOf("photo0.jpg", "photo1.jpg", "photo2.jpg").map { name ->
        mockk<Uri> { every { lastPathSegment } returns name }
    }
    private val pickedItems = uris.mapIndexed { index, uri ->
        UploadRequestItem("photo$index.jpg", 1000L + index, sourceUri = "content://picked/$index")
    }
    private val photosFolder = UploadDestination("42", "Fotos")

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
        val preferencesFile = File(temporaryFolder.newFolder(), "upload.preferences_pb")
        store = UploadDestinationStore(PreferenceDataStoreFactory.create(scope = storeScope) { preferencesFile })
        uris.forEachIndexed { index, uri -> coEvery { transferRepository.persistPickedUri(uri) } returns pickedItems[index] }
        uris.forEachIndexed { index, uri -> coEvery { transferRepository.stageSharedUri(uri) } returns stagedFile(index).absolutePath }
        coEvery { transferRepository.enqueueUploads(any(), any()) } returns "batch"
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private val stagedFiles = mutableMapOf<Int, File>()

    private fun stagedFile(index: Int): File = stagedFiles.getOrPut(index) {
        File(temporaryFolder.newFolder(), "photo$index.jpg").apply { writeBytes(ByteArray(10 + index)) }
    }

    private fun stagedItem(index: Int) =
        UploadRequestItem("photo$index.jpg", 10L + index, stagedFilePath = stagedFile(index).absolutePath)

    private fun newViewModel() = UploadViewModel(transferRepository, store)

    private suspend fun UploadViewModel.awaitQueued(): UploadPhase.Queued =
        state.first { it.phase is UploadPhase.Queued }.phase as UploadPhase.Queued

    private suspend fun UploadViewModel.awaitChoosing(): UploadPhase.ChoosingDestination =
        state.first { it.phase is UploadPhase.ChoosingDestination }.phase as UploadPhase.ChoosingDestination

    @Test
    fun `sharing opens the destination choice with no preselection on the first use`() = runTest {
        val viewModel = newViewModel()

        viewModel.receiveShared(uris)

        assertNull(viewModel.awaitChoosing().lastDestination)
    }

    @Test
    fun `explorer picks are persisted and enqueued with the chosen destination`() = runTest {
        val viewModel = newViewModel()

        viewModel.sendPicked(uris, photosFolder)
        val queued = viewModel.awaitQueued()

        assertEquals(UploadPhase.Queued(3, emptyList()), queued)
        coVerify { transferRepository.enqueueUploads(pickedItems, photosFolder) }
        coVerify(exactly = 0) { transferRepository.stageSharedUri(any()) }
        assertEquals(photosFolder, store.lastDestination.first())
    }

    @Test
    fun `shared files are staged and enqueued after the destination is confirmed`() = runTest {
        val viewModel = newViewModel()
        viewModel.receiveShared(uris)
        viewModel.awaitChoosing()

        viewModel.confirmDestination(photosFolder, shouldAlwaysSendHere = false)
        val queued = viewModel.awaitQueued()

        assertEquals(UploadPhase.Queued(3, emptyList()), queued)
        coVerify { transferRepository.enqueueUploads(listOf(stagedItem(0), stagedItem(1), stagedItem(2)), photosFolder) }
        coVerify(exactly = 0) { transferRepository.persistPickedUri(any()) }
    }

    @Test
    fun `one unreadable shared file does not block the others and is reported by name`() = runTest {
        coEvery { transferRepository.stageSharedUri(uris[1]) } throws IOException("gone")
        val viewModel = newViewModel()
        viewModel.receiveShared(uris)
        viewModel.awaitChoosing()

        viewModel.confirmDestination(photosFolder, shouldAlwaysSendHere = false)
        val queued = viewModel.awaitQueued()

        assertEquals(UploadPhase.Queued(2, listOf("photo1.jpg")), queued)
        coVerify { transferRepository.enqueueUploads(listOf(stagedItem(0), stagedItem(2)), photosFolder) }
    }

    @Test
    fun `an explorer pick refused by the provider is reported and the others are enqueued`() = runTest {
        coEvery { transferRepository.persistPickedUri(uris[0]) } throws SecurityException("denied")
        val viewModel = newViewModel()

        viewModel.sendPicked(uris, photosFolder)
        val queued = viewModel.awaitQueued()

        assertEquals(UploadPhase.Queued(2, listOf("photo0.jpg")), queued)
        coVerify { transferRepository.enqueueUploads(listOf(pickedItems[1], pickedItems[2]), photosFolder) }
    }

    @Test
    fun `when nothing is readable nothing is enqueued and every name is reported`() = runTest {
        uris.forEach { coEvery { transferRepository.persistPickedUri(it) } throws IOException("gone") }
        val viewModel = newViewModel()

        viewModel.sendPicked(uris, photosFolder)
        val queued = viewModel.awaitQueued()

        assertEquals(UploadPhase.Queued(0, listOf("photo0.jpg", "photo1.jpg", "photo2.jpg")), queued)
        coVerify(exactly = 0) { transferRepository.enqueueUploads(any(), any()) }
    }

    @Test
    fun `always send here is stored and the next share goes straight to that folder`() = runTest {
        val firstShare = newViewModel()
        firstShare.receiveShared(uris)
        firstShare.awaitChoosing()
        firstShare.confirmDestination(photosFolder, shouldAlwaysSendHere = true)
        firstShare.awaitQueued()

        assertEquals(photosFolder, store.alwaysSendDestination.first())

        val nextShare = newViewModel()
        nextShare.receiveShared(uris)
        val queued = nextShare.awaitQueued()

        assertEquals(3, queued.queuedCount)
        coVerify(exactly = 2) { transferRepository.enqueueUploads(any(), photosFolder) }
    }

    @Test
    fun `clearing the direct destination makes the next share ask again, preselecting the last folder`() = runTest {
        val firstShare = newViewModel()
        firstShare.receiveShared(uris)
        firstShare.awaitChoosing()
        firstShare.confirmDestination(photosFolder, shouldAlwaysSendHere = true)
        firstShare.awaitQueued()

        store.clearAlwaysSendDestination()
        val nextShare = newViewModel()
        nextShare.receiveShared(uris)

        assertEquals(photosFolder, nextShare.awaitChoosing().lastDestination)
    }

    @Test
    fun `sending to the root enqueues with the root destination`() = runTest {
        val viewModel = newViewModel()

        viewModel.sendPicked(uris, UploadDestination.Root)
        viewModel.awaitQueued()

        coVerify { transferRepository.enqueueUploads(pickedItems, UploadDestination.Root) }
        assertEquals(UploadDestination.Root, store.lastDestination.first())
    }

    @Test
    fun `dismissing the confirmation returns to idle`() = runTest {
        val viewModel = newViewModel()
        viewModel.sendPicked(uris, UploadDestination.Root)
        viewModel.awaitQueued()

        viewModel.dismissResult()

        assertEquals(UploadPhase.Idle, viewModel.state.value.phase)
    }
}
