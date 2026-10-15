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
}
