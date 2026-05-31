package com.kuranas.android.core.network

/**
 * O backend Go expõe apenas a extensão do arquivo (files.FileDto.format, ex.: ".mp3"),
 * não um mime-type. A UI decide ícone/ação por prefixo ("image/", "video/", "audio/"),
 * então mapeamos a extensão para um mime-type aproximado num único lugar.
 */
fun mimeTypeForFormat(format: String): String {
    val ext = format.removePrefix(".").lowercase()
    return when (ext) {
        "jpg", "jpeg", "png", "gif", "webp", "bmp", "heic", "heif", "svg", "tiff" -> "image/$ext"
        "mp4", "mkv", "avi", "mov", "webm", "flv", "wmv", "m4v", "mpeg", "mpg" -> "video/$ext"
        "mp3", "flac", "wav", "aac", "ogg", "m4a", "wma", "opus" -> "audio/$ext"
        "" -> "application/octet-stream"
        else -> "application/$ext"
    }
}
