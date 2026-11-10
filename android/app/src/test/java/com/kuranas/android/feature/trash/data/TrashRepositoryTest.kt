package com.kuranas.android.feature.trash.data

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

class TrashRepositoryTest {

    private val server = MockWebServer()
    private lateinit var repository: TrashRepository

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
            .create(TrashApi::class.java)
        repository = TrashRepository(api)
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    @Test
    fun `getTrashItems asks for the page and reads the wire fields`() = runTest {
        server.enqueue(
            MockResponse().setBody(
                """{"items":[{"id":3,"name":"a.pdf","original_path":"/data/docs/a.pdf","size":2048,
                    "deleted_at":"2026-06-11T10:00:00Z","expires_at":"2026-07-11T10:00:00Z","is_directory":true}],
                    "pagination":{"page":2,"page_size":50,"has_next":true,"has_prev":true}}""",
            ),
        )

        val result = repository.getTrashItems(2)

        assertEquals("/api/v1/trash?page=2&page_size=50", server.takeRequest().path)
        val page = (result as AppResult.Success).data
        assertTrue(page.hasNext)
        assertEquals(
            TrashItemDto(
                id = 3,
                name = "a.pdf",
                originalPath = "/data/docs/a.pdf",
                size = 2048,
                deletedAt = "2026-06-11T10:00:00Z",
                expiresAt = "2026-07-11T10:00:00Z",
                isDirectory = true,
            ),
            page.items.single(),
        )
    }

    @Test
    fun `getTrashItems survives an answer without items or pagination`() = runTest {
        server.enqueue(MockResponse().setBody("{}"))

        val page = (repository.getTrashItems(1) as AppResult.Success).data

        assertTrue(page.items.isEmpty())
        assertFalse(page.hasNext)
    }

    @Test
    fun `restoreItem posts to the item restore route`() = runTest {
        server.enqueue(MockResponse().setBody("""{"message":"ok","path":"/data/docs/a.pdf"}"""))

        val result = repository.restoreItem(3)

        val request = server.takeRequest()
        assertEquals("POST", request.method)
        assertEquals("/api/v1/trash/3/restore", request.path)
        assertTrue(result is AppResult.Success)
    }

    @Test
    fun `restoreItem surfaces the server message of a conflict`() = runTest {
        server.enqueue(
            MockResponse().setResponseCode(409).setBody("""{"error":"local ocupado","path":"/data/docs/a.pdf"}"""),
        )

        val result = repository.restoreItem(3)

        assertEquals("local ocupado", (result as AppResult.Error).message)
        assertEquals(409, result.code)
    }

    @Test
    fun `deleteItemForever sends DELETE to the item route`() = runTest {
        server.enqueue(MockResponse().setBody("""{"message":"ok"}"""))

        repository.deleteItemForever(3)

        val request = server.takeRequest()
        assertEquals("DELETE", request.method)
        assertEquals("/api/v1/trash/3", request.path)
    }

    @Test
    fun `emptyTrash sends DELETE to the trash root`() = runTest {
        server.enqueue(MockResponse().setBody("""{"message":"ok","purged":2,"failed":0}"""))

        repository.emptyTrash()

        val request = server.takeRequest()
        assertEquals("DELETE", request.method)
        assertEquals("/api/v1/trash", request.path)
    }
}
