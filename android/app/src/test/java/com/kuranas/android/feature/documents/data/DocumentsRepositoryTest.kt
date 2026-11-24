package com.kuranas.android.feature.documents.data

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import com.kuranas.android.core.network.AppResult
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import retrofit2.Retrofit
import java.io.File

class DocumentsRepositoryTest {

    private val server = MockWebServer()
    private lateinit var repository: DocumentsRepository

    @Before
    fun setUp() {
        server.start()
        val json = Json {
            ignoreUnknownKeys = true
            coerceInputValues = true
            explicitNulls = false
        }
        val api = Retrofit.Builder()
            .baseUrl(server.url("/"))
            .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
            .build()
            .create(DocumentsApi::class.java)
        repository = DocumentsRepository(api)
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    private fun backendGolden(fileName: String): String =
        File("../../backend/internal/api/v1/documenttext/testdata/$fileName").readText()

    @Test
    fun `getDocuments decodes the backend golden page that has a next page`() = runTest {
        server.enqueue(MockResponse().setBody(backendGolden("documents_page_with_next.json")))

        val result = repository.getDocuments(page = 1)

        assertEquals("/api/v1/documents?page=1&page_size=50", server.takeRequest().path)
        val page = (result as AppResult.Success).data
        assertTrue(page.hasNext)
        assertTrue(page.items.isNotEmpty())
        assertTrue(page.items.all { it.id > 0 && it.name.isNotBlank() && it.format.startsWith(".") && it.indexState.isNotBlank() })
    }

    @Test
    fun `getDocuments decodes every contract field of the backend golden last page`() = runTest {
        server.enqueue(MockResponse().setBody(backendGolden("documents_last_page.json")))

        val result = repository.getDocuments(page = 2)

        assertEquals("/api/v1/documents?page=2&page_size=50", server.takeRequest().path)
        val page = (result as AppResult.Success).data
        assertFalse(page.hasNext)
        assertEquals(
            DocumentDto(
                id = 3,
                name = "documento.pdf",
                format = ".pdf",
                parentPath = "/Casa",
                size = 1024,
                modifiedAt = "2024-03-05T10:00:00Z",
                createdAt = "2024-03-05T10:00:00Z",
                starred = true,
                indexState = "indexed",
            ),
            page.items.single(),
        )
    }

    @Test
    fun `getDocuments sends the format and folder filters when given`() = runTest {
        server.enqueue(MockResponse().setBody(backendGolden("documents_last_page.json")))

        repository.getDocuments(page = 1, format = ".pdf", folderId = 7)

        assertEquals("/api/v1/documents?page=1&page_size=50&format=.pdf&folder_id=7", server.takeRequest().path)
    }

    @Test
    fun `getDocumentsCount reads the total`() = runTest {
        server.enqueue(MockResponse().setBody("""{"total":128}"""))

        val result = repository.getDocumentsCount()

        assertEquals("/api/v1/documents/count", server.takeRequest().path)
        assertEquals(128, (result as AppResult.Success).data)
    }

    @Test
    fun `searchDocuments reads the contract fields of a result`() = runTest {
        server.enqueue(
            MockResponse().setBody(
                """{"items":[{"file_id":9,"name":"contrato.pdf","path":"/Casa/contrato.pdf","parent_path":"/Casa",
                    "format":".pdf","size":2048,"updated_at":"2024-03-05T10:00:00Z","starred":false,
                    "snippet":"cláusula <mark>quinta</mark>","page":4}],
                    "pagination":{"page":1,"page_size":50,"has_next":false,"has_prev":false}}""",
            ),
        )

        val result = repository.searchDocuments("quinta", page = 1)

        assertTrue(server.takeRequest().path!!.startsWith("/api/v1/documents/search?q=quinta&page=1&page_size=50"))
        assertEquals(
            DocumentSearchResultDto(
                fileId = 9,
                name = "contrato.pdf",
                path = "/Casa/contrato.pdf",
                parentPath = "/Casa",
                format = ".pdf",
                size = 2048,
                updatedAt = "2024-03-05T10:00:00Z",
                starred = false,
                snippet = "cláusula <mark>quinta</mark>",
                page = 4,
            ),
            (result as AppResult.Success).data.items.single(),
        )
    }

    @Test
    fun `a refused request shows the server message verbatim`() = runTest {
        server.enqueue(MockResponse().setResponseCode(403).setBody("""{"error":"acesso negado a este IP"}"""))

        val result = repository.getDocuments(page = 1)

        assertEquals(AppResult.Error("acesso negado a este IP", 403), result)
    }
}
