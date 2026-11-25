package com.kuranas.android.feature.ingest.ui

import android.content.Context
import android.content.Intent
import com.kuranas.android.MainActivity
import com.kuranas.android.feature.ingest.data.firstHttpUrlIn

object IngestIntents {
    const val ACTION_OPEN_INGEST = "com.kuranas.android.action.OPEN_INGEST"
    private const val EXTRA_SHARED_URL = "com.kuranas.android.extra.SHARED_URL"

    fun openIngest(context: Context, url: String): Intent =
        Intent(context, MainActivity::class.java)
            .setAction(ACTION_OPEN_INGEST)
            .putExtra(EXTRA_SHARED_URL, url)
            .addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP)

    fun requestedIngestUrl(intent: Intent?): String? =
        if (intent?.action == ACTION_OPEN_INGEST) intent.getStringExtra(EXTRA_SHARED_URL) else null

    fun sharedTextUrl(intent: Intent?): String? {
        if (intent?.action != Intent.ACTION_SEND || intent.type?.startsWith("text/plain") != true) return null
        return firstHttpUrlIn(intent.getStringExtra(Intent.EXTRA_TEXT))
    }
}
