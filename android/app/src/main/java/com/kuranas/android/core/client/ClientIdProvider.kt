package com.kuranas.android.core.client

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import java.util.UUID
import javax.inject.Inject
import javax.inject.Singleton

private val Context.clientIdDataStore: DataStore<Preferences> by preferencesDataStore("client_id_prefs")

@Singleton
class ClientIdProvider(private val dataStore: DataStore<Preferences>) {

    @Inject
    constructor(@ApplicationContext context: Context) : this(context.clientIdDataStore)

    private val clientIdKey = stringPreferencesKey("client_id")

    @Volatile
    private var cachedClientId: String? = null

    /** Returns this installation's stable client id, generating and persisting it on first use. */
    fun clientId(): String {
        cachedClientId?.let { return it }
        return synchronized(this) {
            cachedClientId ?: loadOrCreateClientId().also { cachedClientId = it }
        }
    }

    private fun loadOrCreateClientId(): String = runBlocking {
        dataStore.data.first()[clientIdKey]?.let { return@runBlocking it }
        val generatedClientId = UUID.randomUUID().toString()
        val persisted = dataStore.edit { preferences ->
            if (preferences[clientIdKey] == null) preferences[clientIdKey] = generatedClientId
        }
        persisted[clientIdKey] ?: generatedClientId
    }
}
