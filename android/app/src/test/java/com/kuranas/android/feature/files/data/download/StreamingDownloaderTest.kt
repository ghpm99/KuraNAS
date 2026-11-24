package com.kuranas.android.feature.files.data.download

import android.net.Uri
import com.kuranas.android.feature.files.data.FilesApi
import io.mockk.every
import io.mockk.mockk
import kotlinx.coroutines.test.runTest
import okhttp3.mockwebserver.Dispatcher
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.RecordedRequest
import okhttp3.mockwebserver.SocketPolicy
import okio.Buffer
import org.junit.After
import org.junit.Assert.assertArrayEquals
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder
import retrofit2.Retrofit
import java.io.ByteArrayOutputStream
import java.io.File
import java.io.OutputStream

class StreamingDownloaderTest {

    @get:Rule
    val temporaryFolder = TemporaryFolder()

    private val server = MockWebServer()
    private val committedUri = mockk<Uri>()
    private val fileBytes = ByteArray(3 * 1024 * 1024) { (it * 31 % 251).toByte() }
    private val request = DownloadRequest("7", "relatório final.pdf", fileBytes.size.toLong())
    private var mimeTypeFromExtension: String? = "application/pdf"
    private val mimeTypeResolver = MimeTypeResolver { mimeTypeFromExtension }
    private lateinit var downloader: StreamingDownloader
    private lateinit var destinationFile: File
    private lateinit var destination: CachedFileDownload
    private val destinations = PendingDownloadFactory { mimeType ->
        requestedMimeTypes += mimeType
        destination
    }
    private val requestedMimeTypes = mutableListOf<String>()

