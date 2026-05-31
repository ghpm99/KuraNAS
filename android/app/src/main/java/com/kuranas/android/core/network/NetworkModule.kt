package com.kuranas.android.core.network

import com.jakewharton.retrofit2.converter.kotlinx.serialization.asConverterFactory
import com.kuranas.android.BuildConfig
import com.kuranas.android.core.server.ServerStore
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.json.Json
import okhttp3.HttpUrl
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import okhttp3.Interceptor
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import java.util.concurrent.TimeUnit
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {

    private const val DEFAULT_PORT = 8000

    /**
     * Converte o que o usuário digitou (ex.: "192.168.18.7:8000", "192.168.18.7",
     * "http://host:8000/") numa [HttpUrl] válida. Retorna null em vez de lançar
     * quando o valor é inválido — o interceptor NÃO pode lançar (derruba o app).
     */
    private fun parseServerUrl(raw: String): HttpUrl? {
        val trimmed = raw.trim().trimEnd('/')
        if (trimmed.isEmpty()) return null
        val withScheme = if (trimmed.contains("://")) trimmed else "http://$trimmed"
        val parsed = withScheme.toHttpUrlOrNull() ?: return null
        // Sem porta explícita → assume a porta padrão do servidor (8000), não a 80.
        val authority = withScheme.substringAfter("://").substringBefore("/")
        val hasExplicitPort = authority.substringAfterLast(']').contains(":")
        return if (hasExplicitPort) parsed else parsed.newBuilder().port(DEFAULT_PORT).build()
    }

    @Provides
    @Singleton
    fun provideJson(): Json = Json {
        ignoreUnknownKeys = true
        coerceInputValues = true
        explicitNulls = false
    }

    @Provides
    @Singleton
    fun provideLoggingInterceptor(): HttpLoggingInterceptor =
        HttpLoggingInterceptor().apply {
            level = if (BuildConfig.DEBUG) HttpLoggingInterceptor.Level.BODY
            else HttpLoggingInterceptor.Level.NONE
        }

    @Provides
    @Singleton
    fun provideServerInterceptor(serverStore: ServerStore): Interceptor =
        Interceptor { chain ->
            val original: Request = chain.request()
            val base = runBlocking { serverStore.serverUrl.first() }?.let { parseServerUrl(it) }
            if (base == null) {
                // Sem servidor configurado (ou valor inválido): segue sem reescrever.
                // Não lançar daqui — exceção no interceptor derruba o app.
                chain.proceed(original)
            } else {
                val newUrl = original.url.newBuilder()
                    .scheme(base.scheme)
                    .host(base.host)
                    .port(base.port)
                    .build()
                chain.proceed(original.newBuilder().url(newUrl).build())
            }
        }

    @Provides
    @Singleton
    fun provideOkHttpClient(
        logging: HttpLoggingInterceptor,
        serverInterceptor: Interceptor,
    ): OkHttpClient =
        OkHttpClient.Builder()
            .addInterceptor(serverInterceptor)
            .addInterceptor(logging)
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(60, TimeUnit.SECONDS)
            .writeTimeout(60, TimeUnit.SECONDS)
            .build()

    @Provides
    @Singleton
    fun provideRetrofit(client: OkHttpClient, json: Json): Retrofit {
        val contentType = "application/json".toMediaType()
        return Retrofit.Builder()
            .baseUrl("http://localhost:8000/")
            .client(client)
            .addConverterFactory(json.asConverterFactory(contentType))
            .build()
    }
}
