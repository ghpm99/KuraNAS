package com.kuranas.android.feature.images.data

import kotlinx.serialization.json.Json
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class ImageLibraryItemDtoTest {

    private val json = Json { ignoreUnknownKeys = true }

    @Test
    fun `decodes starred when present`() {
        val item = json.decodeFromString<ImageLibraryItemDto>("""{"file_id":4,"name":"a.jpg","starred":true}""")

        assertEquals("4", item.id)
        assertTrue(item.starred)
    }

    @Test
    fun `defaults starred to false when absent`() {
        val item = json.decodeFromString<ImageLibraryItemDto>("""{"file_id":4,"name":"a.jpg"}""")

        assertFalse(item.starred)
    }
}
