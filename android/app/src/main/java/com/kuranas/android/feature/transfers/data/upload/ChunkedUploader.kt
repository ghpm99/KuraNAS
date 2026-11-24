package com.kuranas.android.feature.transfers.data.upload

import com.kuranas.android.core.network.parseServerErrorMessage
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.toRequestBody
import retrofit2.Response
import java.io.IOException
import java.time.Instant
import javax.inject.Inject
import kotlin.math.min

class ChunkedUploader @Inject constructor(private val api: TransferUploadApi) {

    private val lenientJson = Json { ignoreUnknownKeys = true }

    /**
     * Envia [source] em partes, retomando a sessão de [request] quando houver `uploadId`.
     * [progressListener] recebe o `uploadId` assim que ele é conhecido e a cada parte aceita;
     * [chunkGate] roda antes de cada parte e pode interromper o envio lançando exceção de cancelamento.
     */
    suspend fun upload(
        source: UploadSource,
        request: UploadRequest,
        progressListener: UploadProgressListener,
        chunkGate: UploadChunkGate,
    ): UploadOutcome =
        try {
            UploadRun(source, request, progressListener, chunkGate).execute()
        } catch (exception: UploadSourceUnavailableException) {
            UploadOutcome.SourceMissing
        } catch (exception: IOException) {
            UploadOutcome.ConnectionLost
        }

