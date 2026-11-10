package com.kuranas.android.feature.music.data

import com.kuranas.android.core.network.AppResult
import com.kuranas.android.feature.files.data.FileStarStore
import com.kuranas.android.feature.files.data.FilesRepository
import kotlinx.coroutines.flow.StateFlow
import javax.inject.Inject

class TrackStarToggler @Inject constructor(
    private val filesRepository: FilesRepository,
    private val starStore: FileStarStore,
) {
    val starredById: StateFlow<Map<String, Boolean>> = starStore.starredById

    suspend fun toggle(track: TrackDto): String? {
        val fileId = track.id.toString()
        val previousIsStarred = starredById.value[fileId] ?: track.starred
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
