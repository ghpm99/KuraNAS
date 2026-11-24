package com.kuranas.android.feature.cameraupload.data

import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter

private val yearMonthFormatter = DateTimeFormatter.ofPattern("yyyy/MM")

fun selectMediaToUpload(
    scanned: List<CameraMedia>,
    alreadyRegisteredKeys: Set<CameraMediaKey>,
    includeVideos: Boolean,
): List<CameraMedia> =
    scanned
        .filter { includeVideos || it.kind != CameraMediaKind.VIDEO }
        .filter { it.key !in alreadyRegisteredKeys }
        .distinctBy { it.key }
        .sortedBy { it.dateTakenMillis }

fun relativePathFor(prefix: String, dateTakenMillis: Long, zone: ZoneId): String {
    val yearMonth = yearMonthFormatter.format(Instant.ofEpochMilli(dateTakenMillis).atZone(zone))
    val cleanPrefix = prefix.trim().trim('/')
    return if (cleanPrefix.isEmpty()) yearMonth else "$cleanPrefix/$yearMonth"
}
