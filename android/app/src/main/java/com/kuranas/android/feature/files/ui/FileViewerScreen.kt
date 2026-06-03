package com.kuranas.android.feature.files.ui

import android.widget.Toast
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.InsertDriveFile
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.kuranas.android.core.ui.components.ErrorView
import com.kuranas.android.core.ui.components.FileSizeText
import com.kuranas.android.core.ui.components.GlassLevel
import com.kuranas.android.core.ui.components.KNHeader
import com.kuranas.android.core.ui.components.LoadingView
import com.kuranas.android.core.ui.components.glass
import com.kuranas.android.feature.files.data.FileContent
import com.kuranas.android.feature.files.data.FileItemDto
import java.time.OffsetDateTime
import java.time.format.DateTimeFormatter

@Composable
fun FileViewerScreen(
    fileId: String,
    fileName: String,
    onNavigateBack: () -> Unit,
    viewModel: FileViewerViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val downloading by viewModel.downloading.collectAsStateWithLifecycle()
    val context = LocalContext.current

    LaunchedEffect(fileId) { viewModel.load(fileId) }
    LaunchedEffect(Unit) {
        viewModel.messages.collect { msg -> Toast.makeText(context, msg, Toast.LENGTH_SHORT).show() }
    }

    Column(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        KNHeader(
            title = fileName.ifBlank { "Arquivo" },
            leadingIcon = Icons.AutoMirrored.Filled.ArrowBack,
            onLeadingClick = onNavigateBack,
        )

        when (val s = state) {
            is FileViewerState.Loading -> LoadingView()
            is FileViewerState.Failed -> ErrorView(s.message, onRetry = { viewModel.load(fileId) })
            is FileViewerState.Content -> {
                var showInfo by remember { mutableStateOf(false) }

                ActionsRow(
                    downloading = downloading,
                    showInfo = showInfo,
                    onToggleInfo = { showInfo = !showInfo },
                    onDownload = { viewModel.download(fileId, fileName) },
                )

                if (showInfo) {
                    Spacer(Modifier.height(8.dp))
                    InfoCard(info = s.info, content = s.content)
                }

                Spacer(Modifier.height(8.dp))

                when (val content = s.content) {
                    is FileContent.Text -> TextContent(content)
                    is FileContent.Unsupported -> UnsupportedContent(content)
                }
            }
        }
    }
}

@Composable
private fun ActionsRow(
    downloading: Boolean,
    showInfo: Boolean,
    onToggleInfo: () -> Unit,
    onDownload: () -> Unit,
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        TextButton(onClick = onToggleInfo) {
            Icon(Icons.Default.InsertDriveFile, contentDescription = null, modifier = Modifier.size(18.dp))
            Spacer(Modifier.size(6.dp))
            Text(if (showInfo) "Ocultar detalhes" else "Detalhes")
        }
        OutlinedButton(onClick = onDownload, enabled = !downloading) {
            if (downloading) {
                CircularProgressIndicator(modifier = Modifier.size(18.dp), strokeWidth = 2.dp)
            } else {
                Icon(Icons.Default.Download, contentDescription = null, modifier = Modifier.size(18.dp))
            }
            Spacer(Modifier.size(6.dp))
            Text(if (downloading) "Baixando…" else "Baixar")
        }
    }
}

@Composable
private fun InfoCard(info: FileItemDto?, content: FileContent) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .glass(GlassLevel.Light, radius = 12.dp)
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        Text("Informações", style = MaterialTheme.typography.titleSmall)
        if (info != null) {
            InfoRow("Nome", info.name)
            InfoRow("Caminho", info.path.ifBlank { info.parentPath })
            InfoRow("Formato", info.format.ifBlank { "—" })
            InfoRow("Tipo", if (info.isDir) "Pasta" else "Arquivo")
            InfoLine("Tamanho") { FileSizeText(info.size) }
            if (info.createdAt.isNotBlank()) InfoRow("Criado em", formatDate(info.createdAt))
            if (info.updatedAt.isNotBlank()) InfoRow("Atualizado em", formatDate(info.updatedAt))
        } else {
            // Sem metadados — mostra o que dá pra inferir do conteúdo baixado.
            when (content) {
                is FileContent.Unsupported -> {
                    InfoRow("Tipo de conteúdo", content.contentType)
                    InfoLine("Tamanho") { FileSizeText(content.size) }
                }
                is FileContent.Text -> InfoRow("Tipo de conteúdo", "texto")
            }
        }
    }
}

private val DATE_FORMATTER = DateTimeFormatter.ofPattern("dd/MM/yyyy HH:mm:ss")

/** Converte o ISO/RFC3339 do backend (ex.: 2026-06-02T16:11:16.582097-03:00) para dd/MM/yyyy HH:mm:ss. */
private fun formatDate(raw: String): String =
    runCatching { OffsetDateTime.parse(raw).format(DATE_FORMATTER) }.getOrDefault(raw)

@Composable
private fun InfoRow(label: String, value: String) {
    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        Text(label, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.primary)
        Text(value, style = MaterialTheme.typography.bodySmall, modifier = Modifier.weight(1f))
    }
}

@Composable
private fun InfoLine(label: String, value: @Composable () -> Unit) {
    Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        Text(label, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.primary)
        value()
    }
}

@Composable
private fun TextContent(content: FileContent.Text) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .glass(GlassLevel.Flat, radius = 12.dp)
            .padding(16.dp)
            .verticalScroll(rememberScrollState()),
    ) {
        if (content.truncated) {
            Text(
                "Arquivo grande — exibindo apenas o início.",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.primary,
            )
            Spacer(Modifier.height(8.dp))
        }
        Text(
            text = content.content,
            style = MaterialTheme.typography.bodySmall,
            fontFamily = FontFamily.Monospace,
        )
    }
}

@Composable
private fun UnsupportedContent(content: FileContent.Unsupported) {
    Column(
        modifier = Modifier.fillMaxSize(),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Icon(
            Icons.Default.InsertDriveFile,
            contentDescription = null,
            modifier = Modifier.size(48.dp),
            tint = MaterialTheme.colorScheme.primary,
        )
        Spacer(Modifier.height(12.dp))
        Text(
            "Pré-visualização não disponível para este tipo de arquivo.",
            style = MaterialTheme.typography.bodyMedium,
        )
        Spacer(Modifier.height(4.dp))
        Text("Use \"Baixar\" para salvar no dispositivo.", style = MaterialTheme.typography.bodySmall)
    }
}
