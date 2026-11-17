package com.kuranas.android.feature.files.data

const val INLINE_PREVIEW_LIMIT_MEGABYTES = 20
const val INLINE_PREVIEW_LIMIT_BYTES = INLINE_PREVIEW_LIMIT_MEGABYTES * 1024L * 1024L

fun canPreviewInline(sizeBytes: Long): Boolean = sizeBytes <= INLINE_PREVIEW_LIMIT_BYTES
