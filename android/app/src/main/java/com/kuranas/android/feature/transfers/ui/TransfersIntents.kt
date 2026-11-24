package com.kuranas.android.feature.transfers.ui

import android.content.Context
import android.content.Intent
import com.kuranas.android.MainActivity

object TransfersIntents {
    const val ACTION_OPEN_TRANSFERS = "com.kuranas.android.action.OPEN_TRANSFERS"

    fun openTransfers(context: Context): Intent =
        Intent(context, MainActivity::class.java)
            .setAction(ACTION_OPEN_TRANSFERS)
            .addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP)

    fun requestsOpenTransfers(intent: Intent?): Boolean = intent?.action == ACTION_OPEN_TRANSFERS
}
