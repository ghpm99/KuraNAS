package com.kuranas.android.feature.transfers.data.notification

import dagger.Binds
import dagger.Module
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent

@Module
@InstallIn(SingletonComponent::class)
interface TransferNotificationModule {

    @Binds
    fun bindTransferBatchNotifier(notifier: SystemTransferBatchNotifier): TransferBatchNotifier
}
