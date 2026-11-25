package com.kuranas.android.feature.ingest.data

private val activeStates = setOf("queued", "running")
private val retryableStates = setOf("failed", "canceled")

fun IngestDownloadDto.isActive(): Boolean = state in activeStates

fun IngestDownloadDto.canBeCanceled(): Boolean = isActive() && jobId != null

fun IngestDownloadDto.canBeRetried(): Boolean = state in retryableStates
