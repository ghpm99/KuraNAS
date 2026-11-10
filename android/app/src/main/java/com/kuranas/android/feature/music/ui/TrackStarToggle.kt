package com.kuranas.android.feature.music.ui

import com.kuranas.android.feature.files.ui.FileStarState
import com.kuranas.android.feature.music.data.TrackDto

fun FileStarState.toggle(track: TrackDto) = toggle(track.id.toString(), track.starred)
