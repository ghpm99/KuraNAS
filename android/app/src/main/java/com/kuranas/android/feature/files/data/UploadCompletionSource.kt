package com.kuranas.android.feature.files.data

import com.kuranas.android.feature.transfers.data.TransferDao
import com.kuranas.android.feature.transfers.data.TransferEntity
import kotlinx.coroutines.flow.Flow
import dagger.Binds
import dagger.Module
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Inject

interface UploadCompletionSource {
    fun observeTransfers(): Flow<List<TransferEntity>>
}

class RoomUploadCompletionSource @Inject constructor(
    private val transferDao: TransferDao,
) : UploadCompletionSource {
    override fun observeTransfers(): Flow<List<TransferEntity>> = transferDao.observeAll()
}

@Module
@InstallIn(SingletonComponent::class)
interface UploadCompletionBindingsModule {

    @Binds
    fun bindUploadCompletionSource(source: RoomUploadCompletionSource): UploadCompletionSource
}
