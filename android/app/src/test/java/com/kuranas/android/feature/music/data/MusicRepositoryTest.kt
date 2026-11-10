package com.kuranas.android.feature.music.data

import com.kuranas.android.core.network.PageDto
import com.kuranas.android.core.server.ServerStore
import com.kuranas.android.feature.files.data.FileStarStore
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Test
import java.io.IOException

class MusicRepositoryTest {

    private val api = mockk<MusicApi>()
    private val starStore = FileStarStore()
    private val repository = MusicRepository(api, mockk<ServerStore>(), starStore)

    @Test
    fun `getTrackById returns the track fetched directly from the server`() = runTest {
        val serverTrack = TrackDto(id = 42, name = "song.mp3")
        coEvery { api.getTrackById(42) } returns serverTrack

        assertEquals(serverTrack, repository.getTrackById(42))
        coVerify(exactly = 0) { api.getAllTracks(any(), any()) }
    }

    @Test
    fun `getTrackById falls back to a bare track on network failure`() = runTest {
        coEvery { api.getTrackById(42) } throws IOException("offline")

        assertEquals(TrackDto(id = 42), repository.getTrackById(42))
    }

    @Test
    fun `loaded track page replaces a stale star override with the server value`() = runTest {
        starStore.set("42", true)
        coEvery { api.getAllTracks(1, any()) } returns PageDto(items = listOf(TrackDto(id = 42, starred = false)))

        repository.getAllTracks(1)

        assertEquals(false, starStore.starredById.value["42"])
    }

    @Test
    fun `failed track page keeps the recorded stars`() = runTest {
        starStore.set("42", true)
        coEvery { api.getAllTracks(1, any()) } throws IOException("offline")

        repository.getAllTracks(1)

        assertEquals(true, starStore.starredById.value["42"])
    }
}
