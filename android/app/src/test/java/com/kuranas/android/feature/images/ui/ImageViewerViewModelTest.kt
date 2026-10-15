package com.kuranas.android.feature.images.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.StarredResponse
import com.kuranas.android.feature.images.data.ImageLibraryItemDto
import com.kuranas.android.feature.images.data.ImageStarStore
import com.kuranas.android.feature.images.data.ImagesRepository
import io.mockk.coEvery
import io.mockk.mockk
import kotlinx.coroutines.CompletableDeferred
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
class ImageViewerViewModelTest {

    private val imagesRepository = mockk<ImagesRepository>()
    private val filesRepository = mockk<FilesRepository>()
    private val starStore = ImageStarStore()

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
        starStore.record(listOf(ImageLibraryItemDto(fileId = 7, name = "a.jpg", starred = false)))
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun viewModel() = ImageViewerViewModel(imagesRepository, filesRepository, starStore)

    @Test
    fun `toggle changes the star immediately and reconciles with the server`() {
        val pendingResponse = CompletableDeferred<AppResult<StarredResponse>>()
        coEvery { filesRepository.setFileStarred("7", true) } coAnswers { pendingResponse.await() }
        val viewModel = viewModel()

        viewModel.toggleStarred("7")

        assertEquals(true, viewModel.starredById.value["7"])
        pendingResponse.complete(AppResult.Success(StarredResponse(7, true)))
        assertEquals(true, viewModel.starredById.value["7"])
        assertNull(viewModel.actionErrorMessage.value)
    }

    @Test
    fun `server response overrides the optimistic value`() {
        coEvery { filesRepository.setFileStarred("7", true) } returns AppResult.Success(StarredResponse(7, false))
        val viewModel = viewModel()

        viewModel.toggleStarred("7")

        assertEquals(false, viewModel.starredById.value["7"])
    }

    @Test
    fun `error rolls back and exposes the server message`() {
        coEvery { filesRepository.setFileStarred("7", true) } returns AppResult.Error("forbidden by server")
        val viewModel = viewModel()

        viewModel.toggleStarred("7")

        assertEquals(false, viewModel.starredById.value["7"])
        assertEquals("forbidden by server", viewModel.actionErrorMessage.value)
        viewModel.dismissActionError()
        assertNull(viewModel.actionErrorMessage.value)
    }

    @Test
    fun `toggling an already starred image unstars it`() {
        starStore.set("7", true)
        coEvery { filesRepository.setFileStarred("7", false) } returns AppResult.Success(StarredResponse(7, false))
        val viewModel = viewModel()

        viewModel.toggleStarred("7")

        assertFalse(viewModel.starredById.value["7"] ?: true)
        assertTrue(viewModel.actionErrorMessage.value == null)
    }
}
