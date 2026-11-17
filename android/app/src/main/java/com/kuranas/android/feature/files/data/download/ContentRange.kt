package com.kuranas.android.feature.files.data.download

data class ContentRange(val firstByte: Long, val totalBytes: Long)

private val CONTENT_RANGE_PATTERN = Regex("""bytes (\d+)-\d+/(\d+)""")

fun parseContentRange(headerValue: String?): ContentRange? {
    val match = CONTENT_RANGE_PATTERN.matchEntire(headerValue?.trim().orEmpty()) ?: return null
    val (firstByte, totalBytes) = match.destructured
    return ContentRange(firstByte.toLong(), totalBytes.toLong())
}
