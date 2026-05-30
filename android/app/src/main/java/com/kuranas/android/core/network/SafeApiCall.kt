package com.kuranas.android.core.network

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import retrofit2.HttpException
import java.io.IOException

suspend fun <T> safeApiCall(call: suspend () -> T): AppResult<T> =
    withContext(Dispatchers.IO) {
        try {
            AppResult.Success(call())
        } catch (e: HttpException) {
            AppResult.Error(
                message = e.response()?.errorBody()?.string() ?: "Erro HTTP ${e.code()}",
                code = e.code(),
            )
        } catch (e: IOException) {
            AppResult.Error("Sem conexão com o servidor")
        } catch (e: Exception) {
            AppResult.Error(e.message ?: "Erro desconhecido")
        }
    }
