package com.kuranas.android.feature.jobs.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Cancel
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Error
import androidx.compose.material.icons.filled.Pending
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.core.ui.components.EmptyView
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.jobs.data.JobDto
import com.kuranas.android.ui.theme.StatusAlert
import com.kuranas.android.ui.theme.StatusNegative
import com.kuranas.android.ui.theme.StatusPositive

@Composable
fun JobsScreen(viewModel: JobsViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(title = "Jobs", trailingIcon = Icons.Default.Refresh, onTrailingClick = viewModel::load)
        when {
            state.isLoading -> LoadingView()
            state.error != null -> ErrorView(state.error!!, onRetry = viewModel::load)
            state.jobs.isEmpty() -> EmptyView("Nenhum job encontrado")
            else -> LazyColumn(contentPadding = PaddingValues(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                items(state.jobs, key = { it.id }) { job ->
                    JobCard(job = job, onCancel = { viewModel.cancelJob(job.id) })
                }
            }
        }
    }
}

@Composable
private fun JobCard(job: JobDto, onCancel: () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 16.dp)
            .padding(16.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween,
        ) {
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Icon(
                    imageVector = when (job.status) {
                        "done", "completed" -> Icons.Default.CheckCircle
                        "failed", "error" -> Icons.Default.Error
                        "running", "pending" -> Icons.Default.Pending
                        else -> Icons.Default.Pending
                    },
                    contentDescription = null,
                    modifier = Modifier.size(18.dp),
                    tint = when (job.status) {
                        "done", "completed" -> StatusPositive
                        "failed", "error" -> StatusNegative
                        else -> StatusAlert
                    },
                )
                Text(job.name, style = MaterialTheme.typography.bodyMedium)
            }
            if (job.status in listOf("running", "pending")) {
                IconButton(onClick = onCancel, modifier = Modifier.size(32.dp)) {
                    Icon(Icons.Default.Cancel, contentDescription = "Cancelar", modifier = Modifier.size(18.dp), tint = StatusNegative)
                }
            }
        }
        if (job.status == "running") {
            Spacer(Modifier.height(8.dp))
            LinearProgressIndicator(progress = { job.progress }, modifier = Modifier.fillMaxWidth())
            Text("${(job.progress * 100).toInt()}%", style = MaterialTheme.typography.bodySmall)
        }
        job.error?.let { err ->
            Spacer(Modifier.height(4.dp))
            Text(err, style = MaterialTheme.typography.bodySmall, color = StatusNegative)
        }
        Text(job.updatedAt, style = MaterialTheme.typography.bodySmall, modifier = Modifier.padding(top = 4.dp))
    }
}
