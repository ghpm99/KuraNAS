package com.kuranas.android.feature.documents.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.feature.documents.data.DocumentDto
import com.kuranas.android.feature.documents.data.DocumentSearchResultDto
import com.kuranas.android.feature.documents.data.DocumentsRepository
import com.kuranas.android.feature.files.data.download.DownloadExecutor
import com.kuranas.android.feature.files.data.download.DownloadFailure
import com.kuranas.android.feature.files.data.download.DownloadOutcome
import com.kuranas.android.feature.files.data.download.DownloadPurpose
import com.kuranas.android.feature.files.data.download.DownloadQueue
import com.kuranas.android.feature.files.data.download.DownloadRequest
import com.kuranas.android.feature.files.ui.ExternalDelivery
import kotlinx.coroutines.flow.toList
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.runTest
import android.net.Uri
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import io.mockk.slot
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.TestCoroutineScheduler
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
class DocumentsViewModelTest {

    private val scheduler = TestCoroutineScheduler()
    private val repository = mockk<DocumentsRepository>()
    private val downloadQueue = mockk<DownloadQueue>()
    private val downloadExecutor = mockk<DownloadExecutor>()
    private val enqueuedRequests = slot<List<DownloadRequest>>()

    private fun newViewModel() = DocumentsViewModel(repository, downloadQueue, downloadExecutor)

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher(scheduler))
        coEvery { downloadQueue.enqueue(capture(enqueuedRequests)) } returns listOf(1L)
        coEvery { repository.getDocumentsCount(any(), any()) } returns AppResult.Success(3)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    private fun document(id: Int) = DocumentDto(id = id, name = "doc$id.pdf", format = ".pdf")

    private fun DocumentsViewModel.listedIds() = state.value.documents.map { it.id }

    @Test
    fun `lists the first page and the total from the count`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1), document(2)), hasNext = true))

        val viewModel = newViewModel().also { it.startIfNeeded() }

        assertFalse(viewModel.state.value.isLoading)
        assertEquals(listOf(1, 2), viewModel.listedIds())
        assertTrue(viewModel.state.value.hasMore)
        assertEquals(3, viewModel.state.value.totalDocuments)
    }

    @Test
    fun `startIfNeeded does not reload after a rotation`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = false))
        val viewModel = newViewModel()

        viewModel.startIfNeeded()
        viewModel.startIfNeeded()

        coVerify(exactly = 1) { repository.getDocuments(1, any(), any()) }
        assertEquals(listOf(1), viewModel.listedIds())
    }

    @Test
    fun `loadMore appends the next page`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = true))
        coEvery { repository.getDocuments(2, any(), any()) } returns AppResult.Success(ListPage(listOf(document(2)), hasNext = false))
        val viewModel = newViewModel().also { it.startIfNeeded() }

        viewModel.loadMore()

        assertEquals(listOf(1, 2), viewModel.listedIds())
        assertFalse(viewModel.state.value.hasMore)
    }

    @Test
    fun `a next page error keeps the items and retry loads that page`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = true))
        coEvery { repository.getDocuments(2, any(), any()) } returns AppResult.Error("falha na página")
        val viewModel = newViewModel().also { it.startIfNeeded() }

        viewModel.loadMore()

        assertEquals(listOf(1), viewModel.listedIds())
        assertEquals("falha na página", viewModel.state.value.loadMoreError)
        assertNull(viewModel.state.value.error)

        coEvery { repository.getDocuments(2, any(), any()) } returns AppResult.Success(ListPage(listOf(document(2)), hasNext = false))
        viewModel.retryLoadMore()

        assertEquals(listOf(1, 2), viewModel.listedIds())
        assertNull(viewModel.state.value.loadMoreError)
    }

    @Test
    fun `refresh restarts from the first page`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = false))
        val viewModel = newViewModel().also { it.startIfNeeded() }
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(5), document(6)), hasNext = false))
        coEvery { repository.getDocumentsCount(any(), any()) } returns AppResult.Success(2)

        viewModel.refresh()

        assertEquals(listOf(5, 6), viewModel.listedIds())
        assertEquals(2, viewModel.state.value.totalDocuments)
        assertFalse(viewModel.state.value.isRefreshing)
    }

    @Test
    fun `a refused first load shows the server message verbatim`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Error("acesso negado a este IP", 403)

        val viewModel = newViewModel().also { it.startIfNeeded() }

        assertEquals("acesso negado a este IP", viewModel.state.value.error)
        assertTrue(viewModel.state.value.documents.isEmpty())
    }

    @Test
    fun `a failed count hides the total without breaking the list`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = false))
        coEvery { repository.getDocumentsCount(any(), any()) } returns AppResult.Error("sem total")

        val viewModel = newViewModel().also { it.startIfNeeded() }

        assertNull(viewModel.state.value.totalDocuments)
        assertEquals(listOf(1), viewModel.listedIds())
    }

    private fun searchResult(fileId: Int, snippet: String = "trecho", page: Int = 0) =
        DocumentSearchResultDto(fileId = fileId, name = "achado$fileId.pdf", format = ".pdf", snippet = snippet, page = page)

    private fun listItem(id: Int, format: String) = DocumentListItem(
        id = id, name = "arquivo$id$format", format = format, parentPath = "", size = 10, timestamp = "", isStarred = false,
    )

    private fun emptyFirstPage() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(emptyList(), hasNext = false))
    }

    @Test
    fun `a format chip reloads the list and the count with that format`() {
        emptyFirstPage()
        val viewModel = newViewModel().also { it.startIfNeeded() }

        viewModel.selectFormat(DocumentFormatFilter.SPREADSHEET)

        coVerify { repository.getDocuments(1, "spreadsheet", any()) }
        coVerify { repository.getDocumentsCount("spreadsheet", any()) }
        assertEquals(DocumentFormatFilter.SPREADSHEET, viewModel.state.value.formatFilter)
    }

    @Test
    fun `typing waits the debounce and ignores a single character`() {
        emptyFirstPage()
        coEvery { repository.searchDocuments("contrato", 1, any()) } returns AppResult.Success(ListPage(listOf(searchResult(7)), hasNext = false))
        val viewModel = newViewModel().also { it.startIfNeeded() }

        viewModel.changeSearchText("c")
        scheduler.advanceTimeBy(1_000)
        coVerify(exactly = 0) { repository.searchDocuments(any(), any(), any()) }

        viewModel.changeSearchText("contrato")
        scheduler.advanceTimeBy(SEARCH_DEBOUNCE_MILLIS - 1)
        coVerify(exactly = 0) { repository.searchDocuments(any(), any(), any()) }

        scheduler.advanceTimeBy(1)
        scheduler.runCurrent()
        coVerify(exactly = 1) { repository.searchDocuments("contrato", 1, any()) }
        assertEquals(listOf(7), viewModel.listedIds())
        assertTrue(viewModel.state.value.isSearching)
        assertNull(viewModel.state.value.totalDocuments)
    }

    @Test
    fun `search results carry the snippet and the page and paginate with the format`() {
        emptyFirstPage()
        coEvery { repository.searchDocuments("termo", 1, "pdf") } returns
            AppResult.Success(ListPage(listOf(searchResult(1, "linha um", page = 4)), hasNext = true))
        coEvery { repository.searchDocuments("termo", 2, "pdf") } returns
            AppResult.Success(ListPage(listOf(searchResult(2, "linha dois")), hasNext = false))
        val viewModel = newViewModel().also { it.startIfNeeded() }
        viewModel.selectFormat(DocumentFormatFilter.PDF)

        viewModel.changeSearchText("termo")
        scheduler.advanceTimeBy(SEARCH_DEBOUNCE_MILLIS)
        scheduler.runCurrent()
        viewModel.loadMore()

        val documents = viewModel.state.value.documents
        assertEquals(listOf("linha um", "linha dois"), documents.map { it.snippet })
        assertEquals(listOf<Int?>(4, null), documents.map { it.pageNumber })
    }

    @Test
    fun `a search error shows the server message and keeps the next page error separate`() {
        emptyFirstPage()
        coEvery { repository.searchDocuments("termo", 1, any()) } returns AppResult.Error("busca indisponível", 503)
        val viewModel = newViewModel().also { it.startIfNeeded() }

        viewModel.changeSearchText("termo")
        scheduler.advanceTimeBy(SEARCH_DEBOUNCE_MILLIS)
        scheduler.runCurrent()

        assertEquals("busca indisponível", viewModel.state.value.error)
    }

    @Test
    fun `a search next page error keeps the results`() {
        emptyFirstPage()
        coEvery { repository.searchDocuments("termo", 1, any()) } returns AppResult.Success(ListPage(listOf(searchResult(1)), hasNext = true))
        coEvery { repository.searchDocuments("termo", 2, any()) } returns AppResult.Error("falha na página")
        val viewModel = newViewModel().also { it.startIfNeeded() }
        viewModel.changeSearchText("termo")
        scheduler.advanceTimeBy(SEARCH_DEBOUNCE_MILLIS)
        scheduler.runCurrent()

        viewModel.loadMore()

        assertEquals(listOf(1), viewModel.listedIds())
        assertEquals("falha na página", viewModel.state.value.loadMoreError)
        assertNull(viewModel.state.value.error)
    }

    @Test
    fun `clearing the search returns to the plain list immediately`() {
        coEvery { repository.getDocuments(1, any(), any()) } returns AppResult.Success(ListPage(listOf(document(1)), hasNext = false))
        coEvery { repository.searchDocuments("termo", 1, any()) } returns AppResult.Success(ListPage(listOf(searchResult(9)), hasNext = false))
        val viewModel = newViewModel().also { it.startIfNeeded() }
        viewModel.changeSearchText("termo")
        scheduler.advanceTimeBy(SEARCH_DEBOUNCE_MILLIS)
        scheduler.runCurrent()

        viewModel.clearSearch()

        assertEquals(listOf(1), viewModel.listedIds())
        assertFalse(viewModel.state.value.isSearching)
        assertEquals("", viewModel.state.value.searchText)
        assertEquals(3, viewModel.state.value.totalDocuments)
    }

    @Test
    fun `open decides the viewer for text and open with for the other formats`() {
        assertEquals(DocumentOpenMode.IN_APP_VIEWER, documentOpenMode(".txt"))
        assertEquals(DocumentOpenMode.IN_APP_VIEWER, documentOpenMode(".MD"))
        assertEquals(DocumentOpenMode.OPEN_WITH, documentOpenMode(".pdf"))
        assertEquals(DocumentOpenMode.OPEN_WITH, documentOpenMode(".xlsx"))
        assertEquals(DocumentOpenMode.OPEN_WITH, documentOpenMode(".pptx"))
    }

    @Test
    fun `opening a text document asks the screen for the in app viewer`() = runTest(scheduler) {
        emptyFirstPage()
        val viewModel = newViewModel().also { it.startIfNeeded() }
        val events = mutableListOf<DocumentsEvent>()
        val collecting = launch(UnconfinedTestDispatcher(scheduler)) { viewModel.events.toList(events) }

        viewModel.open(listItem(4, ".txt"))

        assertEquals(listOf<DocumentsEvent>(DocumentsEvent.OpenInViewer(4, "arquivo4.txt")), events)
        collecting.cancel()
    }

    @Test
    fun `opening a pdf downloads it for another app with the open with purpose`() = runTest(scheduler) {
        emptyFirstPage()
        val cachedUri = mockk<Uri>()
        val executedRequest = slot<DownloadRequest>()
        coEvery { downloadExecutor.execute(capture(executedRequest), any()) } returns DownloadOutcome.Completed(cachedUri, "application/pdf")
        val viewModel = newViewModel().also { it.startIfNeeded() }
        val events = mutableListOf<DocumentsEvent>()
        val collecting = launch(UnconfinedTestDispatcher(scheduler)) { viewModel.events.toList(events) }

        viewModel.open(listItem(5, ".pdf"))

        assertEquals(DownloadRequest("5", "arquivo5.pdf", 10, DownloadPurpose.OPEN_WITH), executedRequest.captured)
        val delivery = (events.single() as DocumentsEvent.OpenExternally).delivery
        assertEquals(ExternalDelivery.Ready(cachedUri, "application/pdf", "arquivo5.pdf", DownloadPurpose.OPEN_WITH), delivery)
        assertNull(viewModel.state.value.preparingDocumentId)
        collecting.cancel()
    }

    @Test
    fun `a failed open with preparation reports the failure`() = runTest(scheduler) {
        emptyFirstPage()
        coEvery { downloadExecutor.execute(any(), any()) } returns DownloadOutcome.Failed(DownloadFailure.Server("recusado", 403))
        val viewModel = newViewModel().also { it.startIfNeeded() }
        val events = mutableListOf<DocumentsEvent>()
        val collecting = launch(UnconfinedTestDispatcher(scheduler)) { viewModel.events.toList(events) }

        viewModel.open(listItem(5, ".xlsx"))

        val delivery = (events.single() as DocumentsEvent.OpenExternally).delivery
        assertEquals(ExternalDelivery.Failed(DownloadFailure.Server("recusado", 403)), delivery)
        collecting.cancel()
    }

    @Test
    fun `download enqueues a save to device request`() {
        emptyFirstPage()
        val viewModel = newViewModel().also { it.startIfNeeded() }

        viewModel.download(listItem(8, ".docx"))

        assertEquals(listOf(DownloadRequest("8", "arquivo8.docx", 10, DownloadPurpose.SAVE_TO_DEVICE)), enqueuedRequests.captured)
    }
}
