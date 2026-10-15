package com.kuranas.android.feature.files.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.feature.files.data.FileItemDto
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.StarredKind
import com.kuranas.android.feature.files.data.StarredSort
import com.kuranas.android.feature.files.data.StarredResponse
import io.mockk.coEvery
import io.mockk.coVerify
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
class FavoritesViewModelTest {

    private val repository = mockk<FilesRepository>()

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun givenPage(
        page: Int,
        hasNext: Boolean,
        vararg rawIds: Int,
        kind: StarredKind = StarredKind.ALL,
        sort: StarredSort = StarredSort.NAME,
    ) {
        coEvery { repository.getStarredFiles(page, kind, sort) } returns
            AppResult.Success(ListPage(rawIds.map { FileItemDto(rawId = it, starred = true) }, hasNext))
    }

    @Test
    fun `initial load fills the first page with default kind and sort`() {
        givenPage(1, true, 1, 2)

        val state = FavoritesViewModel(repository).state.value

        assertFalse(state.isLoading)
        assertNull(state.error)
        assertEquals(listOf("1", "2"), state.favorites.map { it.id })
        assertTrue(state.hasMore)
    }

    @Test
    fun `loadMore appends the next page without duplicates`() {
        givenPage(1, true, 1, 2)
        givenPage(2, false, 2, 3)
        val viewModel = FavoritesViewModel(repository)

        viewModel.loadMore()

        assertEquals(listOf("1", "2", "3"), viewModel.state.value.favorites.map { it.id })
        assertFalse(viewModel.state.value.hasMore)
    }

    @Test
    fun `changing the kind chip reloads from page one with the kind`() {
        givenPage(1, true, 1)
        givenPage(2, false, 2)
        givenPage(1, false, 9, kind = StarredKind.FOLDERS)
        val viewModel = FavoritesViewModel(repository)
        viewModel.loadMore()

        viewModel.selectKind(StarredKind.FOLDERS)

        coVerify(exactly = 1) { repository.getStarredFiles(1, StarredKind.FOLDERS, StarredSort.NAME) }
        assertEquals(StarredKind.FOLDERS, viewModel.state.value.kind)
        assertEquals(listOf("9"), viewModel.state.value.favorites.map { it.id })
    }

    @Test
    fun `changing the sort reloads from page one with the sort and keeps the kind`() {
        givenPage(1, false, 1)
        givenPage(1, false, 4, kind = StarredKind.MEDIA)
        givenPage(1, false, 5, kind = StarredKind.MEDIA, sort = StarredSort.STARRED_AT)
        val viewModel = FavoritesViewModel(repository)
        viewModel.selectKind(StarredKind.MEDIA)

        viewModel.selectSort(StarredSort.STARRED_AT)

        coVerify(exactly = 1) { repository.getStarredFiles(1, StarredKind.MEDIA, StarredSort.STARRED_AT) }
        assertEquals(listOf("5"), viewModel.state.value.favorites.map { it.id })
    }

    @Test
    fun `unstar removes the item immediately and keeps it removed on success`() {
        givenPage(1, false, 1, 2)
        val response = CompletableDeferred<AppResult<StarredResponse>>()
        coEvery { repository.setFileStarred("1", false) } coAnswers { response.await() }
        val viewModel = FavoritesViewModel(repository)

        viewModel.unstar(viewModel.state.value.favorites.first())

        assertEquals(listOf("2"), viewModel.state.value.favorites.map { it.id })
        response.complete(AppResult.Success(StarredResponse(1, false)))
        assertEquals(listOf("2"), viewModel.state.value.favorites.map { it.id })
        assertNull(viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `unstar failure reinserts the item at its position with the server message`() {
        givenPage(1, false, 1, 2, 3)
        coEvery { repository.setFileStarred("2", false) } returns AppResult.Error("server said no")
        val viewModel = FavoritesViewModel(repository)

        viewModel.unstar(viewModel.state.value.favorites[1])

        assertEquals(listOf("1", "2", "3"), viewModel.state.value.favorites.map { it.id })
        assertEquals("server said no", viewModel.state.value.actionErrorMessage)
        viewModel.dismissActionError()
        assertNull(viewModel.state.value.actionErrorMessage)
    }

    @Test
    fun `server error on load is shown as received`() {
        coEvery { repository.getStarredFiles(1, StarredKind.ALL, StarredSort.NAME) } returns AppResult.Error("down")

        val state = FavoritesViewModel(repository).state.value

        assertEquals("down", state.error)
        assertFalse(state.isLoading)
    }
}
