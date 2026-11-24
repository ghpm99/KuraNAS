package com.kuranas.android.core.server

/** Resultado da validação de um endereço como servidor KuraNAS. */
sealed interface ServerCheckResult {
    data object Confirmed : ServerCheckResult
    data object NoResponse : ServerCheckResult
    data object NotKuraNas : ServerCheckResult
    data class AccessDenied(val denial: AccessDenial) : ServerCheckResult
}
