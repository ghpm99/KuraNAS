package com.kuranas.android.feature.cameraupload.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.feature.cameraupload.data.CameraUploadScheduler
import com.kuranas.android.feature.cameraupload.data.CameraUploadSettings
import com.kuranas.android.feature.cameraupload.data.CameraUploadSettingsStore
import com.kuranas.android.feature.cameraupload.data.CameraUploadStatus
import com.kuranas.android.feature.cameraupload.data.CameraUploadStatusObserver
import com.kuranas.android.feature.files.data.UploadDestination
import com.kuranas.android.feature.transfers.data.TransferClock
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

enum class EnablingStep { NONE, EXPLAINING_PERMISSIONS, CHOOSING_STARTING_POINT }

data class CameraUploadUiState(
    val settings: CameraUploadSettings = CameraUploadSettings(),
    val status: CameraUploadStatus? = null,
    val enablingStep: EnablingStep = EnablingStep.NONE,
    val isMediaPermissionMissing: Boolean = false,
)

@HiltViewModel
class CameraUploadViewModel @Inject constructor(
    private val settingsStore: CameraUploadSettingsStore,
    private val scheduler: CameraUploadScheduler,
    statusObserver: CameraUploadStatusObserver,
    private val clock: TransferClock,
) : ViewModel() {

    private val _state = MutableStateFlow(CameraUploadUiState())
    val state: StateFlow<CameraUploadUiState> = _state.asStateFlow()

    init {
        viewModelScope.launch {
            settingsStore.settings.collect { settings -> _state.update { it.copy(settings = settings) } }
        }
        viewModelScope.launch {
            statusObserver.status.collect { status -> _state.update { it.copy(status = status) } }
        }
    }

    fun onEnableRequested() {
        _state.update { it.copy(enablingStep = EnablingStep.EXPLAINING_PERMISSIONS) }
    }

    fun onEnableCancelled() {
        _state.update { it.copy(enablingStep = EnablingStep.NONE) }
    }

    fun onPermissionRequestStarted() {
        _state.update { it.copy(enablingStep = EnablingStep.NONE) }
    }

    fun onPermissionRequestFinished(isMediaGranted: Boolean) {
        _state.update {
            it.copy(
                isMediaPermissionMissing = !isMediaGranted,
                enablingStep = if (isMediaGranted) EnablingStep.CHOOSING_STARTING_POINT else EnablingStep.NONE,
            )
        }
    }

    fun onMediaPermissionChecked(isMediaGranted: Boolean) {
        _state.update { it.copy(isMediaPermissionMissing = !isMediaGranted && it.settings.isEnabled) }
    }

    fun enableSendingEverything() = enableFrom(startingPointSeconds = 0)

    fun enableSendingFromNow() = enableFrom(startingPointSeconds = clock.nowMillis() / MILLIS_PER_SECOND)

    private fun enableFrom(startingPointSeconds: Long) {
        _state.update { it.copy(enablingStep = EnablingStep.NONE) }
        saveAndApply {
            setStartingPointSeconds(startingPointSeconds)
            setEnabled(true)
        }
    }

    fun disable() = saveAndApply { setEnabled(false) }

    fun setOnlyOnWifi(isOnlyOnWifi: Boolean) = saveAndApply { setOnlyOnWifi(isOnlyOnWifi) }

    fun setOnlyWhileCharging(isOnlyWhileCharging: Boolean) = saveAndApply { setOnlyWhileCharging(isOnlyWhileCharging) }

    fun setIncludeVideos(shouldIncludeVideos: Boolean) = saveAndApply { setIncludeVideos(shouldIncludeVideos) }

    fun pickDestination(destination: UploadDestination) = saveAndApply {
        setDestination(destination)
        setRelativePrefix(if (destination.isRoot) CameraUploadSettings.DEFAULT_RELATIVE_PREFIX else "")
    }

    fun syncNow() = scheduler.syncNow()

    private fun saveAndApply(change: suspend CameraUploadSettingsStore.() -> Unit) {
        viewModelScope.launch {
            settingsStore.change()
            scheduler.apply(settingsStore.current())
        }
    }
}

private const val MILLIS_PER_SECOND = 1000L
