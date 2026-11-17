package com.kuranas.android.feature.files.data.download

import android.content.Context
import com.kuranas.android.R

/** Mensagem do servidor como veio; falhas do próprio app usam o catálogo de strings. */
fun DownloadFailure.toMessage(context: Context): String = when (this) {
    is DownloadFailure.Server -> message ?: context.getString(R.string.download_failed_http, httpCode)
    DownloadFailure.ConnectionLost -> context.getString(R.string.download_failed_connection)
    DownloadFailure.StorageUnavailable -> context.getString(R.string.download_failed_storage)
}
