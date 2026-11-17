package com.kuranas.android.core.ui.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.kuranas.android.R

/**
 * Footer for paged lists: a progress indicator while the next page loads, or the
 * next-page error message with a retry button, or nothing.
 */
@Composable
fun PagedListFooter(
    isLoadingMore: Boolean,
    loadMoreError: String?,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
) {
    if (loadMoreError == null && !isLoadingMore) return
    Column(
        modifier = modifier.fillMaxWidth().padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        if (loadMoreError != null) {
            Text(text = loadMoreError, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.error)
            TextButton(onClick = onRetry) { Text(stringResource(R.string.action_retry)) }
        } else {
            CircularProgressIndicator(modifier = Modifier.size(24.dp), strokeWidth = 2.dp)
        }
    }
}
