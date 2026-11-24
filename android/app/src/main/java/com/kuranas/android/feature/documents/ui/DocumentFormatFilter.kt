package com.kuranas.android.feature.documents.ui

enum class DocumentFormatFilter(val apiValue: String?) {
    ALL(null),
    PDF("pdf"),
    TEXT("text"),
    SPREADSHEET("spreadsheet"),
    PRESENTATION("presentation"),
    OTHER("other"),
}
