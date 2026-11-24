package com.kuranas.android.feature.cameraupload.ui

import android.content.Context
import android.content.Intent
import android.net.Uri
import android.provider.Settings
import android.text.format.DateUtils
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.annotation.StringRes
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.cameraupload.data.CameraSyncOutcome
import com.kuranas.android.feature.cameraupload.data.CameraUploadSettings
import com.kuranas.android.feature.cameraupload.data.CameraUploadStatus
import com.kuranas.android.feature.files.data.UploadDestination
import com.kuranas.android.feature.files.ui.DestinationChooser
import com.kuranas.android.feature.files.ui.destinationLabel
import com.kuranas.android.ui.theme.StatusNegative

@Composable
fun CameraUploadScreen(
    onNavigateBack: () -> Unit,
    onOpenTransfers: () -> Unit,
    viewModel: CameraUploadViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val context = LocalContext.current
    var isPickingFolder by rememberSaveable { mutableStateOf(false) }
    val includeVideos by rememberUpdatedState(state.settings.includeVideos)

    val permissionLauncher = rememberLauncherForActivityResult(ActivityResultContracts.RequestMultiplePermissions()) {
        viewModel.onPermissionRequestFinished(hasMediaAccess(grantedPermissionsOf(context), includeVideos))
    }

    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) {
        viewModel.onMediaPermissionChecked(hasMediaAccess(grantedPermissionsOf(context), includeVideos))
    }
    LaunchedEffect(state.settings.includeVideos, state.settings.isEnabled) {
        viewModel.onMediaPermissionChecked(hasMediaAccess(grantedPermissionsOf(context), includeVideos))
    }

    if (isPickingFolder) {
        Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp).navigationBarsPadding()) {
            KNHeader(
                title = stringResource(R.string.camera_upload_pick_folder_title),
                leadingIcon = Icons.AutoMirrored.Filled.ArrowBack,
                onLeadingClick = { isPickingFolder = false },
            )
            DestinationChooser(
                selectedFileCount = null,
                lastDestination = state.settings.destination,
                confirmLabelRes = R.string.camera_upload_use_folder,
                onConfirm = { destination, _ ->
                    viewModel.pickDestination(destination)
                    isPickingFolder = false
                },
            )
        }
        return
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 16.dp)
            .verticalScroll(rememberScrollState())
            .navigationBarsPadding(),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        KNHeader(
            title = stringResource(R.string.camera_upload_title),
            leadingIcon = Icons.AutoMirrored.Filled.ArrowBack,
            onLeadingClick = onNavigateBack,
        )
        EnableSection(
            settings = state.settings,
            onToggle = { shouldEnable -> if (shouldEnable) viewModel.onEnableRequested() else viewModel.disable() },
        )
        if (state.isMediaPermissionMissing) {
            MissingPermissionSection(onOpenSystemSettings = { openAppSystemSettings(context) })
        }
        OptionsSection(
            settings = state.settings,
            onPickFolder = { isPickingFolder = true },
            onOnlyOnWifiChange = viewModel::setOnlyOnWifi,
            onOnlyWhileChargingChange = viewModel::setOnlyWhileCharging,
            onIncludeVideosChange = viewModel::setIncludeVideos,
        )
        StatusSection(
            status = state.status,
            isEnabled = state.settings.isEnabled,
            onSyncNow = viewModel::syncNow,
            onOpenTransfers = onOpenTransfers,
        )
        Text(
            stringResource(R.string.camera_upload_never_deletes),
            style = MaterialTheme.typography.bodySmall,
            modifier = Modifier.padding(bottom = 16.dp),
        )
    }

    when (state.enablingStep) {
        EnablingStep.NONE -> Unit
        EnablingStep.EXPLAINING_PERMISSIONS -> PermissionRationaleDialog(
            onContinue = {
                viewModel.onPermissionRequestStarted()
                permissionLauncher.launch(permissionsToRequestFor(includeVideos).toTypedArray())
            },
            onDismiss = viewModel::onEnableCancelled,
        )
        EnablingStep.CHOOSING_STARTING_POINT -> StartingPointDialog(
            onSendEverything = viewModel::enableSendingEverything,
            onSendFromNow = viewModel::enableSendingFromNow,
            onDismiss = viewModel::onEnableCancelled,
        )
    }
}

