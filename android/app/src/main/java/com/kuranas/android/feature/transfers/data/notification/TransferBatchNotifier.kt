package com.kuranas.android.feature.transfers.data.notification

import android.Manifest
import android.annotation.SuppressLint
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.content.pm.ServiceInfo
import android.net.Uri
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.core.content.ContextCompat
import androidx.core.net.toUri
import androidx.work.ForegroundInfo
import com.kuranas.android.R
import com.kuranas.android.feature.files.data.download.UNKNOWN_MIME_TYPE
import com.kuranas.android.feature.transfers.data.TransferDao
import com.kuranas.android.feature.transfers.data.TransferDirection
import com.kuranas.android.feature.transfers.ui.TransfersIntents
import dagger.hilt.android.qualifiers.ApplicationContext
import java.util.concurrent.ConcurrentHashMap
import javax.inject.Inject
import javax.inject.Singleton

interface TransferBatchNotifier {
    suspend fun progressForegroundInfo(batchId: String): ForegroundInfo
    suspend fun refresh(batchId: String)
}

@Singleton
class SystemTransferBatchNotifier @Inject constructor(
    @ApplicationContext private val context: Context,
    private val transferDao: TransferDao,
) : TransferBatchNotifier {

    private val notificationManager = NotificationManagerCompat.from(context)
    private val lastProgressSignatureByBatchId = ConcurrentHashMap<String, ProgressSignature>()

    init {
        createChannel()
    }

    override suspend fun progressForegroundInfo(batchId: String): ForegroundInfo {
        val summary = summarizeBatch(transferDao.findByBatch(batchId))
        val notification = progressNotification(batchId, summary)
        if (summary != null) lastProgressSignatureByBatchId[batchId] = ProgressSignature.of(summary)
        return ForegroundInfo(
            TransferBatchNotificationIds.progress(batchId),
            notification,
            ServiceInfo.FOREGROUND_SERVICE_TYPE_DATA_SYNC,
        )
    }

    override suspend fun refresh(batchId: String) {
        val summary = summarizeBatch(transferDao.findByBatch(batchId))
        when (val plan = planBatchNotification(summary)) {
            is TransferBatchNotificationPlan.Progress -> showProgress(batchId, plan.summary)
            is TransferBatchNotificationPlan.Result -> showResult(batchId, plan.summary)
            TransferBatchNotificationPlan.Nothing -> clear(batchId)
        }
    }

    private fun showProgress(batchId: String, summary: TransferBatchSummary) {
        val signature = ProgressSignature.of(summary)
        if (lastProgressSignatureByBatchId.put(batchId, signature) == signature) return
        notificationManager.cancel(TransferBatchNotificationIds.result(batchId))
        notify(TransferBatchNotificationIds.progress(batchId), progressNotification(batchId, summary))
    }

    private fun showResult(batchId: String, summary: TransferBatchSummary) {
        lastProgressSignatureByBatchId.remove(batchId)
        notificationManager.cancel(TransferBatchNotificationIds.progress(batchId))
        notify(TransferBatchNotificationIds.result(batchId), resultNotification(batchId, summary))
    }

    private fun clear(batchId: String) {
        lastProgressSignatureByBatchId.remove(batchId)
        notificationManager.cancel(TransferBatchNotificationIds.progress(batchId))
        notificationManager.cancel(TransferBatchNotificationIds.result(batchId))
    }

    private fun progressNotification(batchId: String, summary: TransferBatchSummary?): Notification {
        val direction = summary?.direction ?: TransferDirection.DOWNLOAD
        val totalCount = summary?.totalCount ?: 0
        val titleId = if (direction == TransferDirection.UPLOAD) R.plurals.transfer_batch_uploading_title else R.plurals.transfer_batch_downloading_title
        val iconId = if (direction == TransferDirection.UPLOAD) android.R.drawable.stat_sys_upload else android.R.drawable.stat_sys_download
        return NotificationCompat.Builder(context, CHANNEL_ID)
            .setSmallIcon(iconId)
            .setContentTitle(context.resources.getQuantityString(titleId, totalCount, totalCount))
            .setContentText(summary?.let { progressText(it) })
            .setContentIntent(openTransfersPendingIntent(TransferBatchNotificationIds.progress(batchId)))
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setOngoing(true)
            .setOnlyAlertOnce(true)
            .setProgress(PERCENT_SCALE, summary?.percent ?: 0, summary == null)
            .build()
    }

    private fun progressText(summary: TransferBatchSummary): String =
        if (summary.isWaitingForConnection) {
            context.getString(R.string.transfers_status_waiting_for_network)
        } else {
            context.getString(R.string.transfer_batch_progress_text, summary.finishedCount, summary.totalCount, summary.percent)
        }

    private fun resultNotification(batchId: String, summary: TransferBatchSummary): Notification {
        val notificationId = TransferBatchNotificationIds.result(batchId)
        val builder = NotificationCompat.Builder(context, CHANNEL_ID)
            .setPriority(NotificationCompat.PRIORITY_LOW)
            .setAutoCancel(true)
        val savedDownload = summary.singleSavedDownload
        val savedUri = savedDownload?.destinationUri
        if (savedDownload == null || savedUri == null) {
            val iconId = if (summary.direction == TransferDirection.UPLOAD) android.R.drawable.stat_sys_upload_done else android.R.drawable.stat_sys_download_done
            return builder
                .setSmallIcon(iconId)
                .setContentTitle(context.getString(R.string.transfer_batch_result_title))
                .setContentText(resultText(summary))
                .setContentIntent(openTransfersPendingIntent(notificationId))
                .build()
        }
        val openFile = openFilePendingIntent(notificationId, savedUri.toUri(), savedDownload.mimeType ?: UNKNOWN_MIME_TYPE)
        return builder
            .setSmallIcon(android.R.drawable.stat_sys_download_done)
            .setContentTitle(context.getString(R.string.download_notification_completed_title))
            .setContentText(savedDownload.displayName)
            .setContentIntent(openFile)
            .addAction(0, context.getString(R.string.download_open_action), openFile)
            .build()
    }

    private fun resultText(summary: TransferBatchSummary): String =
        batchResultParts(summary).joinToString(PART_SEPARATOR) { part ->
            val textId = resultPartTextId(summary.direction, part.kind)
            context.resources.getQuantityString(textId, part.count, part.count)
        }

    private fun openTransfersPendingIntent(requestCode: Int): PendingIntent =
        PendingIntent.getActivity(
            context,
            requestCode,
            TransfersIntents.openTransfers(context),
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT,
        )

    private fun openFilePendingIntent(requestCode: Int, uri: Uri, mimeType: String): PendingIntent {
        val viewIntent = Intent(Intent.ACTION_VIEW)
            .setDataAndType(uri, mimeType)
            .addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION or Intent.FLAG_ACTIVITY_NEW_TASK)
        return PendingIntent.getActivity(context, requestCode, viewIntent, PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT)
    }

    @SuppressLint("MissingPermission")
    private fun notify(notificationId: Int, notification: Notification) {
        val hasPermission = ContextCompat.checkSelfPermission(context, Manifest.permission.POST_NOTIFICATIONS) ==
            PackageManager.PERMISSION_GRANTED
        if (hasPermission) notificationManager.notify(notificationId, notification)
    }

    private fun createChannel() {
        val channel = NotificationChannel(
            CHANNEL_ID,
            context.getString(R.string.transfer_channel_name),
            NotificationManager.IMPORTANCE_LOW,
        ).apply { description = context.getString(R.string.transfer_channel_description) }
        notificationManager.createNotificationChannel(channel)
    }

    private data class ProgressSignature(
        val finishedCount: Int,
        val totalCount: Int,
        val percent: Int,
        val isWaitingForConnection: Boolean,
    ) {
        companion object {
            fun of(summary: TransferBatchSummary) = ProgressSignature(
                summary.finishedCount,
                summary.totalCount,
                summary.percent,
                summary.isWaitingForConnection,
            )
        }
    }

    private companion object {
        const val CHANNEL_ID = "transfers"
        const val PERCENT_SCALE = 100
        const val PART_SEPARATOR = " · "
    }
}

fun resultPartTextId(direction: TransferDirection, kind: BatchResultPartKind): Int = when (kind) {
    BatchResultPartKind.TRANSFERRED ->
        if (direction == TransferDirection.UPLOAD) R.plurals.transfer_batch_result_uploaded else R.plurals.transfer_batch_result_downloaded
    BatchResultPartKind.ALREADY_EXISTED -> R.plurals.transfer_batch_result_already_existed
    BatchResultPartKind.FAILED -> R.plurals.transfer_batch_result_failed
}
