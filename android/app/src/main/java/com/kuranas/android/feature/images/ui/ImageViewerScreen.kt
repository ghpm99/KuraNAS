package com.kuranas.android.feature.images.ui

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.StarBorder
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.kuranas.android.R

@Composable
fun ImageViewerScreen(
    fileId: String,
    onNavigateBack: () -> Unit,
    viewModel: ImageViewerViewModel = hiltViewModel(),
) {
    val starredById by viewModel.starredById.collectAsStateWithLifecycle()
    val actionErrorMessage by viewModel.actionErrorMessage.collectAsStateWithLifecycle()
    val isStarred = starredById[fileId] ?: false

    val snackbarHostState = remember { SnackbarHostState() }
    LaunchedEffect(actionErrorMessage) {
        val message = actionErrorMessage ?: return@LaunchedEffect
        snackbarHostState.showSnackbar(message)
        viewModel.dismissActionError()
    }

    Box(modifier = Modifier.fillMaxSize()) {
        AsyncImage(
            model = viewModel.getBlobUrl(fileId),
            contentDescription = null,
            contentScale = ContentScale.Fit,
            modifier = Modifier.fillMaxSize(),
        )
        IconButton(
            onClick = onNavigateBack,
            modifier = Modifier.align(Alignment.TopStart).padding(16.dp),
        ) {
            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = stringResource(R.string.action_back))
        }
        IconButton(
            onClick = { viewModel.toggleStarred(fileId) },
            modifier = Modifier.align(Alignment.TopEnd).padding(16.dp),
        ) {
            Icon(
                imageVector = if (isStarred) Icons.Default.Star else Icons.Default.StarBorder,
                contentDescription = stringResource(if (isStarred) R.string.files_unstar_action else R.string.files_star_action),
            )
        }
        SnackbarHost(snackbarHostState, modifier = Modifier.align(Alignment.BottomCenter))
    }
}
