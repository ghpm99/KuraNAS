package com.kuranas.android.feature.cameraupload.data

import com.kuranas.android.feature.transfers.data.KuranasDatabase
import dagger.Binds
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent

@Module
@InstallIn(SingletonComponent::class)
abstract class CameraUploadBindingsModule {

    @Binds
    abstract fun bindServerReachability(reachability: HealthEndpointReachability): ServerReachability
}

@Module
@InstallIn(SingletonComponent::class)
object CameraUploadDatabaseModule {

    @Provides
    fun provideCameraUploadDao(database: KuranasDatabase): CameraUploadDao = database.cameraUploadDao()
}
