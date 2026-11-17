package com.kuranas.android.feature.files.data.download

import com.kuranas.android.feature.files.data.FileItemDto
import org.junit.Assert.assertEquals
import org.junit.Test

class DownloadCollectionTest {

    @Test
    fun `images go to the pictures collection`() {
        assertEquals(DownloadCollection.IMAGES, downloadCollectionFor("image/jpeg"))
        assertEquals("Pictures/KuraNAS", DownloadCollection.IMAGES.relativeDirectory)
    }

    @Test
    fun `videos go to the movies collection`() {
        assertEquals(DownloadCollection.VIDEOS, downloadCollectionFor("video/x-matroska"))
        assertEquals("Movies/KuraNAS", DownloadCollection.VIDEOS.relativeDirectory)
    }

    @Test
    fun `documents and unknown types go to downloads`() {
        assertEquals(DownloadCollection.DOWNLOADS, downloadCollectionFor("application/pdf"))
        assertEquals(DownloadCollection.DOWNLOADS, downloadCollectionFor("audio/mpeg"))
        assertEquals(DownloadCollection.DOWNLOADS, downloadCollectionFor("application/vnd.openxmlformats-officedocument.wordprocessingml.document"))
        assertEquals(DownloadCollection.DOWNLOADS, downloadCollectionFor("application/octet-stream"))
    }

    @Test
    fun `the download request keeps the listed name with accents and spaces`() {
        val file = FileItemDto(rawId = 9, name = "Relatório São João (final).pdf", format = ".pdf", size = 1234)

        val request = file.toDownloadRequest()

        assertEquals(DownloadRequest("9", "Relatório São João (final).pdf", 1234), request)
    }

    @Test
    fun `mime type prefers the extension, then the response content type, then octet-stream`() {
        val resolver = MimeTypeResolver { name -> if (name.endsWith(".pdf")) "application/pdf" else null }

        assertEquals("application/pdf", resolveDownloadMimeType("a.pdf", "text/plain", resolver))
        assertEquals("image/jpeg", resolveDownloadMimeType("sem-extensao", "image/jpeg; charset=x", resolver))
        assertEquals(UNKNOWN_MIME_TYPE, resolveDownloadMimeType("sem-extensao", null, resolver))
        assertEquals(UNKNOWN_MIME_TYPE, resolveDownloadMimeType("sem-extensao", " ", resolver))
    }

    @Test
    fun `content range header is parsed into first byte and total`() {
        assertEquals(ContentRange(100, 16631), parseContentRange("bytes 100-199/16631"))
        assertEquals(null, parseContentRange("bytes */16631"))
        assertEquals(null, parseContentRange(null))
    }
}
