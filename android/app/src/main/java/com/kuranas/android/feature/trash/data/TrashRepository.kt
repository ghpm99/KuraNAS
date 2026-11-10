package com.kuranas.android.feature.trash.data

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

class TrashRepository @Inject constructor(private val api: TrashApi) {

    suspend fun getTrashItems(page: Int): AppResult<ListPage<TrashItemDto>> = safeApiCall {
        api.getTrashItems(page, TRASH_PAGE_SIZE).toListPage()
    }

    suspend fun restoreItem(itemId: Int): AppResult<Unit> = safeApiCall { api.restoreItem(itemId) }

    suspend fun deleteItemForever(itemId: Int): AppResult<Unit> = safeApiCall { api.deleteItemForever(itemId) }

    suspend fun emptyTrash(): AppResult<Unit> = safeApiCall { api.emptyTrash() }

    private companion object {
        const val TRASH_PAGE_SIZE = 50
    }
}

@Module
@InstallIn(SingletonComponent::class)
object TrashModule {
    @Provides
    @Singleton
    fun provideTrashApi(retrofit: Retrofit): TrashApi = retrofit.create(TrashApi::class.java)
}
