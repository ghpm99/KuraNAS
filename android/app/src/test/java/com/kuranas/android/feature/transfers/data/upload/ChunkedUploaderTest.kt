package com.kuranas.android.feature.transfers.data.upload

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.RecordedRequest
import okhttp3.mockwebserver.SocketPolicy
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import retrofit2.Retrofit

class ChunkedUploaderTest {

    private class InMemoryUploadSource(
        private val content: ByteArray,
        private val isMissing: Boolean = false,
    ) : UploadSource {
        override suspend fun sizeBytes(): Long = if (isMissing) throw UploadSourceUnavailableException() else content.size.toLong()
        override suspend fun modifiedAtMillis(): Long? = 1_700_000_000_000
        override suspend fun sha256Hex(): String = "abc123"
        override suspend fun readChunk(offset: Long, length: Int): ByteArray =
            content.copyOfRange(offset.toInt(), offset.toInt() + length)
    }

    private val server = MockWebServer()
    private lateinit var uploader: ChunkedUploader
    private val progressEvents = mutableListOf<Pair<String, Long>>()
    private val progressListener = UploadProgressListener { uploadId, receivedBytes ->
        progressEvents += uploadId to receivedBytes
    }
    private val openGate = UploadChunkGate { }
    private val content = ByteArray(10) { it.toByte() }
    private val freshRequest = UploadRequest(uploadId = null, targetFolderId = 7, displayName = "photo.jpg")

