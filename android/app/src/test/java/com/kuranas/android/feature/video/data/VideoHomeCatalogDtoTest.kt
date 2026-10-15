package com.kuranas.android.feature.video.data

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class VideoHomeCatalogDtoTest {

    private fun sectionOf(key: String, vararg videoIds: Int) = VideoSectionDto(
        key = key,
        items = videoIds.map { VideoCatalogItemDto(video = VideoItemDto(rawId = it)) },
    )

    @Test
    fun recentVideosContainsOnlyRecentSectionWithoutDuplicates() {
        val catalog = VideoHomeCatalogDto(
            sections = listOf(
                sectionOf("continue", 1, 2),
                sectionOf("series", 1, 3),
                sectionOf("recent", 1, 4, 1),
            ),
        )

        assertEquals(listOf("1", "4"), catalog.recentVideos.map { it.id })
    }

    @Test
    fun recentVideosIsEmptyWithoutRecentSection() {
        val catalog = VideoHomeCatalogDto(sections = listOf(sectionOf("series", 1), sectionOf("continue", 2)))

        assertTrue(catalog.recentVideos.isEmpty())
    }
}
