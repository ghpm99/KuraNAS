package com.kuranas.android.feature.settings.ui

import com.kuranas.android.core.server.ChangeServerAction
import com.kuranas.android.core.server.ServerStore
import com.kuranas.android.feature.files.data.UploadDestinationStore
import com.kuranas.android.feature.settings.data.ConfigApi
import androidx.lifecycle.viewModelScope
import io.mockk.coEvery
import io.mockk.coVerify
import io.mockk.every
import io.mockk.mockk
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.awaitCancellation
import kotlinx.coroutines.cancelAndJoin
import kotlinx.coroutines.job
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class SettingsViewModelTest {

    private val api = mockk<ConfigApi>()
    private val serverStore = mockk<ServerStore>()
    private val changeServerAction = mockk<ChangeServerAction>(relaxed = true)
    private val uploadDestinationStore = mockk<UploadDestinationStore>(relaxed = true)

    private lateinit var viewModel: SettingsViewModel

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
        coEvery { api.getAbout() } coAnswers { awaitCancellation() }
        every { serverStore.serverUrl } returns flowOf("http://192.168.0.10:8000")
        every { uploadDestinationStore.alwaysSendDestination } returns flowOf(null)
    }

    @After
    fun tearDown() {
        runBlocking { viewModel.viewModelScope.coroutineContext.job.cancelAndJoin() }
        Dispatchers.resetMain()
    }

    @Test
    fun `trocar servidor delega para a acao compartilhada`() = runTest {
        viewModel = SettingsViewModel(api, serverStore, changeServerAction, uploadDestinationStore)
        viewModel.changeServer()

        coVerify(exactly = 1) { changeServerAction.changeServer() }
    }
}
