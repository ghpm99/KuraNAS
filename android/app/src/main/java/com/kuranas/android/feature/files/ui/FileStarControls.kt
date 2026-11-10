package com.kuranas.android.feature.files.ui

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.StarBorder
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LocalContentColor
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R

@Composable
fun FileStarErrorHost(stars: FileStarState, content: @Composable () -> Unit) {
    val actionErrorMessage by stars.actionErrorMessage.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }
    LaunchedEffect(actionErrorMessage) {
        val message = actionErrorMessage ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(message)
        stars.dismissActionError()
    }
    Box(modifier = Modifier.fillMaxSize()) {
        content()
        SnackbarHost(snackbarHostState, modifier = Modifier.align(Alignment.BottomCenter))
    }
}

@Composable
fun FileStarButton(isStarred: Boolean, onToggleStar: () -> Unit, modifier: Modifier = Modifier) {
    IconButton(onClick = onToggleStar, modifier = modifier) {
        Icon(
            imageVector = if (isStarred) Icons.Default.Star else Icons.Default.StarBorder,
            contentDescription = stringResource(if (isStarred) R.string.files_unstar_action else R.string.files_star_action),
            tint = if (isStarred) MaterialTheme.colorScheme.secondary else LocalContentColor.current,
        )
    }
}
