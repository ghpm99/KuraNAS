package com.kuranas.android.feature.ingest.data

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.core.paging.ListPage
import com.kuranas.android.core.paging.toListPage
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

class IngestRepository @Inject constructor(private val api: IngestApi) {

    suspend fun getTargets(): AppResult<List<IngestTargetDto>> = safeApiCall { api.getTargets() }

    suspend fun getPresets(): AppResult<List<IngestPresetDto>> = safeApiCall { api.getPresets() }

    suspend fun requestDownload(
        url: String,
        presetKey: String,
        targetRootPath: String,
        subfolder: String,
    ): AppResult<IngestFetchResponseDto> = safeApiCall {
        api.fetch(IngestFetchRequestDto(url = url, preset = presetKey, targetRoot = targetRootPath, subfolder = subfolder))
    }

    suspend fun getDownloads(page: Int): AppResult<ListPage<IngestDownloadDto>> = safeApiCall {
        api.getDownloads(page, DOWNLOADS_PAGE_SIZE).toListPage()
    }

    suspend fun cancelDownload(jobId: Int): AppResult<Unit> = safeApiCall { api.cancelJob(jobId) }

    suspend fun retryDownload(downloadId: Int): AppResult<Unit> = safeApiCall { api.retryDownload(downloadId) }

    private companion object {
        const val DOWNLOADS_PAGE_SIZE = 20
    }
}

@Module
@InstallIn(SingletonComponent::class)
object IngestModule {
    @Provides
    @Singleton
    fun provideIngestApi(retrofit: Retrofit): IngestApi = retrofit.create(IngestApi::class.java)
}
