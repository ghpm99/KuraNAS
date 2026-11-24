package com.kuranas.android.core.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.compositionLocalOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberUpdatedState
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.emptyFlow

val LocalServerReconnections = compositionLocalOf<Flow<Unit>> { emptyFlow() }

/** Chama [onReconnected] toda vez que o app volta a falar com o servidor enquanto a tela está visível. */
@Composable
fun ReconnectionRefresh(onReconnected: () -> Unit) {
    val reconnections = LocalServerReconnections.current
    val latestOnReconnected by rememberUpdatedState(onReconnected)
    LaunchedEffect(reconnections) {
        reconnections.collect { latestOnReconnected() }
    }
}
