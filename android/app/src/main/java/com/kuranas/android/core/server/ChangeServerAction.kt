package com.kuranas.android.core.server

import javax.inject.Inject

/**
 * Esquece o servidor lembrado. A navegação raiz volta sozinha para a tela de conexão porque
 * observa [ServerStore.serverState].
 */
class ChangeServerAction @Inject constructor(private val serverStore: ServerStore) {

    suspend fun changeServer() {
        serverStore.clear()
    }
}
