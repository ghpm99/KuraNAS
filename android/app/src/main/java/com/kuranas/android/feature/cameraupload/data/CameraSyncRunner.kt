package com.kuranas.android.feature.cameraupload.data

import com.kuranas.android.feature.transfers.data.TransferClock
import com.kuranas.android.feature.transfers.data.upload.TransferRepository
import com.kuranas.android.feature.transfers.data.upload.UploadRequestItem
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import java.time.ZoneId
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CameraSyncRunner @Inject constructor(
    private val settingsStore: CameraUploadSettingsStore,
    private val reachability: ServerReachability,
    private val scanner: CameraMediaScanner,
    private val uploadDao: CameraUploadDao,
    private val transferRepository: TransferRepository,
    private val clock: TransferClock,
) {

    private val runLock = Mutex()

    suspend fun run(zone: ZoneId = ZoneId.systemDefault()): CameraSyncOutcome = runLock.withLock {
        val settings = settingsStore.current()
        if (!settings.isEnabled) return@withLock CameraSyncOutcome.DISABLED
        if (!reachability.isHomeServerReachable()) return@withLock recordOutcome(CameraSyncOutcome.AWAY_FROM_HOME, syncedAtMillis = null)
        val scanned = try {
            scanner.scan(settings.includeVideos, settings.startingPointSeconds)
        } catch (_: SecurityException) {
            return@withLock recordOutcome(CameraSyncOutcome.MEDIA_PERMISSION_MISSING, syncedAtMillis = null)
        }
        val mediaToUpload = selectMediaToUpload(scanned, uploadDao.registeredKeys().toSet(), settings.includeVideos)
        if (mediaToUpload.isEmpty()) return@withLock recordOutcome(CameraSyncOutcome.UP_TO_DATE, clock.nowMillis())
        enqueueAndRegister(mediaToUpload, settings, zone)
        recordOutcome(CameraSyncOutcome.ENQUEUED, clock.nowMillis())
    }

    private suspend fun enqueueAndRegister(mediaToUpload: List<CameraMedia>, settings: CameraUploadSettings, zone: ZoneId) {
        val items = mediaToUpload.map { media ->
            UploadRequestItem(
                displayName = media.displayName,
                sizeBytes = media.sizeBytes,
                sourceUri = media.contentUri,
                relativePath = relativePathFor(settings.relativePrefix, media.dateTakenMillis, zone),
            )
        }
        val enqueued = transferRepository.enqueueUploadBatch(items, settings.destination, requiresUnmeteredNetwork = settings.onlyOnWifi)
        val registeredAt = clock.nowMillis()
        val registrations = mediaToUpload.zip(enqueued.transferIds) { media, transferId ->
            CameraUploadEntity(
                mediaStoreId = media.id,
                mediaKind = media.kind,
                transferId = transferId,
                displayName = media.displayName,
                dateTakenMillis = media.dateTakenMillis,
                registeredAt = registeredAt,
            )
        }
        uploadDao.registerAll(registrations)
    }

    private suspend fun recordOutcome(outcome: CameraSyncOutcome, syncedAtMillis: Long?): CameraSyncOutcome {
        settingsStore.recordSync(syncedAtMillis, outcome)
        return outcome
    }
}
