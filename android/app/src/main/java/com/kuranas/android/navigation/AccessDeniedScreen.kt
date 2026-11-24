package com.kuranas.android.navigation

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import android.content.ClipData
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.ClipEntry
import androidx.compose.ui.platform.LocalClipboard
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.server.AccessDenial
import com.kuranas.android.core.ui.components.ChangeServerConfirmationDialog
import kotlinx.coroutines.launch

@Composable
fun AccessDeniedScreen(
    denial: AccessDenial,
    modifier: Modifier = Modifier,
    rootViewModel: RootViewModel = hiltViewModel(),
) {
    val isRetrying by rootViewModel.isRetrying.collectAsStateWithLifecycle()
    AccessDeniedContent(
        denial = denial,
        isRetrying = isRetrying,
        onRetry = rootViewModel::retryAfterAccessDenied,
        onChangeServer = rootViewModel::changeServerAfterAccessDenied,
        modifier = modifier,
    )
}

@Composable
internal fun AccessDeniedContent(
    denial: AccessDenial,
    isRetrying: Boolean,
    onRetry: () -> Unit,
    onChangeServer: () -> Unit,
    modifier: Modifier = Modifier,
) {
    var isChangeServerDialogVisible by rememberSaveable { mutableStateOf(false) }
    val clipboard = LocalClipboard.current
    val coroutineScope = rememberCoroutineScope()

    Column(
        modifier = modifier.fillMaxSize().padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp, Alignment.CenterVertically),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Text(
            text = stringResource(R.string.access_denied_title),
            style = MaterialTheme.typography.headlineSmall,
            textAlign = TextAlign.Center,
        )
        Text(
            text = stringResource(R.string.access_denied_explanation),
            style = MaterialTheme.typography.bodyMedium,
            textAlign = TextAlign.Center,
        )
        if (denial.deviceIp == null) {
            Text(
                text = stringResource(R.string.access_denied_ip_unknown),
                style = MaterialTheme.typography.bodyMedium,
                textAlign = TextAlign.Center,
            )
        } else {
            Text(stringResource(R.string.access_denied_ip_label), style = MaterialTheme.typography.labelLarge)
            SelectionContainer {
                Text(
                    text = denial.deviceIp,
                    style = MaterialTheme.typography.headlineMedium,
                    fontWeight = FontWeight.Bold,
                )
            }
            OutlinedButton(
                onClick = {
                    coroutineScope.launch {
                        clipboard.setClipEntry(ClipEntry(ClipData.newPlainText(denial.deviceIp, denial.deviceIp)))
                    }
                },
            ) {
                Text(stringResource(R.string.access_denied_copy_ip))
            }
        }
        denial.serverMessage?.let { serverMessage ->
            Text(text = serverMessage, style = MaterialTheme.typography.bodySmall, textAlign = TextAlign.Center)
        }
        Button(onClick = onRetry, enabled = !isRetrying) {
            Text(stringResource(R.string.access_denied_retry))
        }
        TextButton(onClick = { isChangeServerDialogVisible = true }) {
            Text(stringResource(R.string.action_change_server))
        }
    }

    if (isChangeServerDialogVisible) {
        ChangeServerConfirmationDialog(
            onConfirm = {
                isChangeServerDialogVisible = false
                onChangeServer()
            },
            onDismiss = { isChangeServerDialogVisible = false },
        )
    }
}
