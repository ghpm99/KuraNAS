package com.kuranas.android.feature.files.data

import android.content.ContentResolver
import android.net.Uri
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import okhttp3.MediaType.Companion.toMediaType
import okio.Buffer
import org.junit.Assert.assertArrayEquals
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.FileNotFoundException
import java.io.InputStream

class UriRequestBodyTest {

    private val contentResolver = mockk<ContentResolver>()
    private val uri = mockk<Uri>()
    private val content = ByteArray(200_000) { (it % 251).toByte() }

    @Test
    fun `building the body does not open or read the content`() {
        val stream = CountingStream(content)
        every { contentResolver.openInputStream(uri) } returns stream

        UriRequestBody(contentResolver, uri, "image/jpeg".toMediaType(), content.size.toLong())

        verify(exactly = 0) { contentResolver.openInputStream(any()) }
        assertEquals(0, stream.totalBytesRead)
    }

    @Test
    fun `writeTo streams the content in chunks smaller than the whole file and closes the stream`() {
        val stream = CountingStream(content)
        every { contentResolver.openInputStream(uri) } returns stream
        val sink = Buffer()

        UriRequestBody(contentResolver, uri, null, content.size.toLong()).writeTo(sink)

        assertArrayEquals(content, sink.readByteArray())
        assertTrue(stream.readCalls > 1)
        assertTrue(stream.largestRead < content.size)
        assertTrue(stream.isClosed)
    }

    @Test
    fun `content length is the reported size and negative when unknown`() {
        assertEquals(10L, UriRequestBody(contentResolver, uri, null, 10L).contentLength())
        assertEquals(-1L, UriRequestBody(contentResolver, uri, null, UNKNOWN_SIZE_BYTES).contentLength())
    }

    @Test(expected = FileNotFoundException::class)
    fun `writeTo fails when the provider has no stream for the uri`() {
        every { contentResolver.openInputStream(uri) } returns null
        every { uri.toString() } returns "content://media/1"

        UriRequestBody(contentResolver, uri, null, 1L).writeTo(Buffer())
    }

    private class CountingStream(source: ByteArray) : InputStream() {
        private val delegate = source.inputStream()
        var readCalls = 0
        var largestRead = 0
        var totalBytesRead = 0
        var isClosed = false

        override fun read(): Int = delegate.read()

        override fun read(buffer: ByteArray, offset: Int, length: Int): Int {
            val bytesRead = delegate.read(buffer, offset, length)
            if (bytesRead > 0) {
                readCalls++
                largestRead = maxOf(largestRead, bytesRead)
                totalBytesRead += bytesRead
            }
            return bytesRead
        }

        override fun close() {
            isClosed = true
        }
    }
}
