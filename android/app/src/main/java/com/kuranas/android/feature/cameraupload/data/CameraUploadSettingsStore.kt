package com.kuranas.android.feature.cameraupload.data

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.MutablePreferences
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.longPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.kuranas.android.feature.files.data.UploadDestination
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

private val Context.cameraUploadDataStore: DataStore<Preferences> by preferencesDataStore("camera_upload_prefs")

@Singleton
class CameraUploadSettingsStore(private val dataStore: DataStore<Preferences>) {

    @Inject
    constructor(@ApplicationContext context: Context) : this(context.cameraUploadDataStore)

    private val isEnabledKey = booleanPreferencesKey("is_enabled")
    private val folderIdKey = stringPreferencesKey("destination_folder_id")
    private val folderNameKey = stringPreferencesKey("destination_folder_name")
    private val relativePrefixKey = stringPreferencesKey("relative_prefix")
    private val onlyOnWifiKey = booleanPreferencesKey("only_on_wifi")
    private val onlyWhileChargingKey = booleanPreferencesKey("only_while_charging")
    private val includeVideosKey = booleanPreferencesKey("include_videos")
    private val startingPointSecondsKey = longPreferencesKey("starting_point_seconds")
    private val lastSyncAtMillisKey = longPreferencesKey("last_sync_at_millis")
    private val lastSyncOutcomeKey = stringPreferencesKey("last_sync_outcome")

    val settings: Flow<CameraUploadSettings> = dataStore.data.map { it.toSettings() }

    suspend fun current(): CameraUploadSettings = settings.first()

    suspend fun setEnabled(isEnabled: Boolean) = edit { it[isEnabledKey] = isEnabled }

    suspend fun setDestination(destination: UploadDestination) = edit {
        it[folderIdKey] = destination.folderId.orEmpty()
        it[folderNameKey] = destination.folderName
    }

    suspend fun setRelativePrefix(relativePrefix: String) = edit { it[relativePrefixKey] = relativePrefix }

    suspend fun setOnlyOnWifi(onlyOnWifi: Boolean) = edit { it[onlyOnWifiKey] = onlyOnWifi }

    suspend fun setOnlyWhileCharging(onlyWhileCharging: Boolean) = edit { it[onlyWhileChargingKey] = onlyWhileCharging }

    suspend fun setIncludeVideos(includeVideos: Boolean) = edit { it[includeVideosKey] = includeVideos }

    suspend fun setStartingPointSeconds(startingPointSeconds: Long) = edit { it[startingPointSecondsKey] = startingPointSeconds }

    suspend fun recordSync(syncedAtMillis: Long?, outcome: CameraSyncOutcome) = edit {
        if (syncedAtMillis != null) it[lastSyncAtMillisKey] = syncedAtMillis
        it[lastSyncOutcomeKey] = outcome.name
    }

    private suspend fun edit(change: (MutablePreferences) -> Unit) {
        dataStore.edit { change(it) }
    }

    private fun Preferences.toSettings(): CameraUploadSettings {
        val defaults = CameraUploadSettings()
        return CameraUploadSettings(
            isEnabled = this[isEnabledKey] ?: defaults.isEnabled,
            destination = UploadDestination(
                folderId = this[folderIdKey]?.ifEmpty { null },
                folderName = this[folderNameKey].orEmpty(),
            ),
            relativePrefix = this[relativePrefixKey] ?: defaults.relativePrefix,
            onlyOnWifi = this[onlyOnWifiKey] ?: defaults.onlyOnWifi,
            onlyWhileCharging = this[onlyWhileChargingKey] ?: defaults.onlyWhileCharging,
            includeVideos = this[includeVideosKey] ?: defaults.includeVideos,
            startingPointSeconds = this[startingPointSecondsKey] ?: defaults.startingPointSeconds,
            lastSyncAtMillis = this[lastSyncAtMillisKey],
            lastSyncOutcome = this[lastSyncOutcomeKey]?.let { name -> CameraSyncOutcome.entries.firstOrNull { it.name == name } },
        )
    }
}
