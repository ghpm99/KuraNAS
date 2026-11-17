package com.kuranas.android.feature.files.data

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.MutablePreferences
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

private val Context.uploadDestinationDataStore: DataStore<Preferences> by preferencesDataStore("upload_destination_prefs")

/** Persiste o último destino de envio e o destino fixo de "Sempre enviar para cá". */
@Singleton
class UploadDestinationStore(private val dataStore: DataStore<Preferences>) {

    @Inject
    constructor(@ApplicationContext context: Context) : this(context.uploadDestinationDataStore)

    private val lastFolderIdKey = stringPreferencesKey("last_folder_id")
    private val lastFolderNameKey = stringPreferencesKey("last_folder_name")
    private val alwaysFolderIdKey = stringPreferencesKey("always_folder_id")
    private val alwaysFolderNameKey = stringPreferencesKey("always_folder_name")

    val lastDestination: Flow<UploadDestination?> =
        dataStore.data.map { it.readDestination(lastFolderIdKey, lastFolderNameKey) }

    val alwaysSendDestination: Flow<UploadDestination?> =
        dataStore.data.map { it.readDestination(alwaysFolderIdKey, alwaysFolderNameKey) }

    suspend fun saveLastDestination(destination: UploadDestination) {
        dataStore.edit { it.writeDestination(lastFolderIdKey, lastFolderNameKey, destination) }
    }

    suspend fun saveAlwaysSendDestination(destination: UploadDestination) {
        dataStore.edit { it.writeDestination(alwaysFolderIdKey, alwaysFolderNameKey, destination) }
    }

    suspend fun clearAlwaysSendDestination() {
        dataStore.edit {
            it.remove(alwaysFolderIdKey)
            it.remove(alwaysFolderNameKey)
        }
    }

    private fun Preferences.readDestination(
        folderIdKey: Preferences.Key<String>,
        folderNameKey: Preferences.Key<String>,
    ): UploadDestination? {
        val storedFolderId = this[folderIdKey] ?: return null
        return UploadDestination(
            folderId = storedFolderId.ifEmpty { null },
            folderName = this[folderNameKey].orEmpty(),
        )
    }

    private fun MutablePreferences.writeDestination(
        folderIdKey: Preferences.Key<String>,
        folderNameKey: Preferences.Key<String>,
        destination: UploadDestination,
    ) {
        this[folderIdKey] = destination.folderId.orEmpty()
        this[folderNameKey] = destination.folderName
    }
}
