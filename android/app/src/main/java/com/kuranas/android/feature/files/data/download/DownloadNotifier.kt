package com.kuranas.android.feature.files.data.download

import android.Manifest
import android.annotation.SuppressLint
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.core.content.ContextCompat
import com.kuranas.android.R
import dagger.hilt.android.qualifiers.ApplicationContext
import java.util.concurrent.ConcurrentHashMap
import javax.inject.Inject
import javax.inject.Singleton

interface DownloadNotifier {
    fun showProgress(notificationId: Int, request: DownloadRequest, progress: DownloadProgress)
    fun showCompleted(notificationId: Int, request: DownloadRequest, uri: Uri, mimeType: String)
    fun showFailed(notificationId: Int, request: DownloadRequest, failure: DownloadFailure)
}

@Singleton
class SystemDownloadNotifier @Inject constructor(
    @ApplicationContext private val context: Context,
) : DownloadNotifier {

    private val notificationManager = NotificationManagerCompat.from(context)
    private val lastPercentByNotificationId = ConcurrentHashMap<Int, Int>()

    init {
        createChannel()
    }

    override fun showProgress(notificationId: Int, request: DownloadRequest, progress: DownloadProgress) {
        val percent = progress.fraction?.let { (it * PERCENT_SCALE).toInt() }
        if (percent != null && lastPercentByNotificationId.put(notificationId, percent) == percent) return
        val notification = baseNotification(request)
            .setSmallIcon(android.R.drawable.stat_sys_download)
            .setContentTitle(context.getString(R.string.file_downloading))
            .setOngoing(true)
            .setOnlyAlertOnce(true)
            .setProgress(PERCENT_SCALE, percent ?: 0, percent == null)
            .build()
        notify(notificationId, notification)
    }

    override fun showCompleted(notificationId: Int, request: DownloadRequest, uri: Uri, mimeType: String) {
        lastPercentByNotificationId.remove(notificationId)
        val openFile = openFilePendingIntent(notificationId, uri, mimeType)
        val notification = baseNotification(request)
            .setSmallIcon(android.R.drawable.stat_sys_download_done)
            .setContentTitle(context.getString(R.string.download_notification_completed_title))
            .setContentText(request.fileName)
            .setContentIntent(openFile)
            .addAction(0, context.getString(R.string.download_open_action), openFile)
            .setAutoCancel(true)
            .build()
        notify(notificationId, notification)
    }

    override fun showFailed(notificationId: Int, request: DownloadRequest, failure: DownloadFailure) {
        lastPercentByNotificationId.remove(notificationId)
        val notification = baseNotification(request)
            .setSmallIcon(android.R.drawable.stat_notify_error)
            .setContentTitle(context.getString(R.string.download_notification_failed_title))
            .setContentText(failure.toMessage(context))
            .setAutoCancel(true)
            .build()
        notify(notificationId, notification)
    }

    private fun baseNotification(request: DownloadRequest) =
        NotificationCompat.Builder(context, CHANNEL_ID)
            .setContentText(request.fileName)
            .setPriority(NotificationCompat.PRIORITY_LOW)

    private fun openFilePendingIntent(notificationId: Int, uri: Uri, mimeType: String): PendingIntent {
        val viewIntent = Intent(Intent.ACTION_VIEW)
            .setDataAndType(uri, mimeType)
            .addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION or Intent.FLAG_ACTIVITY_NEW_TASK)
        return PendingIntent.getActivity(
            context,
            notificationId,
            viewIntent,
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT,
        )
    }

    @SuppressLint("MissingPermission")
    private fun notify(notificationId: Int, notification: android.app.Notification) {
        val hasPermission = ContextCompat.checkSelfPermission(context, Manifest.permission.POST_NOTIFICATIONS) ==
            PackageManager.PERMISSION_GRANTED
        if (hasPermission) notificationManager.notify(notificationId, notification)
    }

    private fun createChannel() {
        val channel = NotificationChannel(
            CHANNEL_ID,
            context.getString(R.string.download_channel_name),
            NotificationManager.IMPORTANCE_LOW,
        ).apply { description = context.getString(R.string.download_channel_description) }
        notificationManager.createNotificationChannel(channel)
    }

    private companion object {
        const val CHANNEL_ID = "downloads"
        const val PERCENT_SCALE = 100
    }
}
