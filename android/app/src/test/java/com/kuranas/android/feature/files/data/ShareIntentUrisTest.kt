package com.kuranas.android.feature.files.data

import android.content.Intent
import android.net.Uri
import io.mockk.mockk
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class ShareIntentUrisTest {

    private val firstUri = mockk<Uri>()
    private val secondUri = mockk<Uri>()
    private val thirdUri = mockk<Uri>()

    @Test
    fun `send with a single stream yields that uri`() {
        val uris = sharedUrisFrom(Intent.ACTION_SEND, singleStream = firstUri, multipleStreams = null)

        assertEquals(listOf(firstUri), uris)
    }

    @Test
    fun `send without a stream yields nothing`() {
        assertTrue(sharedUrisFrom(Intent.ACTION_SEND, singleStream = null, multipleStreams = null).isEmpty())
    }

    @Test
    fun `send multiple yields the whole list in order`() {
        val uris = sharedUrisFrom(Intent.ACTION_SEND_MULTIPLE, singleStream = null, multipleStreams = listOf(firstUri, secondUri, thirdUri))

        assertEquals(listOf(firstUri, secondUri, thirdUri), uris)
    }

    @Test
    fun `send multiple without a list yields nothing`() {
        assertTrue(sharedUrisFrom(Intent.ACTION_SEND_MULTIPLE, singleStream = null, multipleStreams = null).isEmpty())
    }

    @Test
    fun `other actions yield nothing even when streams are present`() {
        assertTrue(sharedUrisFrom(Intent.ACTION_VIEW, singleStream = firstUri, multipleStreams = listOf(secondUri)).isEmpty())
        assertTrue(sharedUrisFrom(null, singleStream = firstUri, multipleStreams = listOf(secondUri)).isEmpty())
    }
}
