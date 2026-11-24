package com.kuranas.android.core.server

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject
import javax.inject.Singleton

/** Recusa vigente do servidor a este aparelho, ou null quando o acesso está liberado. */
@Singleton
class AccessDeniedEvents @Inject constructor() {

    private val _denial = MutableStateFlow<AccessDenial?>(null)
    val denial: StateFlow<AccessDenial?> = _denial.asStateFlow()

    fun report(denial: AccessDenial) {
        _denial.value = denial
    }

    fun clear() {
        _denial.value = null
    }
}
