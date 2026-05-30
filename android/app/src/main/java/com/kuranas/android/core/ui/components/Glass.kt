package com.kuranas.android.core.ui.components

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.kuranas.android.ui.theme.GlassFlatBorder
import com.kuranas.android.ui.theme.GlassFlatFill
import com.kuranas.android.ui.theme.GlassLightBorder
import com.kuranas.android.ui.theme.GlassLightFill
import com.kuranas.android.ui.theme.GlassStrongBorder
import com.kuranas.android.ui.theme.GlassStrongFill

enum class GlassLevel { Light, Strong, Flat }

fun Modifier.glass(
    level: GlassLevel = GlassLevel.Light,
    radius: Dp = 20.dp,
): Modifier {
    val shape = RoundedCornerShape(radius)
    val fill = when (level) {
        GlassLevel.Light -> GlassLightFill
        GlassLevel.Strong -> GlassStrongFill
        GlassLevel.Flat -> GlassFlatFill
    }
    val border = when (level) {
        GlassLevel.Light -> GlassLightBorder
        GlassLevel.Strong -> GlassStrongBorder
        GlassLevel.Flat -> GlassFlatBorder
    }
    return this
        .clip(shape)
        .background(fill, shape)
        .border(BorderStroke(1.dp, border), shape)
}