@Composable
private fun EnableSection(settings: CameraUploadSettings, onToggle: (Boolean) -> Unit) {
    Column(modifier = Modifier.fillMaxWidth().glass(GlassLevel.Light).padding(16.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
        SwitchRow(R.string.camera_upload_enable, settings.isEnabled, onToggle)
        Text(stringResource(R.string.camera_upload_enable_description), style = MaterialTheme.typography.bodySmall)
    }
}

@Composable
private fun MissingPermissionSection(onOpenSystemSettings: () -> Unit) {
    Column(modifier = Modifier.fillMaxWidth().glass(GlassLevel.Strong).padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(stringResource(R.string.camera_upload_permission_missing_title), style = MaterialTheme.typography.titleMedium, color = StatusNegative)
        Text(stringResource(R.string.camera_upload_permission_missing_text), style = MaterialTheme.typography.bodyMedium)
        Button(onClick = onOpenSystemSettings) { Text(stringResource(R.string.camera_upload_open_system_settings)) }
    }
}

@Composable
private fun OptionsSection(
    settings: CameraUploadSettings,
    onPickFolder: () -> Unit,
    onOnlyOnWifiChange: (Boolean) -> Unit,
    onOnlyWhileChargingChange: (Boolean) -> Unit,
    onIncludeVideosChange: (Boolean) -> Unit,
) {
    Column(modifier = Modifier.fillMaxWidth().glass(GlassLevel.Flat).padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(stringResource(R.string.camera_upload_destination_title), style = MaterialTheme.typography.titleMedium)
        Text(destinationExample(settings), style = MaterialTheme.typography.bodyMedium)
        OutlinedButton(onClick = onPickFolder) { Text(stringResource(R.string.camera_upload_change_folder)) }
        SwitchRow(R.string.camera_upload_only_wifi, settings.onlyOnWifi, onOnlyOnWifiChange)
        SwitchRow(R.string.camera_upload_only_charging, settings.onlyWhileCharging, onOnlyWhileChargingChange)
        SwitchRow(R.string.camera_upload_include_videos, settings.includeVideos, onIncludeVideosChange)
    }
}

@Composable
private fun StatusSection(
    status: CameraUploadStatus?,
    isEnabled: Boolean,
    onSyncNow: () -> Unit,
    onOpenTransfers: () -> Unit,
) {
    val context = LocalContext.current
    val lastSyncText = status?.lastSyncAtMillis?.let { formatSyncTime(context, it) } ?: stringResource(R.string.camera_upload_never)
    Column(modifier = Modifier.fillMaxWidth().glass(GlassLevel.Flat).padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(stringResource(R.string.camera_upload_status_title), style = MaterialTheme.typography.titleMedium)
        Text(stringResource(R.string.camera_upload_last_sync, lastSyncText), style = MaterialTheme.typography.bodyMedium)
        status?.lastSyncOutcome?.let { outcome ->
            Text(stringResource(outcomeMessageRes(outcome)), style = MaterialTheme.typography.bodySmall)
        }
        Text(stringResource(R.string.camera_upload_pending, status?.pendingCount ?: 0), style = MaterialTheme.typography.bodyMedium)
        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Text(stringResource(R.string.camera_upload_failed, status?.failedCount ?: 0), style = MaterialTheme.typography.bodyMedium)
            if ((status?.failedCount ?: 0) > 0) {
                TextButton(onClick = onOpenTransfers) { Text(stringResource(R.string.camera_upload_open_transfers)) }
            }
        }
        Button(onClick = onSyncNow, enabled = isEnabled) { Text(stringResource(R.string.camera_upload_sync_now)) }
    }
}

@Composable
private fun SwitchRow(@StringRes labelRes: Int, isChecked: Boolean, onCheckedChange: (Boolean) -> Unit) {
    Row(modifier = Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceBetween) {
        Text(stringResource(labelRes), modifier = Modifier.weight(1f), style = MaterialTheme.typography.bodyLarge)
        Switch(checked = isChecked, onCheckedChange = onCheckedChange)
    }
}

@Composable
private fun PermissionRationaleDialog(onContinue: () -> Unit, onDismiss: () -> Unit) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(stringResource(R.string.camera_upload_rationale_title)) },
        text = { Text(stringResource(R.string.camera_upload_rationale_text)) },
        confirmButton = { TextButton(onClick = onContinue) { Text(stringResource(R.string.camera_upload_rationale_continue)) } },
        dismissButton = { TextButton(onClick = onDismiss) { Text(stringResource(R.string.action_cancel)) } },
    )
}

@Composable
private fun StartingPointDialog(onSendEverything: () -> Unit, onSendFromNow: () -> Unit, onDismiss: () -> Unit) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(stringResource(R.string.camera_upload_starting_point_title)) },
        text = { Text(stringResource(R.string.camera_upload_starting_point_text)) },
        confirmButton = { TextButton(onClick = onSendEverything) { Text(stringResource(R.string.camera_upload_send_everything)) } },
        dismissButton = { TextButton(onClick = onSendFromNow) { Text(stringResource(R.string.camera_upload_send_from_now)) } },
    )
}

@Composable
private fun destinationExample(settings: CameraUploadSettings): String {
    val folderLabel = destinationLabel(settings.destination)
    val dateSuffix = stringResource(R.string.camera_upload_date_suffix)
    return listOf(folderLabel, settings.relativePrefix, dateSuffix).filter { it.isNotEmpty() }.joinToString("/")
}

@StringRes
private fun outcomeMessageRes(outcome: CameraSyncOutcome): Int = when (outcome) {
    CameraSyncOutcome.DISABLED -> R.string.camera_upload_outcome_disabled
    CameraSyncOutcome.AWAY_FROM_HOME -> R.string.camera_upload_outcome_away
    CameraSyncOutcome.MEDIA_PERMISSION_MISSING -> R.string.camera_upload_outcome_permission
    CameraSyncOutcome.UP_TO_DATE -> R.string.camera_upload_outcome_up_to_date
    CameraSyncOutcome.ENQUEUED -> R.string.camera_upload_outcome_enqueued
}

private fun formatSyncTime(context: Context, millis: Long): String =
    DateUtils.formatDateTime(
        context,
        millis,
        DateUtils.FORMAT_SHOW_DATE or DateUtils.FORMAT_SHOW_TIME or DateUtils.FORMAT_SHOW_YEAR,
    )

private fun grantedPermissionsOf(context: Context): Set<String> =
    permissionsToRequestFor(includeVideos = true)
        .filter { ContextCompat.checkSelfPermission(context, it) == android.content.pm.PackageManager.PERMISSION_GRANTED }
        .toSet()

private fun openAppSystemSettings(context: Context) {
    val intent = Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS, Uri.fromParts("package", context.packageName, null))
        .addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
    context.startActivity(intent)
}
