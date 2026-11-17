package com.kuranas.android.feature.files.data

import android.content.ContentResolver
import android.content.Context
import android.database.Cursor
import android.net.Uri
import android.provider.OpenableColumns
import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import com.kuranas.android.core.network.AppResult
import io.mockk.every
import io.mockk.mockk
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.SocketPolicy
import okio.ByteString.Companion.toByteString
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import retrofit2.Retrofit
import java.io.InputStream

class UploadFilesRepositoryTest {

    private val server = MockWebServer()
    private val contentResolver = mockk<ContentResolver>()
    private val context = mockk<Context>()
    private val photoUri = mockk<Uri>()
    private val photoBytes = ByteArray(100_000) { (it % 251).toByte() }
    private lateinit var repository: FilesRepository

    private val serverResponseWithRenamedAndFailed = """
        {
          "message": "Upload realizado com sucesso",
          "uploaded": ["/data/Fotos/photo.jpg", "/data/Fotos/photo (1).jpg"],
          "job_id": 12,
          "files": [
            {"name": "photo.jpg", "path": "/data/Fotos/photo.jpg", "status": "uploaded"},
            {"name": "photo.jpg", "path": "/data/Fotos/photo (1).jpg", "status": "renamed"},
            {"name": "broken.jpg", "path": "", "status": "failed", "error": "Falha ao gravar o arquivo"}
          ]
        }
    """.trimIndent()

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
        every { context.contentResolver } returns contentResolver
        every { contentResolver.getType(photoUri) } returns "image/jpeg"
        repository = FilesRepository(api, context)
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    private fun selectedPhoto(sizeBytes: Long = photoBytes.size.toLong()) = SelectedFile(photoUri, "photo.jpg", sizeBytes)

    @Test
    fun `upload sends the files part with the file name, target folder and rename policy`() = runTest {
        every { contentResolver.openInputStream(photoUri) } answers { photoBytes.inputStream() }
        server.enqueue(MockResponse().setResponseCode(202).setBody(serverResponseWithRenamedAndFailed))

        repository.uploadFile(selectedPhoto(), targetFolderId = "42")

        val request = server.takeRequest()
        val body = request.body.readUtf8()
        assertEquals("POST", request.method)
        assertEquals("/api/v1/files/upload", request.path)
        assertTrue(body.contains("""name="files"; filename="photo.jpg""""))
        assertTrue(body.contains("Content-Type: image/jpeg"))
        assertTrue(Regex("""name="target_folder_id"\r\n(Content-[^\r]*\r\n)*\r\n42\r\n""").containsMatchIn(body))
        assertTrue(Regex("""name="on_conflict"\r\n(Content-[^\r]*\r\n)*\r\nrename\r\n""").containsMatchIn(body))
    }

    @Test
    fun `upload to the root sends zero as the target folder`() = runTest {
        every { contentResolver.openInputStream(photoUri) } answers { photoBytes.inputStream() }
        server.enqueue(MockResponse().setResponseCode(202).setBody(serverResponseWithRenamedAndFailed))

        repository.uploadFile(selectedPhoto(), targetFolderId = null)

        val body = server.takeRequest().body.readUtf8()
        assertTrue(Regex("""name="target_folder_id"\r\n(Content-[^\r]*\r\n)*\r\n0\r\n""").containsMatchIn(body))
    }

    @Test
    fun `upload decodes the server response with renamed and failed files`() = runTest {
        every { contentResolver.openInputStream(photoUri) } answers { photoBytes.inputStream() }
        server.enqueue(MockResponse().setResponseCode(202).setBody(serverResponseWithRenamedAndFailed))

        val response = (repository.uploadFile(selectedPhoto(), "42") as AppResult.Success).data

        assertEquals(12, response.jobId)
        assertEquals(listOf("/data/Fotos/photo.jpg", "/data/Fotos/photo (1).jpg"), response.uploaded)
        assertEquals(listOf("uploaded", "renamed", "failed"), response.files.map { it.status })
        assertEquals("/data/Fotos/photo (1).jpg", response.files[1].path)
        assertEquals("Falha ao gravar o arquivo", response.files[2].error)
    }

    @Test
    fun `upload reads the content through the stream only when the request is written`() = runTest {
        val countingStream = CountingInputStream(photoBytes)
        every { contentResolver.openInputStream(photoUri) } returns countingStream
        server.enqueue(MockResponse().setResponseCode(202).setBody(serverResponseWithRenamedAndFailed))

        repository.uploadFile(selectedPhoto(), "42")

        val body = server.takeRequest().body
        assertTrue(countingStream.readCalls > 1)
        assertTrue(countingStream.largestRead < photoBytes.size)
        assertTrue(countingStream.isClosed)
        assertTrue(body.indexOf(photoBytes.toByteString()) >= 0)
    }

    @Test
    fun `upload turns a forbidden response into an error carrying the server message`() = runTest {
        every { contentResolver.openInputStream(photoUri) } answers { photoBytes.inputStream() }
        server.enqueue(MockResponse().setResponseCode(403).setBody("""{"error":"IP não autorizado"}"""))

        val result = repository.uploadFile(selectedPhoto(), "42") as AppResult.Error

        assertEquals("IP não autorizado", result.message)
        assertEquals(403, result.code)
    }

    @Test
    fun `upload failing to reach the server returns an error without an http code`() = runTest {
        every { contentResolver.openInputStream(photoUri) } answers { photoBytes.inputStream() }
        server.enqueue(MockResponse().setSocketPolicy(SocketPolicy.DISCONNECT_AFTER_REQUEST))

        val result = repository.uploadFile(selectedPhoto(), "42") as AppResult.Error

        assertNull(result.code)
    }

    @Test
    fun `describeSelection reads display name and size from the openable columns`() = runTest {
        val cursor = mockk<Cursor>(relaxed = true)
        every { contentResolver.query(photoUri, any(), null, null, null) } returns cursor
        every { cursor.moveToFirst() } returns true
        every { cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME) } returns 0
        every { cursor.getColumnIndex(OpenableColumns.SIZE) } returns 1
        every { cursor.getString(0) } returns "IMG_0001.jpg"
        every { cursor.isNull(1) } returns false
        every { cursor.getLong(1) } returns 2048L

        val selection = repository.describeSelection(listOf(photoUri))

        assertEquals(listOf(SelectedFile(photoUri, "IMG_0001.jpg", 2048L)), selection)
    }

    @Test
    fun `describeSelection keeps the size unknown when the provider does not report it`() = runTest {
        val cursor = mockk<Cursor>(relaxed = true)
        every { contentResolver.query(photoUri, any(), null, null, null) } returns cursor
        every { cursor.moveToFirst() } returns true
        every { cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME) } returns 0
        every { cursor.getColumnIndex(OpenableColumns.SIZE) } returns 1
        every { cursor.getString(0) } returns "IMG_0001.jpg"
        every { cursor.isNull(1) } returns true

        val selection = repository.describeSelection(listOf(photoUri))

        assertEquals(UNKNOWN_SIZE_BYTES, selection.single().sizeBytes)
    }

    private class CountingInputStream(content: ByteArray) : InputStream() {
        private val delegate = content.inputStream()
        var readCalls = 0
        var largestRead = 0
        var isClosed = false

        override fun read(): Int = delegate.read()

        override fun read(buffer: ByteArray, offset: Int, length: Int): Int {
            val bytesRead = delegate.read(buffer, offset, length)
            if (bytesRead > 0) {
                readCalls++
                largestRead = maxOf(largestRead, bytesRead)
            }
            return bytesRead
        }

        override fun close() {
            isClosed = true
        }
    }
}
