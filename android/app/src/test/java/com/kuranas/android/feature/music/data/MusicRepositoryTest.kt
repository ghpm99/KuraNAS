package com.kuranas.android.feature.music.data

import com.kuranas.android.core.server.ServerStore
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Test
import java.io.IOException

class MusicRepositoryTest {

    private val api = mockk<MusicApi>()
    private val repository = MusicRepository(api, mockk<ServerStore>())

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
}