    private inner class UploadRun(
        private val source: UploadSource,
        private val request: UploadRequest,
        private val progressListener: UploadProgressListener,
        private val chunkGate: UploadChunkGate,
    ) {
        private var checksum: String? = null

        suspend fun execute(): UploadOutcome {
            val totalBytes = source.sizeBytes()
            val established = request.uploadId?.let { resumeSession(it) } ?: initSession(totalBytes)
            val session = when (established) {
                is Established.Ready -> established.session
                is Established.Finished -> return established.outcome
            }
            return sendRemaining(totalBytes, session)
        }

        private suspend fun sendRemaining(totalBytes: Long, initialSession: Session): UploadOutcome {
            var session = initialSession
            var offset = session.receivedBytes
            var hasReinitialized = false
            var consecutiveRealignments = 0
            progressListener.onProgress(session.uploadId, offset)
            while (offset < totalBytes) {
                chunkGate.beforeChunk()
                val length = min(session.chunkSize, totalBytes - offset).toInt()
                val response = postChunk(session.uploadId, offset, source.readChunk(offset, length))
                when {
                    response.isSuccessful -> {
                        consecutiveRealignments = 0
                        offset = response.body()?.receivedSize ?: (offset + length)
                    }
                    response.code() == HTTP_CONFLICT -> {
                        val realignedOffset = offsetFromConflict(response) ?: return rejectionOf(response)
                        consecutiveRealignments++
                        if (consecutiveRealignments > MAX_CONSECUTIVE_REALIGNMENTS) return rejectionOf(response)
                        offset = realignedOffset
                    }
                    response.code() == HTTP_NOT_FOUND && !hasReinitialized -> {
                        hasReinitialized = true
                        val reinitialized = initSession(totalBytes)
                        session = (reinitialized as? Established.Ready)?.session
                            ?: return (reinitialized as Established.Finished).outcome
                        offset = session.receivedBytes
                    }
                    else -> return rejectionOf(response)
                }
                progressListener.onProgress(session.uploadId, offset)
            }
            return completeSession(session.uploadId)
        }

        private suspend fun resumeSession(uploadId: String): Established {
            val response = api.getUploadStatus(uploadId)
            if (response.isSuccessful) {
                val receivedBytes = response.body()?.receivedSize ?: 0
                return Established.Ready(Session(uploadId, DEFAULT_CHUNK_SIZE_BYTES, receivedBytes))
            }
            if (response.code() == HTTP_NOT_FOUND) return initSession(source.sizeBytes())
            return Established.Finished(rejectionOf(response))
        }

        private suspend fun initSession(totalBytes: Long): Established {
            val response = api.initUpload(
                UploadInitRequest(
                    fileName = request.displayName,
                    size = totalBytes,
                    targetFolderId = request.targetFolderId?.takeIf { it > 0 },
                    checksum = checksumOfSource(),
                    modifiedAt = source.modifiedAtMillis()?.let { Instant.ofEpochMilli(it).toString() },
                    onConflict = UploadInitRequest.ON_CONFLICT_RENAME,
                ),
            )
            if (!response.isSuccessful) return Established.Finished(rejectionOf(response))
            val body = response.body()
            if (body?.alreadyExists == true) return Established.Finished(UploadOutcome.AlreadyExisted)
            val uploadId = body?.uploadId?.takeIf { it.isNotBlank() }
                ?: return Established.Finished(UploadOutcome.Rejected(null))
            val chunkSize = body.chunkSize.takeIf { it > 0 } ?: DEFAULT_CHUNK_SIZE_BYTES
            return Established.Ready(Session(uploadId, chunkSize, body.receivedSize))
        }

        private suspend fun completeSession(uploadId: String): UploadOutcome {
            val response = api.completeUpload(UploadCompleteRequest(uploadId))
            return if (response.isSuccessful) UploadOutcome.Completed else rejectionOf(response)
        }

        private suspend fun checksumOfSource(): String =
            checksum ?: source.sha256Hex().also { checksum = it }

        private suspend fun postChunk(uploadId: String, offset: Long, bytes: ByteArray): Response<UploadChunkResponse> {
            val chunkPart = MultipartBody.Part.createFormData(
                CHUNK_PART_NAME,
                CHUNK_FILE_NAME,
                bytes.toRequestBody(OCTET_STREAM),
            )
            return api.sendChunk(
                chunk = chunkPart,
                uploadId = uploadId.toRequestBody(TEXT_PLAIN),
                offset = offset.toString().toRequestBody(TEXT_PLAIN),
            )
        }

        private fun offsetFromConflict(response: Response<*>): Long? {
            val rawBody = response.errorBody()?.string()?.takeIf { it.isNotBlank() } ?: return null
            return runCatching { lenientJson.decodeFromString<UploadOffsetMismatchBody>(rawBody).receivedSize }.getOrNull()
        }

        private fun isRetryable(code: Int): Boolean =
            code != HTTP_INSUFFICIENT_STORAGE &&
                (code >= HTTP_SERVER_ERROR_MIN || code == HTTP_REQUEST_TIMEOUT || code == HTTP_TOO_MANY_REQUESTS)

        private fun rejectionOf(response: Response<*>): UploadOutcome {
            val code = response.code()
            if (isRetryable(code)) {
                return UploadOutcome.ConnectionLost
            }
            return UploadOutcome.Rejected(parseServerErrorMessage(response.errorBody()?.string()))
        }
    }

    private data class Session(val uploadId: String, val chunkSize: Long, val receivedBytes: Long)

    private sealed interface Established {
        data class Ready(val session: Session) : Established
        data class Finished(val outcome: UploadOutcome) : Established
    }

    private companion object {
        const val DEFAULT_CHUNK_SIZE_BYTES = 2L * 1024 * 1024
        const val MAX_CONSECUTIVE_REALIGNMENTS = 3
        const val CHUNK_PART_NAME = "chunk"
        const val CHUNK_FILE_NAME = "chunk"
        const val HTTP_NOT_FOUND = 404
        const val HTTP_CONFLICT = 409
        const val HTTP_REQUEST_TIMEOUT = 408
        const val HTTP_TOO_MANY_REQUESTS = 429
        const val HTTP_SERVER_ERROR_MIN = 500
        const val HTTP_INSUFFICIENT_STORAGE = 507
        val OCTET_STREAM = "application/octet-stream".toMediaType()
        val TEXT_PLAIN = "text/plain".toMediaType()
    }
}