    @Before
    fun setUp() {
        val api = Retrofit.Builder().baseUrl(server.url("/")).build().create(FilesApi::class.java)
        downloader = StreamingDownloader(api, mimeTypeResolver, retryDelayMillis = 0)
        destinationFile = File(temporaryFolder.root, "shared/7/${request.fileName}")
        destination = CachedFileDownload(destinationFile) { committedUri }
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    private val receivedRequests = mutableListOf<RecordedRequest>()

    private fun serve(respond: (attempt: Int, rangeStart: Long?) -> MockResponse) {
        var attempt = 0
        server.dispatcher = object : Dispatcher() {
            override fun dispatch(request: RecordedRequest): MockResponse {
                attempt += 1
                receivedRequests += request
                val rangeStart = request.getHeader("Range")?.removePrefix("bytes=")?.removeSuffix("-")?.toLong()
                return respond(attempt, rangeStart)
            }
        }
    }

    private fun fullResponse(shouldDisconnect: Boolean = false) = MockResponse()
        .setBody(Buffer().write(fileBytes))
        .apply { if (shouldDisconnect) socketPolicy = SocketPolicy.DISCONNECT_DURING_RESPONSE_BODY }

    private fun partialResponse(firstByte: Long) = MockResponse()
        .setResponseCode(206)
        .setHeader("Accept-Ranges", "bytes")
        .setHeader("Content-Range", "bytes $firstByte-${fileBytes.size - 1}/${fileBytes.size}")
        .setBody(Buffer().write(fileBytes, firstByte.toInt(), fileBytes.size - firstByte.toInt()))

    private fun requestedRanges(): List<String?> = receivedRequests.map { it.getHeader("Range") }

    @Test
    fun `copies the body in bounded chunks and reports growing progress`() = runTest {
        serve { _, _ -> fullResponse() }
        val writtenChunkSizes = mutableListOf<Int>()
        val written = ByteArrayOutputStream()
        val countingDestination = object : PendingDownload {
            override fun openOutput(shouldAppend: Boolean): OutputStream = object : OutputStream() {
                override fun write(byte: Int) = written.write(byte)
                override fun write(buffer: ByteArray, offset: Int, length: Int) {
                    writtenChunkSizes += length
                    written.write(buffer, offset, length)
                }
            }
            override fun commit(): Uri = committedUri
            override fun discard() = Unit
        }
        val progressUpdates = mutableListOf<DownloadProgress>()

        val outcome = downloader.download(request, PendingDownloadFactory { countingDestination }) { progressUpdates += it }

        assertEquals(DownloadOutcome.Completed(committedUri, "application/pdf"), outcome)
        assertArrayEquals(fileBytes, written.toByteArray())
        assertTrue(writtenChunkSizes.max() <= 64 * 1024)
        assertTrue(writtenChunkSizes.size > 40)
        assertEquals(progressUpdates.map { it.receivedBytes }, progressUpdates.map { it.receivedBytes }.sorted())
        assertEquals(DownloadProgress(fileBytes.size.toLong(), fileBytes.size.toLong()), progressUpdates.last())
    }

    @Test
    fun `resumes after the connection drops from the received byte on the same destination`() = runTest {
        serve { attempt, rangeStart ->
            if (attempt == 1) fullResponse(shouldDisconnect = true) else partialResponse(rangeStart!!)
        }

        val outcome = downloader.download(request, destinations) {}

        assertEquals(DownloadOutcome.Completed(committedUri, "application/pdf"), outcome)
        assertArrayEquals(fileBytes, destinationFile.readBytes())
        val ranges = requestedRanges()
        assertNull(ranges[0])
        val resumedFrom = ranges[1]!!.removePrefix("bytes=").removeSuffix("-").toLong()
        assertTrue(resumedFrom > 0 && resumedFrom < fileBytes.size)
    }

    @Test
    fun `restarts from zero when the server answers 200 to a range request`() = runTest {
        serve { attempt, _ -> fullResponse(shouldDisconnect = attempt == 1) }

        val outcome = downloader.download(request, destinations) {}

        assertEquals(DownloadOutcome.Completed(committedUri, "application/pdf"), outcome)
        assertEquals(fileBytes.size.toLong(), destinationFile.length())
        assertArrayEquals(fileBytes, destinationFile.readBytes())
        assertEquals(2, server.requestCount)
    }

    private fun droppedBeforeAnyByte() = MockResponse().setSocketPolicy(SocketPolicy.DISCONNECT_AFTER_REQUEST)

    @Test
    fun `fails with a connection error after three drops in a row without any new byte`() = runTest {
        serve { _, _ -> droppedBeforeAnyByte() }

        val outcome = downloader.download(request, destinations) {}

        assertEquals(DownloadOutcome.Failed(DownloadFailure.ConnectionLost), outcome)
        assertEquals(3, server.requestCount)
        assertFalse(destinationFile.exists())
    }

    @Test
    fun `three fruitless drops after some progress still fail and leave no file`() = runTest {
        serve { attempt, _ -> if (attempt == 1) fullResponse(shouldDisconnect = true) else droppedBeforeAnyByte() }

        val outcome = downloader.download(request, destinations) {}

        assertEquals(DownloadOutcome.Failed(DownloadFailure.ConnectionLost), outcome)
        assertEquals(4, server.requestCount)
        assertFalse(destinationFile.exists())
    }

    @Test
    fun `four drops in a row that each advanced still complete the whole file`() = runTest {
        serve { attempt, rangeStart ->
            when {
                attempt == 1 -> fullResponse(shouldDisconnect = true)
                attempt <= 4 -> partialResponse(rangeStart!!).setSocketPolicy(SocketPolicy.DISCONNECT_DURING_RESPONSE_BODY)
                else -> partialResponse(rangeStart!!)
            }
        }

        val outcome = downloader.download(request, destinations) {}

        assertEquals(DownloadOutcome.Completed(committedUri, "application/pdf"), outcome)
        assertEquals(5, server.requestCount)
        assertArrayEquals(fileBytes, destinationFile.readBytes())
    }

    @Test
    fun `the destination mime type comes from the file extension`() = runTest {
        serve { _, _ -> fullResponse().setHeader("Content-Type", "text/plain") }

        downloader.download(request, destinations) {}

        assertEquals(listOf("application/pdf"), requestedMimeTypes)
    }

    @Test
    fun `an unknown extension falls back to the response content type without parameters`() = runTest {
        mimeTypeFromExtension = null
        serve { _, _ -> fullResponse().setHeader("Content-Type", "Video/x-Matroska; charset=binary") }

        val outcome = downloader.download(request, destinations) {}

        assertEquals(listOf("video/x-matroska"), requestedMimeTypes)
        assertEquals(DownloadOutcome.Completed(committedUri, "video/x-matroska"), outcome)
    }

    @Test
    fun `shows the server message of a 404 without retrying or creating the file`() = runTest {
        serve { _, _ ->
            MockResponse().setResponseCode(404).setBody("""{"error":"Arquivo não encontrado"}""")
        }

        val outcome = downloader.download(request, destinations) {}

        assertEquals(DownloadOutcome.Failed(DownloadFailure.Server("Arquivo não encontrado", 404)), outcome)
        assertEquals(1, server.requestCount)
        assertFalse(destinationFile.exists())
    }

    @Test
    fun `sends no range header on the first request`() = runTest {
        serve { _, _ -> fullResponse() }

        downloader.download(request, destinations) {}

        assertEquals(listOf<String?>(null), requestedRanges())
        assertEquals("/api/v1/files/download/7", receivedRequests.single().path)
    }

    private fun resumeAt(offset: Int): DownloadResume {
        destinationFile.parentFile!!.mkdirs()
        destinationFile.writeBytes(fileBytes.copyOfRange(0, offset))
        return DownloadResume(destination, offset.toLong(), "application/pdf")
    }

    @Test
    fun `starts at a persisted offset with a range header and appends to the reopened destination`() = runTest {
        val offset = 1_000_000
        serve { _, rangeStart -> partialResponse(rangeStart!!) }
        val progressUpdates = mutableListOf<DownloadProgress>()

        val outcome = downloader.download(request, destinations, resume = resumeAt(offset)) { progressUpdates += it }

        assertEquals(DownloadOutcome.Completed(committedUri, "application/pdf"), outcome)
        assertEquals(listOf<String?>("bytes=$offset-"), requestedRanges())
        assertArrayEquals(fileBytes, destinationFile.readBytes())
        assertTrue(progressUpdates.first().receivedBytes > offset)
        assertTrue(requestedMimeTypes.isEmpty())
    }

    @Test
    fun `restarts from zero when the server ignores the range of a persisted offset`() = runTest {
        serve { _, _ -> fullResponse() }

        val outcome = downloader.download(request, destinations, resume = resumeAt(1_000_000)) {}

        assertEquals(DownloadOutcome.Completed(committedUri, "application/pdf"), outcome)
        assertArrayEquals(fileBytes, destinationFile.readBytes())
    }

    @Test
    fun `reports the destination reference once before the first progress`() = runTest {
        serve { _, _ -> fullResponse() }
        val events = mutableListOf<String>()

        downloader.download(
            request,
            destinations,
            onDestinationReady = { reference, mimeType -> events += "ready:$reference:$mimeType" },
        ) { if (events.size < 2) events += "progress" }

        assertEquals(listOf("ready:${destinationFile.absolutePath}:application/pdf", "progress"), events)
    }

    @Test
    fun `keeps the partial file when the connection is lost and a later resume is expected`() = runTest {
        serve { attempt, rangeStart ->
            if (attempt == 1) fullResponse(shouldDisconnect = true) else droppedBeforeAnyByte()
        }

        val outcome = downloader.download(request, destinations, shouldKeepPartialForLaterResume = true) {}

        assertEquals(DownloadOutcome.Failed(DownloadFailure.ConnectionLost), outcome)
        assertTrue(destinationFile.exists())
        assertTrue(destinationFile.length() > 0)
    }

    @Test
    fun `a missing persisted destination makes the resumable executor start from zero without a range`() = runTest {
        serve { _, _ -> fullResponse() }
        val restorer = mockk<DownloadDestinationRestorer>()
        every { restorer.restore(any(), any(), any()) } returns null
        val cacheDestinations = mockk<CacheDownloadDestinations>()
        every { cacheDestinations.create(any()) } returns destination
        val executor = ResumableDownloadExecutor(downloader, mockk(), cacheDestinations, restorer)
        val sharedRequest = request.copy(purpose = DownloadPurpose.SHARE)

        val outcome = executor.execute(sharedRequest, PersistedDownloadDestination("gone", "application/pdf"), { _, _ -> }) {}

        assertEquals(DownloadOutcome.Completed(committedUri, "application/pdf"), outcome)
        assertEquals(listOf<String?>(null), requestedRanges())
        assertArrayEquals(fileBytes, destinationFile.readBytes())
    }

    @Test
    fun `a restored destination resumes the resumable executor from its written bytes`() = runTest {
        val offset = 500_000
        serve { _, rangeStart -> partialResponse(rangeStart!!) }
        val resume = resumeAt(offset)
        val restorer = mockk<DownloadDestinationRestorer>()
        every { restorer.restore(any(), any(), any()) } returns RestoredDownload(resume.destination, offset.toLong())
        val executor = ResumableDownloadExecutor(downloader, mockk(), mockk(), restorer)

        val outcome = executor.execute(
            request.copy(purpose = DownloadPurpose.SHARE),
            PersistedDownloadDestination(destinationFile.absolutePath, "application/pdf"),
            { _, _ -> },
        ) {}

        assertEquals(DownloadOutcome.Completed(committedUri, "application/pdf"), outcome)
        assertEquals(listOf<String?>("bytes=$offset-"), requestedRanges())
        assertArrayEquals(fileBytes, destinationFile.readBytes())
    }
}
