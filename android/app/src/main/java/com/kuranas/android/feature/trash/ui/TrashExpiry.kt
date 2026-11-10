package com.kuranas.android.feature.trash.ui

import java.time.Duration
import java.time.Instant
import java.time.OffsetDateTime
import java.time.format.DateTimeParseException
import kotlin.math.ceil

private const val SECONDS_PER_DAY = 86_400.0

/** Whole days left until [expiresAt] (rounded up, never negative); null when the timestamp is unreadable. */
fun daysUntilExpiry(expiresAt: String, currentInstant: Instant): Int? {
    val expiryInstant = try {
        OffsetDateTime.parse(expiresAt).toInstant()
    } catch (unreadableTimestamp: DateTimeParseException) {
        return null
    }
    val remainingSeconds = Duration.between(currentInstant, expiryInstant).seconds
    return ceil(remainingSeconds / SECONDS_PER_DAY).toInt().coerceAtLeast(0)
}

/** Folder the item was deleted from; "/" for top-level items. */
fun originFolderOf(originalPath: String): String {
    val lastSeparatorIndex = originalPath.lastIndexOfAny(charArrayOf('/', '\\'))
    return if (lastSeparatorIndex <= 0) "/" else originalPath.substring(0, lastSeparatorIndex)
}
