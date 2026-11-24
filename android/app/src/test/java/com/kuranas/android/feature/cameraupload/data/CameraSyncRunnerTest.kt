package com.kuranas.android.feature.cameraupload.data

import com.kuranas.android.feature.files.data.UploadDestination
import com.kuranas.android.feature.transfers.data.TransferClock
import com.kuranas.android.feature.transfers.data.upload.EnqueuedUploads
import com.kuranas.android.feature.transfers.data.upload.TransferRepository
import com.kuranas.android.feature.transfers.data.upload.UploadRequestItem
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.mockk
import io.mockk.slot
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test
import java.time.ZoneOffset

class CameraSyncRunnerTest {

    private val settingsStore = mockk<CameraUploadSettingsStore>(relaxed = true)
    private val reachability = mockk<ServerReachability>()
    private val scanner = mockk<CameraMediaScanner>()
    private val uploadDao = mockk<CameraUploadDao>(relaxed = true)
    private val transferRepository = mockk<TransferRepository>()
    private val runner = CameraSyncRunner(settingsStore, reachability, scanner, uploadDao, transferRepository, TransferClock { 5_000L })

    private val novemberMillis = java.time.LocalDateTime.of(2026, 11, 10, 12, 0).toInstant(ZoneOffset.UTC).toEpochMilli()
    private val photosFolder = UploadDestination("7", "Fotos")
    private val enabledSettings = CameraUploadSettings(isEnabled = true, destination = photosFolder, onlyOnWifi = true)

    private fun media(id: Long) = CameraMedia(id, CameraMediaKind.IMAGE, "content://media/$id", "IMG_$id.jpg", 10, novemberMillis, novemberMillis)

    @Before
    fun setUp() {
        coEvery { settingsStore.current() } returns enabledSettings
        coEvery { reachability.isHomeServerReachable() } returns true
        coEvery { uploadDao.registeredKeys() } returns emptyList()
    }

    @Test
    fun skipsWhenDisabled() = runTest {
        coEvery { settingsStore.current() } returns enabledSettings.copy(isEnabled = false)
        assertEquals(CameraSyncOutcome.DISABLED, runner.run(ZoneOffset.UTC))
        coVerify(exactly = 0) { reachability.isHomeServerReachable() }
        coVerify(exactly = 0) { settingsStore.recordSync(any(), any()) }
    }

    @Test
    fun awayFromHomeDoesNotScanOrEnqueue() = runTest {
        coEvery { reachability.isHomeServerReachable() } returns false
        assertEquals(CameraSyncOutcome.AWAY_FROM_HOME, runner.run(ZoneOffset.UTC))
        coVerify(exactly = 0) { scanner.scan(any(), any()) }
        coVerify(exactly = 0) { transferRepository.enqueueUploadBatch(any(), any(), any()) }
        coVerify { settingsStore.recordSync(null, CameraSyncOutcome.AWAY_FROM_HOME) }
    }

    @Test
    fun enqueuesOneBatchAndRegistersMedia() = runTest {
        coEvery { scanner.scan(true, 0) } returns listOf(media(1), media(2))
        val items = slot<List<UploadRequestItem>>()
        coEvery { transferRepository.enqueueUploadBatch(capture(items), photosFolder, true) } returns EnqueuedUploads("batch", listOf(11L, 12L))
        val registrations = slot<List<CameraUploadEntity>>()
        coEvery { uploadDao.registerAll(capture(registrations)) } returns Unit

        assertEquals(CameraSyncOutcome.ENQUEUED, runner.run(ZoneOffset.UTC))

        assertEquals(listOf("Fotos/Celular/2026/11", "Fotos/Celular/2026/11"), items.captured.map { it.relativePath })
        assertEquals(listOf("content://media/1", "content://media/2"), items.captured.map { it.sourceUri })
        assertEquals(listOf(11L, 12L), registrations.captured.map { it.transferId })
        assertEquals(listOf(1L, 2L), registrations.captured.map { it.mediaStoreId })
        coVerify(exactly = 1) { transferRepository.enqueueUploadBatch(any(), any(), any()) }
        coVerify { settingsStore.recordSync(5_000L, CameraSyncOutcome.ENQUEUED) }
    }

    @Test
    fun alreadyRegisteredMediaIsSkipped() = runTest {
        coEvery { scanner.scan(true, 0) } returns listOf(media(1), media(2))
        coEvery { uploadDao.registeredKeys() } returns listOf(CameraMediaKey(1, CameraMediaKind.IMAGE))
        val items = slot<List<UploadRequestItem>>()
        coEvery { transferRepository.enqueueUploadBatch(capture(items), any(), any()) } returns EnqueuedUploads("batch", listOf(20L))

        runner.run(ZoneOffset.UTC)

        assertEquals(listOf("IMG_2.jpg"), items.captured.map { it.displayName })
    }

    @Test
    fun emptySelectionDoesNotEnqueue() = runTest {
        coEvery { scanner.scan(true, 0) } returns listOf(media(1))
        coEvery { uploadDao.registeredKeys() } returns listOf(CameraMediaKey(1, CameraMediaKind.IMAGE))

        assertEquals(CameraSyncOutcome.UP_TO_DATE, runner.run(ZoneOffset.UTC))
        coVerify(exactly = 0) { transferRepository.enqueueUploadBatch(any(), any(), any()) }
    }

    @Test
    fun missingMediaPermissionIsReported() = runTest {
        coEvery { scanner.scan(any(), any()) } throws SecurityException()
        assertEquals(CameraSyncOutcome.MEDIA_PERMISSION_MISSING, runner.run(ZoneOffset.UTC))
    }
}
