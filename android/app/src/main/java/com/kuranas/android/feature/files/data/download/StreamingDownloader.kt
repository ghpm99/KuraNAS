package com.kuranas.android.feature.files.data.download

import com.kuranas.android.core.network.parseServerErrorMessage
import com.kuranas.android.feature.files.data.FilesApi
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.currentCoroutineContext
import kotlinx.coroutines.delay
import kotlinx.coroutines.ensureActive
import kotlinx.coroutines.withContext
import okhttp3.ResponseBody
import retrofit2.Response
import java.io.IOException
import java.io.OutputStream

/**
 * Copia `GET /files/download/{id}` em blocos para um [PendingDownload], sem nunca carregar o arquivo na memória.
 * O destino só é criado na primeira resposta bem-sucedida, já com o MIME resolvido. Quedas de conexão retomam
 * com `Range` de onde parou; o download falha após [MAX_FRUITLESS_ATTEMPTS] quedas seguidas sem nenhum byte novo.
 * Em qualquer falha definitiva o destino é descartado, sem deixar arquivo parcial.
 */
class StreamingDownloader(
    private val api: FilesApi,
    private val mimeTypeResolver: MimeTypeResolver,
    private val retryDelayMillis: Long = DEFAULT_RETRY_DELAY_MILLIS,
) {

    suspend fun download(
        request: DownloadRequest,
        destinations: PendingDownloadFactory,
        onProgress: (DownloadProgress) -> Unit,
    ): DownloadOutcome = withContext(Dispatchers.IO) {
        val session = DownloadSession(request, destinations)
        try {
            transfer(session, onProgress)
        } catch (cancellation: CancellationException) {
            session.destination?.discard()
            throw cancellation
        }
    }

    private suspend fun transfer(session: DownloadSession, onProgress: (DownloadProgress) -> Unit): DownloadOutcome {
        var fruitlessAttempts = 0
        var furthestByte = 0L
        while (true) {
            when (val attempt = runAttempt(session, onProgress)) {
                AttemptOutcome.Finished -> return commit(session)
                is AttemptOutcome.Rejected -> return fail(session, attempt.failure)
                AttemptOutcome.Interrupted -> {
                    val madeProgress = session.receivedBytes > furthestByte
                    furthestByte = maxOf(furthestByte, session.receivedBytes)
                    fruitlessAttempts = if (madeProgress) 0 else fruitlessAttempts + 1
                    if (fruitlessAttempts >= MAX_FRUITLESS_ATTEMPTS) return fail(session, DownloadFailure.ConnectionLost)
                    delay(retryDelayMillis)
                }
            }
        }
    }

    private fun commit(session: DownloadSession): DownloadOutcome {
        val destination = session.destination ?: return fail(session, DownloadFailure.StorageUnavailable)
        return try {
            DownloadOutcome.Completed(destination.commit(), session.mimeType)
        } catch (failure: IOException) {
            fail(session, DownloadFailure.StorageUnavailable)
        }
    }

    private fun fail(session: DownloadSession, failure: DownloadFailure): DownloadOutcome {
        session.destination?.discard()
        return DownloadOutcome.Failed(failure)
    }

    private suspend fun runAttempt(session: DownloadSession, onProgress: (DownloadProgress) -> Unit): AttemptOutcome {
        val response = try {
            api.downloadFile(session.request.fileId, rangeHeader(session.receivedBytes))
        } catch (failure: IOException) {
            return AttemptOutcome.Interrupted
        }
        if (!response.isSuccessful) return AttemptOutcome.Rejected(serverFailure(response))
        val body = response.body() ?: return AttemptOutcome.Interrupted
        return body.use { openBody ->
            val resume = resolveResume(response, session.receivedBytes) ?: return@use AttemptOutcome.Interrupted
            if (!resume.isContinuing) session.receivedBytes = 0
            val destination = session.destination ?: createDestination(session, openBody)
            val totalBytes = resume.totalBytes ?: openBody.contentLength().takeIf { it > 0 } ?: session.request.sizeBytes
            copyBody(openBody, destination, session, resume.isContinuing, totalBytes, onProgress)
        }
    }

    private fun createDestination(session: DownloadSession, body: ResponseBody): PendingDownload {
        session.mimeType = resolveDownloadMimeType(session.request.fileName, body.contentType()?.toString(), mimeTypeResolver)
        return session.destinations.create(session.mimeType).also { session.destination = it }
    }

    private fun rangeHeader(receivedBytes: Long): String? =
        if (receivedBytes > 0) "bytes=$receivedBytes-" else null

    private fun serverFailure(response: Response<ResponseBody>): DownloadFailure.Server {
        val rawBody = response.errorBody()?.use { it.string() }
        return DownloadFailure.Server(parseServerErrorMessage(rawBody), response.code())
    }

    private fun resolveResume(response: Response<ResponseBody>, receivedBytes: Long): ResumePoint? {
        if (response.code() != HTTP_PARTIAL_CONTENT) return ResumePoint(isContinuing = false, totalBytes = null)
        val range = parseContentRange(response.headers()["Content-Range"]) ?: return null
        if (range.firstByte != receivedBytes) return null
        return ResumePoint(isContinuing = true, totalBytes = range.totalBytes)
    }

    private suspend fun copyBody(
        body: ResponseBody,
        destination: PendingDownload,
        session: DownloadSession,
        shouldAppend: Boolean,
        totalBytes: Long,
        onProgress: (DownloadProgress) -> Unit,
    ): AttemptOutcome {
        val output = try {
            destination.openOutput(shouldAppend)
        } catch (failure: IOException) {
            return AttemptOutcome.Rejected(DownloadFailure.StorageUnavailable)
        }
        try {
            val outcome = copyChunks(body, output, session, totalBytes, onProgress)
            if (outcome == AttemptOutcome.Finished) output.flush()
            return outcome
        } catch (failure: IOException) {
            return AttemptOutcome.Rejected(DownloadFailure.StorageUnavailable)
        } finally {
            runCatching { output.close() }
        }
    }

    private suspend fun copyChunks(
        body: ResponseBody,
        output: OutputStream,
        session: DownloadSession,
        totalBytes: Long,
        onProgress: (DownloadProgress) -> Unit,
    ): AttemptOutcome {
        val source = body.source()
        val chunk = ByteArray(COPY_CHUNK_BYTES)
        while (true) {
            currentCoroutineContext().ensureActive()
            val readCount = try {
                source.read(chunk)
            } catch (failure: IOException) {
                return AttemptOutcome.Interrupted
            }
            if (readCount == END_OF_STREAM) break
            output.write(chunk, 0, readCount)
            session.receivedBytes += readCount
            onProgress(DownloadProgress(session.receivedBytes, totalBytes))
        }
        val isTruncated = totalBytes > 0 && session.receivedBytes < totalBytes
        return if (isTruncated) AttemptOutcome.Interrupted else AttemptOutcome.Finished
    }

    private class DownloadSession(val request: DownloadRequest, val destinations: PendingDownloadFactory) {
        var receivedBytes: Long = 0
        var mimeType: String = UNKNOWN_MIME_TYPE
        var destination: PendingDownload? = null
    }

    private class ResumePoint(val isContinuing: Boolean, val totalBytes: Long?)

    private sealed interface AttemptOutcome {
        data object Finished : AttemptOutcome
        data object Interrupted : AttemptOutcome
        data class Rejected(val failure: DownloadFailure) : AttemptOutcome
    }

    companion object {
        const val MAX_FRUITLESS_ATTEMPTS = 3
        const val DEFAULT_RETRY_DELAY_MILLIS = 2_000L
        private const val COPY_CHUNK_BYTES = 64 * 1024
        private const val HTTP_PARTIAL_CONTENT = 206
        private const val END_OF_STREAM = -1
    }
}
