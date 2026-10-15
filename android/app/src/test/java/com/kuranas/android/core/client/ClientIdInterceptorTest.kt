package com.kuranas.android.core.client

import io.mockk.every
import io.mockk.mockk
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Before
import org.junit.Test

class ClientIdInterceptorTest {

    private val server = MockWebServer()
    private val expectedClientId = "11111111-2222-3333-4444-555555555555"
    private lateinit var client: OkHttpClient

    @Before
    fun setUp() {
        server.start()
        val provider = mockk<ClientIdProvider>()
        every { provider.clientId() } returns expectedClientId
        client = OkHttpClient.Builder().addInterceptor(ClientIdInterceptor(provider)).build()
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    private fun sentHeaderFor(path: String): String? {
        server.enqueue(MockResponse().setBody("{}"))
        client.newCall(Request.Builder().url(server.url(path)).build()).execute().close()
        return server.takeRequest().getHeader(CLIENT_ID_HEADER)
    }

    @Test
    fun `video requests carry the client id header`() {
        assertEquals(expectedClientId, sentHeaderFor("/api/v1/video/playlists"))
    }

    @Test
    fun `music requests carry the client id header`() {
        assertEquals(expectedClientId, sentHeaderFor("/api/v1/music/playlists"))
    }

    @Test
    fun `file requests do not carry the client id header`() {
        assertNull(sentHeaderFor("/api/v1/files/tree"))
    }
}
