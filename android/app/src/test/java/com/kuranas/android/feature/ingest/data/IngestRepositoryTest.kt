package com.kuranas.android.feature.ingest.data

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import com.kuranas.android.core.network.AppResult
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import retrofit2.Retrofit

class IngestRepositoryTest {

    private val server = MockWebServer()
    private lateinit var repository: IngestRepository

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
            .create(IngestApi::class.java)
        repository = IngestRepository(api)
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    @Test
    fun `getTargets reads label and path`() = runTest {
        server.enqueue(MockResponse().setBody("""[{"label":"Vídeos","path":"/data/videos"}]"""))

        val targets = (repository.getTargets() as AppResult.Success).data

        assertEquals("/api/v1/ingest/targets", server.takeRequest().path)
        assertEquals(listOf(IngestTargetDto(label = "Vídeos", path = "/data/videos")), targets)
    }

    @Test
    fun `getPresets reads key and label`() = runTest {
        server.enqueue(MockResponse().setBody("""[{"key":"audio_mp3","label":"Áudio"}]"""))

        val presets = (repository.getPresets() as AppResult.Success).data

        assertEquals("/api/v1/ingest/presets", server.takeRequest().path)
        assertEquals(listOf(IngestPresetDto(key = "audio_mp3", label = "Áudio")), presets)
    }

    @Test
    fun `requestDownload posts the wire fields and reads the ids`() = runTest {
        server.enqueue(MockResponse().setResponseCode(202).setBody("""{"job_id":345,"download_id":12}"""))

        val answer = repository.requestDownload("https://exemplo.com/v", "video_1080", "/data/videos", "Shows")

        val request = server.takeRequest()
        assertEquals("POST", request.method)
        assertEquals("/api/v1/ingest/fetch", request.path)
        val body = Json.parseToJsonElement(request.body.readUtf8()).toString()
        assertEquals(
            """{"url":"https://exemplo.com/v","preset":"video_1080","target_root":"/data/videos","subfolder":"Shows"}""",
            body,
        )
        assertEquals(IngestFetchResponseDto(jobId = 345, downloadId = 12), (answer as AppResult.Success).data)
    }

    @Test
    fun `requestDownload surfaces the server message of a rejected url`() = runTest {
        server.enqueue(MockResponse().setResponseCode(400).setBody("""{"error":"endereço inválido"}"""))

        val answer = repository.requestDownload("x", "video_1080", "/data/videos", "")

        assertEquals("endereço inválido", (answer as AppResult.Error).message)
    }

    @Test
    fun `getDownloads asks for the page and reads the wire fields with nulls`() = runTest {
        server.enqueue(
            MockResponse().setBody(
                """{"items":[{"id":12,"job_id":345,"url":"https://exemplo.com/video","title":null,
                    "preset":"video_1080","target_root_id":2,"target_root_name":"Vídeos","subfolder":"Shows",
                    "state":"running","progress":42,"speed_bytes_per_second":1048576,"eta_seconds":95,
                    "file_id":null,"error":"","created_at":"2026-11-16T20:15:00-03:00","ended_at":null}],
                    "pagination":{"page":2,"page_size":20,"has_next":true,"has_prev":true}}""",
            ),
        )

        val page = (repository.getDownloads(2) as AppResult.Success).data

        assertEquals("/api/v1/ingest/downloads?page=2&page_size=20", server.takeRequest().path)
        assertTrue(page.hasNext)
        val download = page.items.single()
        assertEquals(12, download.id)
        assertEquals(345, download.jobId)
        assertNull(download.title)
        assertEquals(2, download.targetRootId)
        assertEquals("Vídeos", download.targetRootName)
        assertEquals(1048576L, download.speedBytesPerSecond)
        assertEquals(95, download.etaSeconds)
        assertNull(download.fileId)
        assertNull(download.endedAt)
    }

    @Test
    fun `cancelDownload posts to the job cancel route`() = runTest {
        server.enqueue(MockResponse().setBody("""{"message":"ok"}"""))

        val answer = repository.cancelDownload(345)

        val request = server.takeRequest()
        assertEquals("POST", request.method)
        assertEquals("/api/v1/jobs/345/cancel", request.path)
        assertTrue(answer is AppResult.Success)
    }

    @Test
    fun `retryDownload posts to the download retry route`() = runTest {
        server.enqueue(MockResponse().setResponseCode(201).setBody("""{"id":13}"""))

        val answer = repository.retryDownload(12)

        val request = server.takeRequest()
        assertEquals("POST", request.method)
        assertEquals("/api/v1/ingest/downloads/12/retry", request.path)
        assertTrue(answer is AppResult.Success)
    }

    @Test
    fun `retryDownload surfaces the conflict message`() = runTest {
        server.enqueue(MockResponse().setResponseCode(409).setBody("""{"error":"pedido não pode ser repetido"}"""))

        val answer = repository.retryDownload(12)

        assertEquals("pedido não pode ser repetido", (answer as AppResult.Error).message)
        assertEquals(409, answer.code)
    }
}
