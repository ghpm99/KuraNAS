package com.kuranas.android.core.server

import kotlinx.coroutines.runBlocking
import kotlinx.serialization.json.Json
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test
import java.util.concurrent.TimeUnit

class ServerHealthCheckerTest {

    private lateinit var server: MockWebServer
    private val checker = ServerHealthChecker(Json)

    @Before
    fun startServer() {
        server = MockWebServer()
        server.start()
    }

    @After
    fun stopServer() {
        runCatching { server.shutdown() }
    }

    private fun baseUrl(): String = server.url("/").toString()

    private fun check(baseUrl: String = baseUrl()): ServerCheckResult =
        runBlocking { checker.check(baseUrl) }

    @Test
    fun `200 com payload esperado e campos extras confirma`() {
        server.enqueue(MockResponse().setBody("""{"status":"ok","service":"kuranas","version":"1"}"""))
        assertEquals(ServerCheckResult.Confirmed, check())
    }

    @Test
    fun `pede o path api v1 health`() {
        server.enqueue(MockResponse().setBody("""{"status":"ok","service":"kuranas"}"""))
        check()
        assertEquals("/api/v1/health", server.takeRequest().path)
    }

    @Test
    fun `base url sem barra final funciona`() {
        server.enqueue(MockResponse().setBody("""{"status":"ok","service":"kuranas"}"""))
        assertEquals(ServerCheckResult.Confirmed, check(baseUrl().trimEnd('/')))
        assertEquals("/api/v1/health", server.takeRequest().path)
    }

    @Test
    fun `service diferente nao e KuraNAS`() {
        server.enqueue(MockResponse().setBody("""{"status":"ok","service":"outro"}"""))
        assertEquals(ServerCheckResult.NotKuraNas, check())
    }

    @Test
    fun `200 com html nao e KuraNAS`() {
        server.enqueue(MockResponse().setBody("<html><body>router</body></html>"))
        assertEquals(ServerCheckResult.NotKuraNas, check())
    }

    @Test
    fun `404 nao e KuraNAS`() {
        server.enqueue(MockResponse().setResponseCode(404))
        assertEquals(ServerCheckResult.NotKuraNas, check())
    }

    @Test
    fun `403 com ip e mensagem no corpo carrega os dois`() {
        server.enqueue(MockResponse().setResponseCode(403).setBody("""{"error":"Blocked 10.0.0.9","ip":"10.0.0.9"}"""))
        assertEquals(ServerCheckResult.AccessDenied(AccessDenial("10.0.0.9", "Blocked 10.0.0.9")), check())
    }

    @Test
    fun `403 sem corpo legivel e acesso negado com ip desconhecido`() {
        server.enqueue(MockResponse().setResponseCode(403).setBody("<html>no</html>"))
        assertEquals(ServerCheckResult.AccessDenied(AccessDenial(null, null)), check())
    }

    @Test
    fun `servidor desligado nao responde`() {
        val offlineUrl = baseUrl()
        server.shutdown()
        assertEquals(ServerCheckResult.NoResponse, check(offlineUrl))
    }

    @Test
    fun `resposta atrasada alem de 3 segundos nao responde`() {
        server.enqueue(
            MockResponse()
                .setBody("""{"status":"ok","service":"kuranas"}""")
                .setBodyDelay(4, TimeUnit.SECONDS),
        )
        assertEquals(ServerCheckResult.NoResponse, check())
    }

    @Test
    fun `url invalida nao responde`() {
        assertEquals(ServerCheckResult.NoResponse, check("nao e url"))
    }
}
