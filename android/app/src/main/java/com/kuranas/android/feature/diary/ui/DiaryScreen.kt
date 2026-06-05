package com.kuranas.android.feature.diary.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.WindowInsetsSides
import androidx.compose.foundation.layout.only
import androidx.compose.foundation.layout.systemBars
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.R
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.diary.data.DiaryEntryDto

@Composable
fun DiaryScreen(viewModel: DiaryViewModel = hiltViewModel()) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    var title by remember { mutableStateOf("") }
    var content by remember { mutableStateOf("") }

    Scaffold(
        containerColor = Color.Transparent,
        contentWindowInsets = WindowInsets.systemBars.only(WindowInsetsSides.Horizontal + WindowInsetsSides.Bottom),
        floatingActionButton = {
            FloatingActionButton(
                onClick = { viewModel.toggleCreateDialog(true) },
                containerColor = MaterialTheme.colorScheme.primary,
            ) { Icon(Icons.Default.Add, contentDescription = stringResource(R.string.diary_new_entry)) }
        },
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(horizontal = 16.dp)) {
            KNHeader(title = stringResource(R.string.nav_diary))
            when {
                state.isLoading -> LoadingView()
                state.error != null -> ErrorView(state.error!!, onRetry = viewModel::load)
                else -> LazyColumn(contentPadding = PaddingValues(bottom = 80.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    items(state.entries, key = { it.id }) { entry ->
                        DiaryEntryCard(entry = entry, onEdit = { viewModel.setEditingEntry(entry) })
                    }
                }
            }
        }
    }

    if (state.showCreateDialog) {
        AlertDialog(
            onDismissRequest = { viewModel.toggleCreateDialog(false) },
            title = { Text(stringResource(R.string.diary_new_entry)) },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedTextField(value = title, onValueChange = { title = it }, label = { Text(stringResource(R.string.diary_title_label)) }, singleLine = true, modifier = Modifier.fillMaxWidth())
                    OutlinedTextField(value = content, onValueChange = { content = it }, label = { Text(stringResource(R.string.diary_content_label)) }, minLines = 4, modifier = Modifier.fillMaxWidth())
                }
            },
            confirmButton = {
                TextButton(onClick = { viewModel.createEntry(title, content); title = ""; content = "" }) { Text(stringResource(R.string.action_create)) }
            },
            dismissButton = { TextButton(onClick = { viewModel.toggleCreateDialog(false) }) { Text(stringResource(R.string.action_cancel)) } },
        )
    }

    state.editingEntry?.let { entry ->
        var editTitle by remember(entry.id) { mutableStateOf(entry.title) }
        var editContent by remember(entry.id) { mutableStateOf(entry.content) }
        AlertDialog(
            onDismissRequest = { viewModel.setEditingEntry(null) },
            title = { Text(stringResource(R.string.diary_edit_entry)) },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedTextField(value = editTitle, onValueChange = { editTitle = it }, label = { Text(stringResource(R.string.diary_title_label)) }, singleLine = true, modifier = Modifier.fillMaxWidth())
                    OutlinedTextField(value = editContent, onValueChange = { editContent = it }, label = { Text(stringResource(R.string.diary_content_label)) }, minLines = 4, modifier = Modifier.fillMaxWidth())
                }
            },
            confirmButton = {
                TextButton(onClick = { viewModel.updateEntry(entry.id, editTitle, editContent) }) { Text(stringResource(R.string.action_save)) }
            },
            dismissButton = { TextButton(onClick = { viewModel.setEditingEntry(null) }) { Text(stringResource(R.string.action_cancel)) } },
        )
    }
}

@Composable
private fun DiaryEntryCard(entry: DiaryEntryDto, onEdit: () -> Unit) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 16.dp)
            .clickable(onClick = onEdit)
            .padding(16.dp),
    ) {
        Text(entry.title, style = MaterialTheme.typography.titleMedium)
        Spacer(Modifier.height(4.dp))
        Text(entry.content, style = MaterialTheme.typography.bodyMedium, maxLines = 3)
        Spacer(Modifier.height(8.dp))
        Text(entry.createdAt, style = MaterialTheme.typography.bodySmall)
    }
}
