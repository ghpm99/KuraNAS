package com.kuranas.android.feature.transfers.data.download

import com.kuranas.android.feature.files.data.download.DownloadDestinationRestorer
import com.kuranas.android.feature.files.data.download.DownloadNotifier
import com.kuranas.android.feature.files.data.download.UNKNOWN_MIME_TYPE
import com.kuranas.android.feature.transfers.data.TransferEntity
import javax.inject.Inject

class PartialDownloadDiscarder @Inject constructor(
    private val restorer: DownloadDestinationRestorer,
    private val notifier: DownloadNotifier,
) {

    fun discard(transfer: TransferEntity) {
        notifier.dismiss(transfer.id.toInt())
        val reference = transfer.destinationUri ?: return
        restorer.discard(transfer.toDownloadRequest(), reference, transfer.mimeType ?: UNKNOWN_MIME_TYPE)
    }
}
