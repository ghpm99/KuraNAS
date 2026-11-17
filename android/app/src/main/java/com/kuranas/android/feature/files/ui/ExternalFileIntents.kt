package com.kuranas.android.feature.files.ui

import android.content.Context
import android.content.Intent
import com.kuranas.android.R
import com.kuranas.android.feature.files.data.download.DownloadPurpose

fun buildExternalFileIntent(context: Context, delivery: ExternalDelivery.Ready): Intent {
    val actionIntent = when (delivery.purpose) {
        DownloadPurpose.SHARE -> Intent(Intent.ACTION_SEND)
            .setType(delivery.mimeType)
            .putExtra(Intent.EXTRA_STREAM, delivery.uri)
        else -> Intent(Intent.ACTION_VIEW).setDataAndType(delivery.uri, delivery.mimeType)
    }
    actionIntent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
    val chooserTitle = context.getString(
        if (delivery.purpose == DownloadPurpose.SHARE) R.string.file_share else R.string.file_open_with,
    )
    return Intent.createChooser(actionIntent, chooserTitle).addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
}
