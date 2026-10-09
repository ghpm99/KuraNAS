package com.kuranas.android.feature.music.ui

import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember

private const val PREFETCH_DISTANCE_ITEMS = 10

@Composable
fun LoadMoreOnNearEnd(listState: LazyListState, itemCount: Int, hasMore: Boolean, onLoadMore: () -> Unit) {
    val isNearEnd by remember(listState) {
        derivedStateOf {
            val lastVisibleIndex = listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            lastVisibleIndex >= listState.layoutInfo.totalItemsCount - PREFETCH_DISTANCE_ITEMS
        }
    }
    LaunchedEffect(isNearEnd, itemCount, hasMore) {
        if (isNearEnd) onLoadMore()
    }
}
