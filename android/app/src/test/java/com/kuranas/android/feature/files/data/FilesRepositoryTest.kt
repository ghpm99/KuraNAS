package com.kuranas.android.feature.files.data

import android.content.Context
import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import com.kuranas.android.core.network.AppResult
import io.mockk.mockk
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

class FilesRepositoryTest {

    private val server = MockWebServer()
    private lateinit var repository: FilesRepository

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
            .create(FilesApi::class.java)
        repository = FilesRepository(api, mockk<Context>())
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    @Test
    fun `setFileStarred sends PUT with the starred body`() = runTest {
        server.enqueue(MockResponse().setBody("""{"id":7,"starred":true}"""))

        repository.setFileStarred("7", true)

        val request = server.takeRequest()
        assertEquals("PUT", request.method)
        assertEquals("/api/v1/files/starred/7", request.path)
        assertEquals("""{"starred":true}""", request.body.readUtf8())
    }

    @Test
    fun `setFileStarred deserializes the returned starred state`() = runTest {
        server.enqueue(MockResponse().setBody("""{"id":7,"starred":false}"""))

        val result = repository.setFileStarred("7", false)

        val response = (result as AppResult.Success).data
        assertEquals(7, response.id)
        assertEquals(false, response.starred)
    }

    @Test
    fun `setFileStarred turns a server error into AppResult Error with the server message`() = runTest {
        server.enqueue(MockResponse().setResponseCode(404).setBody("""{"error":"Arquivo não encontrado"}"""))

        val result = repository.setFileStarred("7", true)

        assertTrue(result is AppResult.Error)
        assertEquals("Arquivo não encontrado", (result as AppResult.Error).message)
        assertEquals(404, result.code)
    }

    @Test
    fun `getStarredFiles sends exact page, page size, kind and sort`() = runTest {
        server.enqueue(MockResponse().setBody("""{"items":[],"pagination":{"has_next":false}}"""))

        repository.getStarredFiles(3, StarredKind.MEDIA, StarredSort.STARRED_AT)

        val request = server.takeRequest()
        assertEquals("GET", request.method)
        assertEquals("/api/v1/files/starred?page=3&page_size=50&kind=media&sort=starred_at", request.path)
    }

    @Test
    fun `getStarredFiles omits kind and sort when not provided`() = runTest {
        server.enqueue(MockResponse().setBody("""{"items":[],"pagination":{"has_next":false}}"""))

        repository.getStarredFiles(1)

        assertEquals("/api/v1/files/starred?page=1&page_size=50", server.takeRequest().path)
    }

    @Test
    fun `getStarredFiles maps items and has_next`() = runTest {
        server.enqueue(
            MockResponse().setBody("""{"items":[{"id":4,"name":"a","starred":true}],"pagination":{"has_next":true}}"""),
        )

        val result = repository.getStarredFiles(1) as AppResult.Success

        assertEquals(listOf("4"), result.data.items.map { it.id })
        assertTrue(result.data.hasNext)
    }

    private val firstPageWithNextJson = """
        {
          "items": [
            {"id": 11, "name": "docs", "path": "/data/docs", "parent_path": "/data", "type": 1, "format": "", "size": 0,
             "starred": false, "directory_content_count": 3, "created_at": "2026-01-01T10:00:00Z", "updated_at": "2026-01-02T10:00:00Z"},
            {"id": 12, "name": "a.txt", "path": "/data/a.txt", "parent_path": "/data", "type": 2, "format": ".txt", "size": 120,
             "starred": true, "directory_content_count": 0, "created_at": "2026-01-01T10:00:00Z", "updated_at": "2026-01-02T10:00:00Z"}
          ],
          "pagination": {"page": 1, "page_size": 2, "has_next": true, "has_prev": false}
        }
    """.trimIndent()

    private val lastPageJson = """
        {"items":[{"id":13,"name":"b.txt","type":2}],"pagination":{"page":2,"page_size":50,"has_next":false,"has_prev":true}}
    """.trimIndent()

    @Test
    fun `getRootFiles sends page and page size and maps the page with next`() = runTest {
        server.enqueue(MockResponse().setBody(firstPageWithNextJson))

        val result = repository.getRootFiles(1) as AppResult.Success

        val request = server.takeRequest()
        assertEquals("GET", request.method)
        assertEquals("/api/v1/files/?page=1&page_size=50", request.path)
        assertEquals(listOf("11", "12"), result.data.items.map { it.id })
        assertTrue(result.data.items.first().isDir)
        assertTrue(result.data.hasNext)
    }

    @Test
    fun `getChildrenById sends id, page and page size and reports the last page`() = runTest {
        server.enqueue(MockResponse().setBody(lastPageJson))

        val result = repository.getChildrenById("11", 2) as AppResult.Success

        assertEquals("/api/v1/files/11?page=2&page_size=50", server.takeRequest().path)
        assertEquals(listOf("13"), result.data.items.map { it.id })
        assertFalse(result.data.hasNext)
    }

    @Test
    fun `getFileInfo asks for a single item page`() = runTest {
        server.enqueue(MockResponse().setBody(firstPageWithNextJson))

        val result = repository.getFileInfo("12") as AppResult.Success

        assertEquals("/api/v1/files/12?page=1&page_size=1", server.takeRequest().path)
        assertEquals("12", result.data.id)
    }
}
