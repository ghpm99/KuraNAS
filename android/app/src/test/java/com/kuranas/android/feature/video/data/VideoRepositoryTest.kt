package com.kuranas.android.feature.video.data

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.server.ServerStore
import io.mockk.mockk
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import retrofit2.Retrofit

class VideoRepositoryTest {

    private val server = MockWebServer()
    private lateinit var repository: VideoRepository

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
            .create(VideoApi::class.java)
        repository = VideoRepository(api, mockk<ServerStore>())
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    @Test
    fun `getPlaylistsBySection requests the section path with page and page size`() = runTest {
        server.enqueue(MockResponse().setBody("""{"items":[],"pagination":{"page":1,"page_size":24}}"""))

        repository.getPlaylistsBySection(VideoPlaylistSection.Movies)

        val request = server.takeRequest()
        assertEquals("/api/v1/video/playlists/section/movies?page=1&page_size=24", request.path)
    }

    @Test
    fun `getPlaylistsBySection deserializes the playlists of the page`() = runTest {
        server.enqueue(
            MockResponse().setBody(
                """{"items":[{"id":7,"name":"Matrix","item_count":3,"cover_video_id":55,"type":"movie","is_auto":true}],
                |"pagination":{"page":1,"page_size":24,"has_next":false}}""".trimMargin(),
            ),
        )

        val result = repository.getPlaylistsBySection(VideoPlaylistSection.Series)

        val playlists = (result as AppResult.Success).data
        assertEquals(1, playlists.size)
        assertEquals(7, playlists[0].id)
        assertEquals("Matrix", playlists[0].name)
        assertEquals(3, playlists[0].itemCount)
        assertEquals(55, playlists[0].coverVideoId)
    }

    @Test
    fun `getPlaylistsBySection survives a payload without optional fields`() = runTest {
        server.enqueue(MockResponse().setBody("""{"items":[{"id":9}]}"""))

        val result = repository.getPlaylistsBySection(VideoPlaylistSection.Clips)

        val playlist = (result as AppResult.Success).data.single()
        assertEquals(9, playlist.id)
        assertEquals("", playlist.name)
        assertEquals(0, playlist.itemCount)
        assertNull(playlist.coverVideoId)
    }

    @Test
    fun `getPlaylistsBySection surfaces the server error message`() = runTest {
        server.enqueue(MockResponse().setResponseCode(500).setBody("falha do servidor"))

        val result = repository.getPlaylistsBySection(VideoPlaylistSection.Folders)

        assertTrue(result is AppResult.Error)
        assertEquals("falha do servidor", (result as AppResult.Error).message)
    }

    @Test
    fun `getPlaylistHeader asks for the playlist without items`() = runTest {
        server.enqueue(MockResponse().setBody("""{"id":7,"name":"Matrix","item_count":3,"items":[]}"""))

        val result = repository.getPlaylistHeader(7)

        assertEquals("/api/v1/video/playlists/7?include_items=false", server.takeRequest().path)
        assertEquals(3, (result as AppResult.Success).data.itemCount)
    }

    @Test
    fun `getPlaylistItems requests the page and deserializes has_next`() = runTest {
        server.enqueue(
            MockResponse().setBody(
                """{"items":[{"video":{"id":5,"name":"a.mp4"},"status":"in_progress","progress_pct":10.5}],""" +
                    """"pagination":{"page":2,"page_size":50,"has_next":true,"has_prev":true}}""",
            ),
        )

        val result = repository.getPlaylistItems(7, 2)

        assertEquals("/api/v1/video/playlists/7/items?page=2&page_size=50", server.takeRequest().path)
        val page = (result as AppResult.Success).data
        assertEquals(1, page.items.size)
        assertTrue(page.hasNext)
    }
}
