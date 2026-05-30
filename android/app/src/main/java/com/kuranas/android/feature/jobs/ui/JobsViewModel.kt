package com.kuranas.android.feature.jobs.ui

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.kuranas.android.core.network.AppResult
import com.kuranas.android.core.network.safeApiCall
import com.kuranas.android.feature.jobs.data.JobDto
import com.kuranas.android.feature.jobs.data.JobsApi
import dagger.Module
import dagger.Provides
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.components.SingletonComponent
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import retrofit2.Retrofit
import javax.inject.Inject
import javax.inject.Singleton

data class JobsUiState(
    val isLoading: Boolean = true,
    val jobs: List<JobDto> = emptyList(),
    val error: String? = null,
)

@HiltViewModel
class JobsViewModel @Inject constructor(private val api: JobsApi) : ViewModel() {

    private val _state = MutableStateFlow(JobsUiState())
    val state: StateFlow<JobsUiState> = _state.asStateFlow()

    init {
        load()
        startPolling()
    }

    fun load() {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true) }
            when (val r = safeApiCall { api.listJobs() }) {
                is AppResult.Success -> _state.update { it.copy(isLoading = false, jobs = r.data) }
                is AppResult.Error -> _state.update { it.copy(isLoading = false, error = r.message) }
            }
        }
    }

    fun cancelJob(id: String) {
        viewModelScope.launch {
            safeApiCall { api.cancelJob(id) }
            load()
        }
    }

    private fun startPolling() {
        viewModelScope.launch {
            while (true) {
                delay(5_000)
                val hasActive = _state.value.jobs.any { it.status in listOf("running", "pending") }
                if (hasActive) load()
            }
        }
    }
}

@Module
@dagger.hilt.InstallIn(SingletonComponent::class)
object JobsModule {
    @Provides
    @Singleton
    fun provideJobsApi(retrofit: Retrofit): JobsApi = retrofit.create(JobsApi::class.java)
}
