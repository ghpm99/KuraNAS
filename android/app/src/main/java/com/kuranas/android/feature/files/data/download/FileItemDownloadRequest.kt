package com.kuranas.android.feature.files.data.download

import com.kuranas.android.feature.files.data.FileItemDto

fun FileItemDto.toDownloadRequest(purpose: DownloadPurpose = DownloadPurpose.SAVE_TO_DEVICE) = DownloadRequest(
    fileId = id,
    fileName = name,
    sizeBytes = size,
    purpose = purpose,
)
