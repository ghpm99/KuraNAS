package com.kuranas.android.feature.files.data

import com.kuranas.android.core.network.AppResult
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject

class FileStarToggler @Inject constructor(
    private val filesRepository: FilesRepository,
    private val starStore: FileStarStore,
) {
    val starredById: StateFlow<Map<String, Boolean>> = starStore.starredById

    suspend fun toggle(fileId: String, isStarredFallback: Boolean): String? {
        val previousIsStarred = starredById.value[fileId] ?: isStarredFallback
        starStore.set(fileId, !previousIsStarred)
        return when (val response = filesRepository.setFileStarred(fileId, !previousIsStarred)) {
            is AppResult.Success -> {
                starStore.set(fileId, response.data.starred)
                null
            }
            is AppResult.Error -> {
                starStore.set(fileId, previousIsStarred)
                response.message
            }
        }
    }
}
