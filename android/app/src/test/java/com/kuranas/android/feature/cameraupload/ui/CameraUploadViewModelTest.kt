package com.kuranas.android.feature.cameraupload.ui

import com.kuranas.android.feature.transfers.data.TransferClock

import com.kuranas.android.feature.cameraupload.data.CameraSyncOutcome
import com.kuranas.android.feature.cameraupload.data.CameraUploadScheduler
import com.kuranas.android.feature.cameraupload.data.CameraUploadSettings
import com.kuranas.android.feature.cameraupload.data.CameraUploadSettingsStore
import com.kuranas.android.feature.cameraupload.data.CameraUploadStatus
import com.kuranas.android.feature.cameraupload.data.CameraUploadStatusObserver
import com.kuranas.android.feature.files.data.UploadDestination
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.coVerifyOrder
import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class CameraUploadViewModelTest {

    private val settingsFlow = MutableStateFlow(CameraUploadSettings())
    private val statusFlow = MutableStateFlow(CameraUploadStatus(0, 0, null, null))
    private val settingsStore = mockk<CameraUploadSettingsStore>(relaxed = true)
    private val scheduler = mockk<CameraUploadScheduler>(relaxed = true)
    private val statusObserver = mockk<CameraUploadStatusObserver>()
    private val nowSeconds = 1_700_000_000L
    private val clock = TransferClock { nowSeconds * 1000 }
    private val photosFolder = UploadDestination("42", "Fotos")

    @Before
    fun setUp() {
        kotlinx.coroutines.Dispatchers.setMain(UnconfinedTestDispatcher())
        every { settingsStore.settings } returns settingsFlow
        every { statusObserver.status } returns statusFlow
        coEvery { settingsStore.current() } answers { settingsFlow.value }
    }

    @After
    fun tearDown() {
        kotlinx.coroutines.Dispatchers.resetMain()
    }

    private fun newViewModel() = CameraUploadViewModel(settingsStore, scheduler, statusObserver, clock)

    @Test
    fun `starts disabled with no enabling step`() {
        val state = newViewModel().state.value

        assertFalse(state.settings.isEnabled)
        assertEquals(EnablingStep.NONE, state.enablingStep)
        assertFalse(state.isMediaPermissionMissing)
    }

    @Test
    fun `requesting to enable explains permissions before asking`() {
        val viewModel = newViewModel()

        viewModel.onEnableRequested()

        assertEquals(EnablingStep.EXPLAINING_PERMISSIONS, viewModel.state.value.enablingStep)
        coVerify(exactly = 0) { settingsStore.setEnabled(any()) }
    }

    @Test
    fun `granted media permission moves to the starting point choice`() {
        val viewModel = newViewModel()

        viewModel.onPermissionRequestFinished(isMediaGranted = true)

        assertEquals(EnablingStep.CHOOSING_STARTING_POINT, viewModel.state.value.enablingStep)
        assertFalse(viewModel.state.value.isMediaPermissionMissing)
    }

    @Test
    fun `denied media permission keeps the feature off and flags the missing permission`() {
        val viewModel = newViewModel()

        viewModel.onPermissionRequestFinished(isMediaGranted = false)

        assertEquals(EnablingStep.NONE, viewModel.state.value.enablingStep)
        assertTrue(viewModel.state.value.isMediaPermissionMissing)
        coVerify(exactly = 0) { settingsStore.setEnabled(any()) }
    }

    @Test
    fun `enabling with everything saves starting point zero then enables and applies`() {
        val viewModel = newViewModel()

        viewModel.enableSendingEverything()

        coVerifyOrder {
            settingsStore.setStartingPointSeconds(0)
            settingsStore.setEnabled(true)
            scheduler.apply(any())
        }
        assertEquals(EnablingStep.NONE, viewModel.state.value.enablingStep)
    }

    @Test
    fun `enabling from now saves the current time in seconds`() {
        newViewModel().enableSendingFromNow()

        coVerifyOrder {
            settingsStore.setStartingPointSeconds(nowSeconds)
            settingsStore.setEnabled(true)
            scheduler.apply(any())
        }
    }

    @Test
    fun `applies the scheduler with the settings read after saving`() {
        val enabledSettings = CameraUploadSettings(isEnabled = true)
        coEvery { settingsStore.current() } returns enabledSettings

        newViewModel().enableSendingEverything()

        verify { scheduler.apply(enabledSettings) }
    }

    @Test
    fun `disabling saves disabled and applies so the schedule is cancelled`() {
        newViewModel().disable()

        coVerifyOrder {
            settingsStore.setEnabled(false)
            scheduler.apply(any())
        }
    }

    @Test
    fun `toggling only on wifi saves and reapplies`() {
        newViewModel().setOnlyOnWifi(false)

        coVerifyOrder {
            settingsStore.setOnlyOnWifi(false)
            scheduler.apply(any())
        }
    }

    @Test
    fun `toggling only while charging saves and reapplies`() {
        newViewModel().setOnlyWhileCharging(true)

        coVerifyOrder {
            settingsStore.setOnlyWhileCharging(true)
            scheduler.apply(any())
        }
    }

    @Test
    fun `toggling include videos saves and reapplies`() {
        newViewModel().setIncludeVideos(false)

        coVerifyOrder {
            settingsStore.setIncludeVideos(false)
            scheduler.apply(any())
        }
    }

    @Test
    fun `picking a non root folder clears the relative prefix`() {
        newViewModel().pickDestination(photosFolder)

        coVerifyOrder {
            settingsStore.setDestination(photosFolder)
            settingsStore.setRelativePrefix("")
            scheduler.apply(any())
        }
    }

    @Test
    fun `picking the root restores the default relative prefix`() {
        newViewModel().pickDestination(UploadDestination.Root)

        coVerify { settingsStore.setRelativePrefix(CameraUploadSettings.DEFAULT_RELATIVE_PREFIX) }
    }

    @Test
    fun `sync now delegates to the scheduler`() {
        newViewModel().syncNow()

        verify { scheduler.syncNow() }
    }

    @Test
    fun `exposes the status emitted by the observer`() {
        val viewModel = newViewModel()
        val status = CameraUploadStatus(3, 1, 1234L, CameraSyncOutcome.AWAY_FROM_HOME)

        statusFlow.value = status

        assertEquals(status, viewModel.state.value.status)
    }

    @Test
    fun `flags missing permission when enabled but media access was revoked`() {
        settingsFlow.value = CameraUploadSettings(isEnabled = true)
        val viewModel = newViewModel()

        viewModel.onMediaPermissionChecked(isMediaGranted = false)

        assertTrue(viewModel.state.value.isMediaPermissionMissing)
    }

    @Test
    fun `does not flag missing permission while the feature is off`() {
        val viewModel = newViewModel()
        viewModel.onPermissionRequestFinished(isMediaGranted = false)

        viewModel.onMediaPermissionChecked(isMediaGranted = false)

        assertFalse(viewModel.state.value.isMediaPermissionMissing)
    }

    @Test
    fun `clears the missing permission flag once access is granted`() {
        settingsFlow.value = CameraUploadSettings(isEnabled = true)
        val viewModel = newViewModel()
        viewModel.onMediaPermissionChecked(isMediaGranted = false)

        viewModel.onMediaPermissionChecked(isMediaGranted = true)

        assertFalse(viewModel.state.value.isMediaPermissionMissing)
    }
}
