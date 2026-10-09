package com.kuranas.android.feature.search.data

import kotlinx.serialization.json.Json
import org.junit.Assert.assertEquals
import org.junit.Test

class SearchResultsDtoTest {

    private val json = Json { ignoreUnknownKeys = true }

    @Test
    fun `tracks become playable music and are counted once`() {
        val payload = """
            {"query":"time","files":[],"tracks":[
              {"file_id":7,"title":"Time","artist":"Pink Floyd","album":"The Dark Side","duration":413.5,"path":"/Music/time.mp3"}
            ]}
        """.trimIndent()

        val results = json.decodeFromString<SearchResultsDto>(payload)

        assertEquals(listOf("7"), results.music.map { it.id })
        assertEquals("Time", results.music.single().name)
        assertEquals("audio/mp3", results.music.single().mimeType)
        assertEquals(1, results.total)
    }

    @Test
    fun `payload without tracks and images keeps decoding`() {
        val results = json.decodeFromString<SearchResultsDto>("""{"query":"x","files":[{"id":1,"name":"a.txt","format":".txt"}]}""")

        assertEquals(1, results.files.size)
        assertEquals(0, results.music.size)
    }

    @Test
    fun `images stay listed with files`() {
        val results = json.decodeFromString<SearchResultsDto>("""{"images":[{"id":3,"name":"p.jpg","format":".jpg"}]}""")

        assertEquals(listOf("3"), results.files.map { it.id })
    }
}
