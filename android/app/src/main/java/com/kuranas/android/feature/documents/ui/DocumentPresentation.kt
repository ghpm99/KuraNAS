package com.kuranas.android.feature.documents.ui

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Article
import androidx.compose.material.icons.filled.Description
import androidx.compose.material.icons.filled.PictureAsPdf
import androidx.compose.material.icons.filled.TableChart
import androidx.compose.ui.graphics.vector.ImageVector
import java.time.OffsetDateTime
import java.time.format.DateTimeFormatter

private val MODIFIED_DATE_FORMATTER = DateTimeFormatter.ofPattern("dd/MM/yyyy HH:mm")
private val SPREADSHEET_FORMATS = setOf(".xls", ".xlsx", ".ods", ".csv")
private val PLAIN_TEXT_FORMATS = setOf(".txt", ".md", ".rtf")

internal fun documentFormatIcon(format: String): ImageVector = when (format.lowercase()) {
    ".pdf" -> Icons.Default.PictureAsPdf
    in SPREADSHEET_FORMATS -> Icons.Default.TableChart
    in PLAIN_TEXT_FORMATS -> Icons.Default.Article
    else -> Icons.Default.Description
}

internal fun formatDocumentDate(rawTimestamp: String): String =
    runCatching { OffsetDateTime.parse(rawTimestamp).format(MODIFIED_DATE_FORMATTER) }.getOrDefault(rawTimestamp)
