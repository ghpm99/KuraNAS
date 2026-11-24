package com.kuranas.android.core.ui.components

import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import com.kuranas.android.R

@Composable
fun ChangeServerConfirmationDialog(onConfirm: () -> Unit, onDismiss: () -> Unit) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(stringResource(R.string.change_server_dialog_title)) },
        text = { Text(stringResource(R.string.change_server_dialog_text)) },
        confirmButton = { TextButton(onClick = onConfirm) { Text(stringResource(R.string.action_change_server)) } },
        dismissButton = { TextButton(onClick = onDismiss) { Text(stringResource(R.string.action_cancel)) } },
    )
}
