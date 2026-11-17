package com.kuranas.android.feature.files.data

/** Pasta de destino de um envio. [folderId] nulo é a raiz do servidor; [folderName] é vazio na raiz. */
data class UploadDestination(val folderId: String?, val folderName: String) {
    val isRoot: Boolean get() = folderId == null

    companion object {
        val Root = UploadDestination(folderId = null, folderName = "")
    }
}
