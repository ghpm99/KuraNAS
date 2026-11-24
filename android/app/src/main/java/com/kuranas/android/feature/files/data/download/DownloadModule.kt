package com.kuranas.android.feature.files.data.download

import com.kuranas.android.feature.files.data.FilesApi
import dagger.Binds
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import com.kuranas.android.feature.transfers.data.TransferClock
import com.kuranas.android.feature.transfers.data.TransferDao
import com.kuranas.android.feature.transfers.data.download.PersistentDownloadQueue
import com.kuranas.android.feature.transfers.data.upload.TransferWorkScheduler
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object DownloadModule {

    @Provides
    @Singleton
    fun provideStreamingDownloader(api: FilesApi, mimeTypeResolver: MimeTypeResolver): StreamingDownloader =
        StreamingDownloader(api, mimeTypeResolver)

    @Provides
    @Singleton
    fun provideDownloadQueue(
        transferDao: TransferDao,
        scheduler: TransferWorkScheduler,
        clock: TransferClock,
    ): DownloadQueue = PersistentDownloadQueue(
        scope = CoroutineScope(SupervisorJob() + Dispatchers.Default),
        transferDao = transferDao,
        scheduler = scheduler,
        clock = clock,
    )
}

@Module
@InstallIn(SingletonComponent::class)
interface DownloadBindingsModule {

    @Binds
    fun bindDownloadExecutor(executor: StreamingDownloadExecutor): DownloadExecutor

    @Binds
    fun bindMimeTypeResolver(resolver: AndroidMimeTypeResolver): MimeTypeResolver

    @Binds
    fun bindDownloadNotifier(notifier: SystemDownloadNotifier): DownloadNotifier
}
