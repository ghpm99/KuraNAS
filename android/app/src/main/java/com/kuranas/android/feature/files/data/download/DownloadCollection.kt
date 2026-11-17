package com.kuranas.android.feature.files.data.download

private const val APP_FOLDER_NAME = "KuraNAS"

enum class DownloadCollection(val relativeDirectory: String) {
    IMAGES("Pictures/$APP_FOLDER_NAME"),
    VIDEOS("Movies/$APP_FOLDER_NAME"),
    DOWNLOADS("Download"),
}

fun downloadCollectionFor(mimeType: String): DownloadCollection = when {
    mimeType.startsWith("image/") -> DownloadCollection.IMAGES
    mimeType.startsWith("video/") -> DownloadCollection.VIDEOS
    else -> DownloadCollection.DOWNLOADS
}
