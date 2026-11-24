package com.kuranas.android.feature.transfers.data.upload

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import okhttp3.MultipartBody
import okhttp3.RequestBody
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part
import retrofit2.http.Path

interface TransferUploadApi {
    @POST("api/v1/files/upload/init")
    suspend fun initUpload(@Body body: UploadInitRequest): Response<UploadInitResponse>

    @Multipart
    @POST("api/v1/files/upload/chunk")
    suspend fun sendChunk(
        @Part chunk: MultipartBody.Part,
        @Part("upload_id") uploadId: RequestBody,
        @Part("offset") offset: RequestBody,
    ): Response<UploadChunkResponse>

    @GET("api/v1/files/upload/{uploadId}")
    suspend fun getUploadStatus(@Path("uploadId") uploadId: String): Response<UploadStatusResponse>

    @POST("api/v1/files/upload/complete")
    suspend fun completeUpload(@Body body: UploadCompleteRequest): Response<UploadCompleteResponse>

    @DELETE("api/v1/files/upload/{uploadId}")
    suspend fun cancelUpload(@Path("uploadId") uploadId: String): Response<Unit>
}

@Serializable
data class UploadInitRequest(
    @SerialName("file_name") val fileName: String,
    val size: Long,
    @SerialName("target_folder_id") val targetFolderId: Long? = null,
    val checksum: String,
    @SerialName("modified_at") val modifiedAt: String? = null,
    @SerialName("on_conflict") val onConflict: String,
) {
    companion object {
        const val ON_CONFLICT_RENAME = "rename"
    }
}

@Serializable
data class UploadInitResponse(
    @SerialName("upload_id") val uploadId: String? = null,
    @SerialName("chunk_size") val chunkSize: Long = 0,
    @SerialName("received_size") val receivedSize: Long = 0,
    val resumed: Boolean = false,
    @SerialName("already_exists") val alreadyExists: Boolean = false,
)

@Serializable
data class UploadChunkResponse(@SerialName("received_size") val receivedSize: Long = 0)

@Serializable
data class UploadOffsetMismatchBody(
    val error: String? = null,
    @SerialName("received_size") val receivedSize: Long? = null,
)

@Serializable
data class UploadStatusResponse(
    @SerialName("upload_id") val uploadId: String = "",
    val size: Long = 0,
    @SerialName("received_size") val receivedSize: Long = 0,
)

@Serializable
data class UploadCompleteRequest(@SerialName("upload_id") val uploadId: String)

@Serializable
data class UploadCompleteResponse(
    val file: UploadCompletedFile? = null,
    @SerialName("job_id") val jobId: Int = 0,
)

@Serializable
data class UploadCompletedFile(
    val name: String = "",
    val path: String = "",
    val status: String = "",
)
