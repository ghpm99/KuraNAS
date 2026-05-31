package com.kuranas.android.feature.files.data

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import dagger.Module
import dagger.Provides
import dagger.hilt.components.SingletonComponent
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

class FilesRepository @Inject constructor(private val api: FilesApi) {

    suspend fun getRootFiles(): AppResult<List<FileItemDto>> = safeApiCall {
        api.getRootFiles().items
    }

    suspend fun getChildrenById(id: String): AppResult<List<FileItemDto>> = safeApiCall {
        api.getChildrenById(id).items
    }

    suspend fun getImages(): AppResult<List<FileItemDto>> = safeApiCall {
        api.getImages().items
    }

    suspend fun createFolder(name: String, parentId: String?): AppResult<FileItemDto> = safeApiCall {
        api.createFolder(CreateFolderRequest(name, parentId))
    }

    suspend fun renameFile(id: String, name: String): AppResult<FileItemDto> = safeApiCall {
        api.renameFile(RenameRequest(id, name))
    }

    suspend fun deleteFile(path: String): AppResult<Unit> = safeApiCall {
        api.deleteFile(path)
    }

    suspend fun starFile(id: String): AppResult<Unit> = safeApiCall {
        api.starFile(id)
    }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object FilesModule {
    @Provides
    @Singleton
    fun provideFilesApi(retrofit: Retrofit): FilesApi = retrofit.create(FilesApi::class.java)
}
