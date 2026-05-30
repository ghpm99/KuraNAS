package com.kuranas.android.core.server

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import javax.inject.Inject
import javax.inject.Singleton

private val Context.dataStore: DataStore<Preferences> by preferencesDataStore("server_prefs")

@Singleton
class ServerStore @Inject constructor(@ApplicationContext private val context: Context) {

    private val urlKey = stringPreferencesKey("server_url")
    private val nameKey = stringPreferencesKey("server_name")

    val serverUrl: Flow<String?> = context.dataStore.data.map { it[urlKey] }

    val serverState: Flow<ServerState> = context.dataStore.data.map { prefs ->
        val url = prefs[urlKey]
        val name = prefs[nameKey]
        when {
            url == null -> ServerState.NotConfigured
            name != null -> ServerState.Found(url, name)
            else -> ServerState.Manual(url)
        }
    }

    suspend fun saveDiscovered(url: String, name: String) {
        context.dataStore.edit {
            it[urlKey] = url
            it[nameKey] = name
        }
    }

    suspend fun saveManual(url: String) {
        context.dataStore.edit {
            it[urlKey] = url
            it.remove(nameKey)
        }
    }

    suspend fun clear() {
        context.dataStore.edit {
            it.remove(urlKey)
            it.remove(nameKey)
        }
    }
}
