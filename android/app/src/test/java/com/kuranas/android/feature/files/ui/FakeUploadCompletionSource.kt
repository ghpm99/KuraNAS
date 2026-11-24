package com.kuranas.android.feature.files.ui

import com.kuranas.android.feature.files.data.UploadCompletionSource
import com.kuranas.android.feature.transfers.data.TransferEntity
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow

class FakeUploadCompletionSource : UploadCompletionSource {
    val transfers = MutableStateFlow<List<TransferEntity>>(emptyList())

    override fun observeTransfers(): Flow<List<TransferEntity>> = transfers
}
