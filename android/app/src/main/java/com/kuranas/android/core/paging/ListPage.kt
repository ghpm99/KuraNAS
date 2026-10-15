package com.kuranas.android.core.paging

import com.kuranas.android.core.network.PageDto

const val FIRST_PAGE = 1

data class ListPage<T>(val items: List<T>, val hasNext: Boolean)

fun <T> PageDto<T>.toListPage() = ListPage(items, pagination.hasNext)
