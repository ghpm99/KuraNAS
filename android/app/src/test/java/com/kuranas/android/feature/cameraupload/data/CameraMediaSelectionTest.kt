package com.kuranas.android.feature.cameraupload.data

import org.junit.Assert.assertEquals
import org.junit.Test
import java.time.ZoneId
import java.time.ZoneOffset

class CameraMediaSelectionTest {

    private fun media(id: Long, kind: CameraMediaKind = CameraMediaKind.IMAGE, takenMillis: Long = id * 1000) = CameraMedia(
        id = id,
        kind = kind,
        contentUri = "content://media/$id",
        displayName = "IMG_$id.jpg",
        sizeBytes = 10,
        dateTakenMillis = takenMillis,
        dateModifiedMillis = takenMillis,
    )

    @Test
    fun skipsMediaAlreadyRegistered() {
        val selected = selectMediaToUpload(
            scanned = listOf(media(1), media(2)),
            alreadyRegisteredKeys = setOf(CameraMediaKey(1, CameraMediaKind.IMAGE)),
            includeVideos = true,
        )
        assertEquals(listOf(2L), selected.map { it.id })
    }

    @Test
    fun sameIdOfDifferentKindIsDifferentMedia() {
        val selected = selectMediaToUpload(
            scanned = listOf(media(1, CameraMediaKind.VIDEO)),
            alreadyRegisteredKeys = setOf(CameraMediaKey(1, CameraMediaKind.IMAGE)),
            includeVideos = true,
        )
        assertEquals(1, selected.size)
    }

    @Test
    fun excludesVideosWhenDisabled() {
        val selected = selectMediaToUpload(
            scanned = listOf(media(1), media(2, CameraMediaKind.VIDEO)),
            alreadyRegisteredKeys = emptySet(),
            includeVideos = false,
        )
        assertEquals(listOf(1L), selected.map { it.id })
    }

    @Test
    fun ordersOldestFirst() {
        val selected = selectMediaToUpload(listOf(media(1, takenMillis = 500), media(2, takenMillis = 100)), emptySet(), true)
        assertEquals(listOf(2L, 1L), selected.map { it.id })
    }

    @Test
    fun buildsPrefixYearMonth() {
        val novemberFirst = java.time.LocalDateTime.of(2026, 11, 1, 0, 0).toInstant(ZoneOffset.UTC).toEpochMilli()
        assertEquals("Fotos/Celular/2026/11", relativePathFor("Fotos/Celular", novemberFirst, ZoneOffset.UTC))
    }

    @Test
    fun monthBoundaryDependsOnZone() {
        val lastInstantOfOctoberUtc = java.time.LocalDateTime.of(2026, 10, 31, 23, 30).toInstant(ZoneOffset.UTC).toEpochMilli()
        assertEquals("2026/10", relativePathFor("", lastInstantOfOctoberUtc, ZoneOffset.UTC))
        assertEquals("2026/11", relativePathFor("", lastInstantOfOctoberUtc, ZoneId.of("Asia/Tokyo")))
        assertEquals("2026/10", relativePathFor("", lastInstantOfOctoberUtc, ZoneId.of("America/Sao_Paulo")))
    }

    @Test
    fun blankPrefixAndSlashesAreTrimmed() {
        val moment = java.time.LocalDateTime.of(2026, 1, 15, 12, 0).toInstant(ZoneOffset.UTC).toEpochMilli()
        assertEquals("2026/01", relativePathFor("  ", moment, ZoneOffset.UTC))
        assertEquals("A/B/2026/01", relativePathFor("/A/B/", moment, ZoneOffset.UTC))
    }
}
