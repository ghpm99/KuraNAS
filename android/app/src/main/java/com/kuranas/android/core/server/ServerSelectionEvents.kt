package com.kuranas.android.core.server

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Sinaliza que o usuário escolheu um servidor na tela de conexão, mesmo quando o valor salvo
 * não muda. [selectionCount] só cresce; quem precisa reagir guarda o valor visto e compara.
 */
@Singleton
class ServerSelectionEvents @Inject constructor() {

    private val _selectionCount = MutableStateFlow(0)
    val selectionCount: StateFlow<Int> = _selectionCount.asStateFlow()

    fun notifyServerSelected() {
        _selectionCount.update { it + 1 }
    }
}
