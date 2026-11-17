package com.kuranas.android.feature.files.ui

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FilesRepository
import com.kuranas.android.feature.files.data.SelectedFile
import com.kuranas.android.feature.files.data.UploadedFileDto
import javax.inject.Inject

private const val STATUS_UPLOADED = "uploaded"
private const val STATUS_REPLACED = "replaced"
private const val STATUS_RENAMED = "renamed"
private const val STATUS_SKIPPED = "skipped"
private const val HTTP_FORBIDDEN = 403

/**
 * Envia os arquivos um por requisição, em sequência, de modo que a falha de um não derruba os outros.
 * Uma falha de conexão ou de whitelist vale para o servidor inteiro: os arquivos restantes não são tentados
 * e entram nas falhas com o mesmo motivo.
 */
class UploadBatchRunner @Inject constructor(private val repository: FilesRepository) {

    suspend fun upload(
        files: List<SelectedFile>,
        targetFolderId: String?,
        onFileStarted: (current: Int, total: Int, fileName: String) -> Unit,
    ): UploadSummary {
        var summary = UploadSummary()
        var serverWideFailure: UploadFailureReason? = null
        files.forEachIndexed { index, file ->
            serverWideFailure?.let { reason ->
                summary = summary.withFailure(file, reason)
                return@forEachIndexed
            }
            onFileStarted(index + 1, files.size, file.name)
            when (val response = repository.uploadFile(file, targetFolderId)) {
                is AppResult.Success -> summary = summary.withOutcome(file, response.data.files.firstOrNull(), response.data.message)
                is AppResult.Error -> {
                    val reason = failureReasonFor(response)
                    if (isServerWide(response)) serverWideFailure = reason
                    summary = summary.withFailure(file, reason)
                }
            }
        }
        return summary
    }

    private fun isServerWide(error: AppResult.Error): Boolean = error.code == null || error.code == HTTP_FORBIDDEN

    private fun failureReasonFor(error: AppResult.Error): UploadFailureReason =
        if (error.code == null) UploadFailureReason.NoServer else UploadFailureReason.ServerMessage(error.message)

    private fun UploadSummary.withFailure(file: SelectedFile, reason: UploadFailureReason) =
        copy(failures = failures + UploadFailure(file, reason))

    private fun UploadSummary.withOutcome(file: SelectedFile, outcome: UploadedFileDto?, responseMessage: String) =
        when (outcome?.status) {
            STATUS_UPLOADED, STATUS_REPLACED -> copy(uploadedCount = uploadedCount + 1)
            STATUS_RENAMED -> copy(renamedCount = renamedCount + 1)
            STATUS_SKIPPED -> copy(skippedCount = skippedCount + 1)
            null -> withFailure(file, UploadFailureReason.ServerMessage(responseMessage))
            else -> withFailure(file, UploadFailureReason.ServerMessage(outcome.error.ifBlank { responseMessage }))
        }
}
