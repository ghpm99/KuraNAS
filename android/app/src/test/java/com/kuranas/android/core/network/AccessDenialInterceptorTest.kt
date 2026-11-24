package com.kuranas.android.core.network

import com.kuranas.android.core.server.AccessDenial
import com.kuranas.android.core.server.AccessDeniedEvents
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Before
import org.junit.Test

class AccessDenialInterceptorTest {

    private lateinit var server: MockWebServer
    private val accessDeniedEvents = AccessDeniedEvents()
    private val client = OkHttpClient.Builder()
        .addInterceptor(AccessDenialInterceptor(accessDeniedEvents, Json { ignoreUnknownKeys = true }))
        .build()

    @Before
    fun startServer() {
        server = MockWebServer()
        server.start()
    }

    @After
    fun stopServer() {
        server.shutdown()
    }

    private fun fetchBody(): String =
        client.newCall(Request.Builder().url(server.url("/api/v1/files")).build()).execute().use { it.body!!.string() }

    @Test
    fun `403 com ip publica a recusa e preserva o corpo para quem chamou`() {
        val body = """{"error":"Blocked 10.0.0.9","ip":"10.0.0.9"}"""
        server.enqueue(MockResponse().setResponseCode(403).setBody(body))

        assertEquals(body, fetchBody())
        assertEquals(AccessDenial("10.0.0.9", "Blocked 10.0.0.9"), accessDeniedEvents.denial.value)
    }

    @Test
    fun `403 sem ip nao publica recusa`() {
        server.enqueue(MockResponse().setResponseCode(403).setBody("""{"error":"forbidden"}"""))

        fetchBody()

        assertNull(accessDeniedEvents.denial.value)
    }

    @Test
    fun `403 com corpo que nao e json nao lanca nem publica`() {
        server.enqueue(MockResponse().setResponseCode(403).setBody("<html>no</html>"))

        assertEquals("<html>no</html>", fetchBody())
        assertNull(accessDeniedEvents.denial.value)
    }

    @Test
    fun `resposta de sucesso nao publica recusa`() {
        server.enqueue(MockResponse().setBody("""{"ip":"10.0.0.9"}"""))

        fetchBody()

        assertNull(accessDeniedEvents.denial.value)
    }
}
