package com.kuranas.android.feature.diary.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.feature.diary.data.CreateDiaryRequest
import com.kuranas.android.feature.diary.data.DiaryApi
import com.kuranas.android.feature.diary.data.DiaryEntryDto
import dagger.Module
import dagger.Provides
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.components.SingletonComponent
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

data class DiaryUiState(
    val isLoading: Boolean = true,
    val entries: List<DiaryEntryDto> = emptyList(),
    val error: String? = null,
    val showCreateDialog: Boolean = false,
    val editingEntry: DiaryEntryDto? = null,
)

@HiltViewModel
class DiaryViewModel @Inject constructor(private val api: DiaryApi) : ViewModel() {

    private val _state = MutableStateFlow(DiaryUiState())
    val state: StateFlow<DiaryUiState> = _state.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true, error = null) }
            when (val r = safeApiCall { api.getDiary() }) {
                is AppResult.Success -> _state.update { it.copy(isLoading = false, entries = r.data.items) }
                is AppResult.Error -> _state.update { it.copy(isLoading = false, error = r.message) }
            }
        }
    }

    fun createEntry(title: String, content: String) {
        viewModelScope.launch {
            safeApiCall { api.createEntry(CreateDiaryRequest(name = title, description = content)) }
            _state.update { it.copy(showCreateDialog = false) }
            load()
        }
    }

    // O backend só persiste o nome da atividade neste endpoint (campo de formulário `data`).
    fun updateEntry(id: Int, title: String, content: String) {
        viewModelScope.launch {
            safeApiCall { api.updateEntry(id, title) }
            _state.update { it.copy(editingEntry = null) }
            load()
        }
    }

    fun toggleCreateDialog(show: Boolean) { _state.update { it.copy(showCreateDialog = show) } }
    fun setEditingEntry(entry: DiaryEntryDto?) { _state.update { it.copy(editingEntry = entry) } }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object DiaryModule {
    @Provides
    @Singleton
    fun provideDiaryApi(retrofit: Retrofit): DiaryApi = retrofit.create(DiaryApi::class.java)
}
