package com.kuranas.android.core.paging

import com.kuranas.android.core.network.AppResult
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class PagedListLoaderTest {

    private val requestedPages = mutableListOf<Int>()
    private val failingPages = mutableSetOf<Int>()
    private val loader = PagedListLoader<Int>(CoroutineScope(UnconfinedTestDispatcher()), identityOf = { it })

    private fun fetchPage(page: Int): AppResult<ListPage<Int>> {
        requestedPages += page
        if (page in failingPages) return AppResult.Error("page $page failed")
        return AppResult.Success(ListPage(listOf(page * 10, page * 10 + 1), hasNext = page < 3))
    }

    private fun loadFirstPageAndFailSecond() {
        failingPages += 2
        loader.load { fetchPage(it) }
        loader.loadMore()
    }

    @Test
    fun `failed next page keeps items and exposes the error`() {
        loadFirstPageAndFailSecond()

        val state = loader.state.value
        assertEquals(listOf(10, 11), state.items)
        assertEquals("page 2 failed", state.loadMoreError)
        assertFalse(state.isLoadingMore)
        assertTrue(state.hasMore)
        assertEquals(2, state.nextPage)
        assertNull(state.error)
    }

    @Test
    fun `automatic loadMore does not repeat while the error is present`() {
        loadFirstPageAndFailSecond()
        failingPages.clear()

        loader.loadMore()

        assertEquals(listOf(1, 2), requestedPages)
        assertEquals("page 2 failed", loader.state.value.loadMoreError)
    }

    @Test
    fun `retryLoadMore requests the same page and clears the error on success`() {
        loadFirstPageAndFailSecond()
        failingPages.clear()

        loader.retryLoadMore()

        val state = loader.state.value
        assertEquals(listOf(1, 2, 2), requestedPages)
        assertEquals(listOf(10, 11, 20, 21), state.items)
        assertNull(state.loadMoreError)
        assertEquals(3, state.nextPage)
    }

    @Test
    fun `retryLoadMore failing again keeps the error`() {
        loadFirstPageAndFailSecond()

        loader.retryLoadMore()

        assertEquals(listOf(1, 2, 2), requestedPages)
        assertEquals("page 2 failed", loader.state.value.loadMoreError)
        assertEquals(listOf(10, 11), loader.state.value.items)
    }

    @Test
    fun `refresh clears the next page error`() {
        loadFirstPageAndFailSecond()
        failingPages.clear()

        loader.refresh()

        val state = loader.state.value
        assertNull(state.loadMoreError)
        assertEquals(listOf(10, 11), state.items)
        assertEquals(2, state.nextPage)
    }
}
