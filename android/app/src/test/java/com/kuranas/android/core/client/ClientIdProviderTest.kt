package com.kuranas.android.core.client

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import androidx.datastore.preferences.core.Preferences
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.rules.TemporaryFolder
import java.io.File

class ClientIdProviderTest {

    @get:Rule
    val temporaryFolder = TemporaryFolder()

    private val backendClientIdPattern = Regex("^[A-Za-z0-9-]{8,64}$")

    private fun newDataStore(
        file: File,
        scope: CoroutineScope = CoroutineScope(SupervisorJob() + Dispatchers.IO),
    ): DataStore<Preferences> =
        PreferenceDataStoreFactory.create(scope = scope, produceFile = { file })

    @Test
    fun `generates a client id accepted by the backend pattern`() {
        val provider = ClientIdProvider(newDataStore(temporaryFolder.newFile("a.preferences_pb").also { it.delete() }))

        assertTrue(backendClientIdPattern.matches(provider.clientId()))
    }

    @Test
    fun `second call returns the same client id`() {
        val provider = ClientIdProvider(newDataStore(temporaryFolder.newFile("b.preferences_pb").also { it.delete() }))

        assertEquals(provider.clientId(), provider.clientId())
    }

    @Test
    fun `persisted client id survives a new provider instance`() {
        val file = temporaryFolder.newFile("c.preferences_pb").also { it.delete() }
        val firstStoreScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
        val firstClientId = ClientIdProvider(newDataStore(file, firstStoreScope)).clientId()
        firstStoreScope.cancel()

        val secondClientId = ClientIdProvider(newDataStore(file)).clientId()

        assertEquals(firstClientId, secondClientId)
    }
}
