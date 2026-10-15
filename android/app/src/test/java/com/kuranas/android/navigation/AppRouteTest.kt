package com.kuranas.android.navigation

import org.junit.Assert.assertEquals
import org.junit.Test

class AppRouteTest {

    @Test
    fun `video player route without playlist carries only the video id`() {
        assertEquals("video-player/42", AppRoute.videoPlayer("42"))
    }

    @Test
    fun `video player route with playlist carries the playlist id`() {
        assertEquals("video-player/42?playlistId=9", AppRoute.videoPlayer("42", 9))
    }
}