    @Before
    fun setUp() {
        server.start()
        val json = Json { ignoreUnknownKeys = true }
        val retrofit = Retrofit.Builder()
            .baseUrl(server.url("/"))
            .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
            .build()
        uploader = ChunkedUploader(retrofit.create(TransferUploadApi::class.java))
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    private fun json(code: Int, body: String) =
        MockResponse().setResponseCode(code).setHeader("Content-Type", "application/json").setBody(body)

    private fun initResponse(chunkSize: Int = 4, receivedSize: Int = 0) =
        json(201, """{"upload_id":"up-1","chunk_size":$chunkSize,"received_size":$receivedSize,"resumed":false,"already_exists":false}""")

    private fun chunkResponse(receivedSize: Int) = json(200, """{"received_size":$receivedSize}""")

    private fun completeResponse() = json(201, """{"file":{"name":"photo.jpg","path":"/p","status":"uploaded"},"job_id":1}""")

    private fun chunkOffsets(requests: List<RecordedRequest>): List<String> =
        requests.filter { it.path == "/api/v1/files/upload/chunk" }.map { request ->
            val body = request.body.clone().readUtf8()
            Regex("name=\"offset\"[^\\r]*\\r\\n(?:[^\\r]+\\r\\n)*\\r\\n(\\d+)").find(body)!!.groupValues[1]
        }

    private fun drain(count: Int): List<RecordedRequest> = List(count) { server.takeRequest() }

    @Test
    fun freshUploadSendsChunksAtIncreasingOffsetsAndCompletes() = runTest {
        server.enqueue(initResponse())
        server.enqueue(chunkResponse(4))
        server.enqueue(chunkResponse(8))
        server.enqueue(chunkResponse(10))
        server.enqueue(completeResponse())

        val outcome = uploader.upload(InMemoryUploadSource(content), freshRequest, progressListener, openGate)

        assertEquals(UploadOutcome.Completed, outcome)
        val requests = drain(5)
        assertEquals(listOf("0", "4", "8"), chunkOffsets(requests))
        val initBody = requests[0].body.readUtf8()
        assertTrue(initBody, initBody.contains("\"checksum\":\"abc123\""))
        assertTrue(initBody, initBody.contains("\"on_conflict\":\"rename\""))
        assertTrue(initBody, initBody.contains("\"target_folder_id\":7"))
        assertTrue(initBody, initBody.contains("\"modified_at\":\"2023-11-14T22:13:20Z\""))
        assertEquals("/api/v1/files/upload/complete", requests[4].path)
        assertEquals("up-1" to 10L, progressEvents.last())
    }

    @Test
    fun resumeWithSavedUploadIdContinuesFromServerReceivedSize() = runTest {
        server.enqueue(json(200, """{"upload_id":"up-1","size":10,"received_size":4}"""))
        server.enqueue(chunkResponse(10))
        server.enqueue(completeResponse())

        val outcome = uploader.upload(
            InMemoryUploadSource(content),
            freshRequest.copy(uploadId = "up-1"),
            progressListener,
            openGate,
        )

        assertEquals(UploadOutcome.Completed, outcome)
        val requests = drain(3)
        assertEquals("/api/v1/files/upload/up-1", requests[0].path)
        assertEquals(listOf("4"), chunkOffsets(requests))
        assertEquals("up-1" to 4L, progressEvents.first())
    }

    @Test
    fun offsetConflictRealignsToReceivedSizeFromBody() = runTest {
        server.enqueue(initResponse(chunkSize = 5))
        server.enqueue(json(409, """{"error":"offset","received_size":2}"""))
        server.enqueue(chunkResponse(7))
        server.enqueue(chunkResponse(10))
        server.enqueue(completeResponse())

        val outcome = uploader.upload(InMemoryUploadSource(content), freshRequest, progressListener, openGate)

        assertEquals(UploadOutcome.Completed, outcome)
        assertEquals(listOf("0", "2", "7"), chunkOffsets(drain(5)))
    }

    @Test
    fun alreadyExistsShortCircuitsWithoutSendingChunks() = runTest {
        server.enqueue(json(201, """{"chunk_size":4,"received_size":10,"already_exists":true}"""))

        val outcome = uploader.upload(InMemoryUploadSource(content), freshRequest, progressListener, openGate)

        assertEquals(UploadOutcome.AlreadyExisted, outcome)
        assertEquals(1, server.requestCount)
    }

    @Test
    fun missingSessionOnResumeReinitializesFromZero() = runTest {
        server.enqueue(json(404, """{"error":"gone"}"""))
        server.enqueue(initResponse(chunkSize = 10))
        server.enqueue(chunkResponse(10))
        server.enqueue(completeResponse())

        val outcome = uploader.upload(
            InMemoryUploadSource(content),
            freshRequest.copy(uploadId = "stale"),
            progressListener,
            openGate,
        )

        assertEquals(UploadOutcome.Completed, outcome)
        val requests = drain(4)
        assertEquals("/api/v1/files/upload/init", requests[1].path)
        assertEquals(listOf("0"), chunkOffsets(requests))
    }

    @Test
    fun missingSessionOnChunkReinitializesOnce() = runTest {
        server.enqueue(initResponse(chunkSize = 10))
        server.enqueue(json(404, """{"error":"gone"}"""))
        server.enqueue(initResponse(chunkSize = 10))
        server.enqueue(chunkResponse(10))
        server.enqueue(completeResponse())

        val outcome = uploader.upload(InMemoryUploadSource(content), freshRequest, progressListener, openGate)

        assertEquals(UploadOutcome.Completed, outcome)
        assertEquals(5, server.requestCount)
    }

    @Test
    fun droppedConnectionIsConnectionLost() = runTest {
        server.enqueue(MockResponse().setSocketPolicy(SocketPolicy.DISCONNECT_AT_START))

        val outcome = uploader.upload(InMemoryUploadSource(content), freshRequest, progressListener, openGate)

        assertEquals(UploadOutcome.ConnectionLost, outcome)
    }

    @Test
    fun serverErrorIsConnectionLost() = runTest {
        server.enqueue(json(503, """{"error":"busy"}"""))

        val outcome = uploader.upload(InMemoryUploadSource(content), freshRequest, progressListener, openGate)

        assertEquals(UploadOutcome.ConnectionLost, outcome)
    }

    @Test
    fun insufficientStorageOnInitIsRejectedWithServerMessage() = runTest {
        server.enqueue(json(507, """{"error":"Sem espaço em disco"}"""))

        val outcome = uploader.upload(InMemoryUploadSource(content), freshRequest, progressListener, openGate)

        assertEquals(UploadOutcome.Rejected("Sem espaço em disco"), outcome)
    }

    @Test
    fun unavailableSourceIsSourceMissingWithoutCallingServer() = runTest {
        val outcome = uploader.upload(
            InMemoryUploadSource(content, isMissing = true),
            freshRequest,
            progressListener,
            openGate,
        )

        assertEquals(UploadOutcome.SourceMissing, outcome)
        assertEquals(0, server.requestCount)
    }
}
