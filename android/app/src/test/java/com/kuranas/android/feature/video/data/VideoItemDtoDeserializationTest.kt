package com.kuranas.android.feature.video.data

import kotlinx.serialization.json.Json
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class VideoItemDtoDeserializationTest {

    private val json = Json { ignoreUnknownKeys = true }

    @Test
    fun `video with starred true is decoded as starred`() {
        val video = json.decodeFromString<VideoItemDto>("""{"id":5,"name":"a.mp4","starred":true}""")

        assertTrue(video.starred)
    }

    @Test
    fun `video without starred defaults to not starred`() {
        val video = json.decodeFromString<VideoItemDto>("""{"id":5,"name":"a.mp4"}""")

        assertFalse(video.starred)
    }
}
