package com.kuranas.android.feature.video.data

import kotlinx.serialization.decodeFromString
import kotlinx.serialization.json.Json
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

    @Test
    fun continueWatchingItemsContainsOnlyContinueSectionKeepingProgress() {
        val catalog = VideoHomeCatalogDto(
            sections = listOf(
                sectionOf("recent", 9),
                VideoSectionDto(
                    key = "continue",
                    items = listOf(
                        VideoCatalogItemDto(video = VideoItemDto(rawId = 1), status = "in_progress", progressPct = 42.5),
                        VideoCatalogItemDto(video = VideoItemDto(rawId = 2), status = "in_progress", progressPct = 10.0),
                        VideoCatalogItemDto(video = VideoItemDto(rawId = 1), status = "in_progress", progressPct = 42.5),
                    ),
                ),
            ),
        )

        assertEquals(listOf("1", "2"), catalog.continueWatchingItems.map { it.video.id })
        assertEquals(listOf(42.5, 10.0), catalog.continueWatchingItems.map { it.progressPct })
    }

    @Test
    fun continueWatchingItemsIsEmptyWithoutContinueSection() {
        val catalog = VideoHomeCatalogDto(sections = listOf(sectionOf("recent", 1)))

        assertTrue(catalog.continueWatchingItems.isEmpty())
    }

    @Test
    fun catalogItemWithoutProgressPctDefaultsToZero() {
        val item = Json { ignoreUnknownKeys = true }.decodeFromString<VideoCatalogItemDto>("""{"video":{"id":3}}""")

        assertEquals(0.0, item.progressPct, 0.0)
    }
}
