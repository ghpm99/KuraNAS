package com.kuranas.android.feature.files.data

import android.content.ContentResolver
import android.content.Context
import android.database.Cursor
import android.net.Uri
import android.provider.OpenableColumns
import io.mockk.every
import io.mockk.mockk
import org.junit.Assert.assertEquals
import org.junit.Test

class SelectedFileReaderTest {

    private val contentResolver = mockk<ContentResolver>()
    private val context = mockk<Context>()
    private val photoUri = mockk<Uri>()

    private fun cursorReporting(sizeBytes: Long?): Cursor {
        val cursor = mockk<Cursor>(relaxed = true)
        every { contentResolver.query(photoUri, any(), null, null, null) } returns cursor
        every { cursor.moveToFirst() } returns true
        every { cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME) } returns 0
        every { cursor.getColumnIndex(OpenableColumns.SIZE) } returns 1
        every { cursor.getString(0) } returns "IMG_0001.jpg"
        every { cursor.isNull(1) } returns (sizeBytes == null)
        if (sizeBytes != null) every { cursor.getLong(1) } returns sizeBytes
        return cursor
    }

    @Test
    fun `reads display name and size from the openable columns`() {
        cursorReporting(sizeBytes = 2048L)

        val selectedFile = describeSelectedFile(context, contentResolver, photoUri)

        assertEquals(SelectedFile(photoUri, "IMG_0001.jpg", 2048L), selectedFile)
    }

    @Test
    fun `keeps the size unknown when the provider does not report it`() {
        cursorReporting(sizeBytes = null)

        val selectedFile = describeSelectedFile(context, contentResolver, photoUri)

        assertEquals(UNKNOWN_SIZE_BYTES, selectedFile.sizeBytes)
    }
}
