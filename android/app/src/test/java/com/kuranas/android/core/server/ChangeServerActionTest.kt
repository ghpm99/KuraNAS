package com.kuranas.android.core.server

import io.mockk.coVerify
import io.mockk.mockk
import kotlinx.coroutines.test.runTest
import org.junit.Test

class ChangeServerActionTest {

    private val serverStore = mockk<ServerStore>(relaxed = true)
    private val changeServerAction = ChangeServerAction(serverStore)

    @Test
    fun `trocar servidor limpa o servidor lembrado`() = runTest {
        changeServerAction.changeServer()

        coVerify(exactly = 1) { serverStore.clear() }
    }
}
