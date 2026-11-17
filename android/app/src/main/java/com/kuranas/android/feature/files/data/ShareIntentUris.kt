package com.kuranas.android.feature.files.data

import android.content.Intent
import android.net.Uri

/** Uris compartilhados: um em ACTION_SEND, a lista em ACTION_SEND_MULTIPLE, nenhum para qualquer outra ação. */
fun sharedUrisFrom(action: String?, singleStream: Uri?, multipleStreams: List<Uri>?): List<Uri> = when (action) {
    Intent.ACTION_SEND -> listOfNotNull(singleStream)
    Intent.ACTION_SEND_MULTIPLE -> multipleStreams.orEmpty()
    else -> emptyList()
}

fun Intent.sharedUris(): List<Uri> = sharedUrisFrom(
    action = action,
    singleStream = getParcelableExtra(Intent.EXTRA_STREAM, Uri::class.java),
    multipleStreams = getParcelableArrayListExtra(Intent.EXTRA_STREAM, Uri::class.java),
)
