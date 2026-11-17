package com.kuranas.android.feature.files.data.download

import com.kuranas.android.feature.files.data.FilesApi
import dagger.Binds
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
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
    fun provideDownloadQueue(executor: DownloadExecutor, notifier: DownloadNotifier): DownloadQueue =
        InMemoryDownloadQueue(CoroutineScope(SupervisorJob() + Dispatchers.Default), executor, notifier)
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
