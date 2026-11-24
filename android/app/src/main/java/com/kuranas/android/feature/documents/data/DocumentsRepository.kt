package com.kuranas.android.feature.documents.data

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

class DocumentsRepository @Inject constructor(private val api: DocumentsApi) {

    suspend fun getDocuments(page: Int, format: String? = null, folderId: Int? = null): AppResult<ListPage<DocumentDto>> =
        safeApiCall { api.getDocuments(page, DOCUMENTS_PAGE_SIZE, format, folderId).toListPage() }

    suspend fun getDocumentsCount(format: String? = null, folderId: Int? = null): AppResult<Int> =
        safeApiCall { api.getDocumentsCount(format, folderId).total }

    suspend fun searchDocuments(query: String, page: Int, format: String? = null): AppResult<ListPage<DocumentSearchResultDto>> =
        safeApiCall { api.searchDocuments(query, page, DOCUMENTS_PAGE_SIZE, format).toListPage() }

    private companion object {
        const val DOCUMENTS_PAGE_SIZE = 50
    }
}

@Module
@InstallIn(SingletonComponent::class)
object DocumentsModule {
    @Provides
    @Singleton
    fun provideDocumentsApi(retrofit: Retrofit): DocumentsApi = retrofit.create(DocumentsApi::class.java)
}
