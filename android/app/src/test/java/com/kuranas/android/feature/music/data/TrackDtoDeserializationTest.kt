package com.kuranas.android.feature.music.data

import kotlinx.serialization.json.Json
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TrackDtoDeserializationTest {

    private val json = Json { ignoreUnknownKeys = true }

    @Test
    fun `track with starred true is decoded as starred`() {
        val track = json.decodeFromString<TrackDto>("""{"id":5,"name":"a.mp3","starred":true}""")

        assertTrue(track.starred)
    }

    @Test
    fun `track without starred defaults to not starred`() {
        val track = json.decodeFromString<TrackDto>("""{"id":5,"name":"a.mp3"}""")

        assertFalse(track.starred)
    }
}
