package com.kuranas.android.feature.ingest.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.feature.ingest.data.IngestDownloadDto
import com.kuranas.android.feature.ingest.data.IngestFetchResponseDto
import com.kuranas.android.feature.ingest.data.IngestPresetDto
import com.kuranas.android.feature.ingest.data.IngestRepository
import com.kuranas.android.feature.ingest.data.IngestTargetDto
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class IngestViewModelTest {

    private val repository = mockk<IngestRepository>()
    private val videosTarget = IngestTargetDto(label = "Vídeos", path = "/data/videos")
    private val musicTarget = IngestTargetDto(label = "Música", path = "/data/music")
    private val audioPreset = IngestPresetDto(key = "audio_mp3", label = "Áudio")
    private val videoPreset = IngestPresetDto(key = "video_1080", label = "Vídeo 1080p")
    private val runningDownload = IngestDownloadDto(id = 1, jobId = 10, url = "https://a", state = "running")
    private val failedDownload = IngestDownloadDto(id = 2, jobId = 11, url = "https://b", state = "failed")

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
        coEvery { repository.getTargets() } returns AppResult.Success(listOf(videosTarget, musicTarget))
        coEvery { repository.getPresets() } returns AppResult.Success(listOf(audioPreset, videoPreset))
        coEvery { repository.getDownloads(1) } returns
            AppResult.Success(ListPage(listOf(runningDownload, failedDownload), hasNext = false))
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun loadedViewModel() = IngestViewModel(repository).also { it.load() }

    @Test
    fun `load fills the options with the first target and preset selected and lists the downloads`() {
        val viewModel = loadedViewModel()

        val form = viewModel.state.value.form
        assertFalse(form.isLoadingOptions)
        assertEquals("/data/videos", form.selectedTargetPath)
        assertEquals("audio_mp3", form.selectedPresetKey)
        assertEquals(listOf(1, 2), viewModel.state.value.downloads.map { it.id })
    }

    @Test
    fun `a failed options load shows the server message and blocks submit`() {
        coEvery { repository.getTargets() } returns AppResult.Error("destinos indisponíveis")
        val viewModel = loadedViewModel()
        viewModel.changeUrl("https://a")

        val form = viewModel.state.value.form
        assertEquals("destinos indisponíveis", form.optionsError)
        assertFalse(form.canSubmit)
    }

    @Test
    fun `submit needs a url`() {
        val viewModel = loadedViewModel()

        assertFalse(viewModel.state.value.form.canSubmit)
        viewModel.changeUrl("https://a")
        assertTrue(viewModel.state.value.form.canSubmit)
    }

    @Test
    fun `submit sends the chosen destination and format then clears the form and reloads`() {
        coEvery { repository.requestDownload(any(), any(), any(), any()) } returns
            AppResult.Success(IngestFetchResponseDto(jobId = 5, downloadId = 3))
        val viewModel = loadedViewModel()
        viewModel.changeUrl(" https://a ")
        viewModel.selectTarget("/data/music")
        viewModel.selectPreset("video_1080")
        viewModel.changeSubfolder(" Shows ")

        viewModel.submit()

        coVerify { repository.requestDownload("https://a", "video_1080", "/data/music", "Shows") }
        val form = viewModel.state.value.form
        assertEquals("", form.url)
        assertEquals("", form.subfolder)
        assertFalse(form.isSubmitting)
        coVerify(atLeast = 2) { repository.getDownloads(1) }
    }

    @Test
    fun `a rejected submit keeps the url and shows the server message`() {
        coEvery { repository.requestDownload(any(), any(), any(), any()) } returns AppResult.Error("endereço inválido")
        val viewModel = loadedViewModel()
        viewModel.changeUrl("x")

        viewModel.submit()

        val form = viewModel.state.value.form
        assertEquals("x", form.url)
        assertEquals("endereço inválido", form.submitError)
        assertFalse(form.isSubmitting)
    }

    @Test
    fun `editing the url clears the previous submit error`() {
        coEvery { repository.requestDownload(any(), any(), any(), any()) } returns AppResult.Error("endereço inválido")
        val viewModel = loadedViewModel()
        viewModel.changeUrl("x")
        viewModel.submit()

        viewModel.changeUrl("xy")

        assertNull(viewModel.state.value.form.submitError)
    }

    @Test
    fun `cancel sends the job id and reloads`() {
        coEvery { repository.cancelDownload(10) } returns AppResult.Success(Unit)
        val viewModel = loadedViewModel()

        viewModel.cancel(runningDownload)

        coVerify { repository.cancelDownload(10) }
        coVerify(atLeast = 2) { repository.getDownloads(1) }
        assertNull(viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `retry sends the download id and a failure shows the server message`() {
        coEvery { repository.retryDownload(2) } returns AppResult.Error("pedido não pode ser repetido")
        val viewModel = loadedViewModel()

        viewModel.retry(failedDownload)

        coVerify { repository.retryDownload(2) }
        assertEquals("pedido não pode ser repetido", viewModel.state.value.actionErrorMessage)
        viewModel.dismissActionError()
        assertNull(viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `a failed downloads load shows the error`() {
        coEvery { repository.getDownloads(1) } returns AppResult.Error("histórico indisponível")
        val viewModel = loadedViewModel()

        assertEquals("histórico indisponível", viewModel.state.value.error)
    }
}
