package com.kuranas.android.core.server

sealed class ServerState {
    object NotConfigured : ServerState()
    object Discovering : ServerState()
    data class Found(val url: String, val name: String) : ServerState()
    data class Manual(val url: String) : ServerState()

    val baseUrl: String?
        get() = when (this) {
            is Found -> url
            is Manual -> url
            else -> null
        }

    val isConnected: Boolean
        get() = this is Found || this is Manual
}
