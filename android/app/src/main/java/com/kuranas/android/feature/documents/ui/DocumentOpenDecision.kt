package com.kuranas.android.feature.documents.ui

enum class DocumentOpenMode { IN_APP_VIEWER, OPEN_WITH }

private val IN_APP_VIEWER_FORMATS = setOf(".txt", ".md")

internal fun documentOpenMode(format: String): DocumentOpenMode =
    if (format.lowercase() in IN_APP_VIEWER_FORMATS) DocumentOpenMode.IN_APP_VIEWER else DocumentOpenMode.OPEN_WITH
