package com.kuranas.android.feature.transfers.data.upload

import android.content.ContentResolver
import androidx.core.net.toUri
import com.kuranas.android.feature.transfers.data.TransferEntity
import java.io.File
import javax.inject.Inject

class UploadSourceFactory @Inject constructor(private val contentResolver: ContentResolver) {

    fun open(transfer: TransferEntity): UploadSource {
        transfer.stagedFilePath?.let { return StagedFileUploadSource(File(it)) }
        val sourceUri = transfer.sourceUri ?: throw UploadSourceUnavailableException()
        return ContentUriUploadSource(contentResolver, sourceUri.toUri())
    }
}
