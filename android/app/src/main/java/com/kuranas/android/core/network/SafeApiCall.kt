package com.kuranas.android.core.network

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import retrofit2.HttpException
import java.io.IOException

suspend fun <T> safeApiCall(call: suspend () -> T): AppResult<T> =
    withContext(Dispatchers.IO) {
        try {
            AppResult.Success(call())
        } catch (e: HttpException) {
            AppResult.Error(
                message = serverErrorMessage(e) ?: "Erro HTTP ${e.code()}",
                code = e.code(),
            )
        } catch (e: IOException) {
            AppResult.Error("Sem conexão com o servidor")
        } catch (e: Exception) {
            AppResult.Error(e.message ?: "Erro desconhecido")
        }
    }

private fun serverErrorMessage(exception: HttpException): String? {
    val body = exception.response()?.errorBody()?.string()?.takeIf { it.isNotBlank() } ?: return null
    val serverMessage = runCatching {
        Json.parseToJsonElement(body).jsonObject["error"]?.jsonPrimitive?.contentOrNull
    }.getOrNull()
    return serverMessage ?: body
}
