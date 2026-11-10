package com.kuranas.android.feature.video.ui

import com.kuranas.android.feature.files.ui.FileStarState
import com.kuranas.android.feature.video.data.VideoItemDto

fun FileStarState.toggle(video: VideoItemDto) = toggle(video.id, video.starred)
