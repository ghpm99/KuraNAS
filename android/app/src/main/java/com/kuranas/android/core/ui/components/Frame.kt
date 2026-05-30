package com.kuranas.android.core.ui.components

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.DrawScope
import com.kuranas.android.ui.theme.BgCanvas

@Composable
fun KNFrame(
    modifier: Modifier = Modifier,
    content: @Composable BoxScope.() -> Unit,
) {
    Box(
        modifier = modifier
            .fillMaxSize()
            .drawBehind { drawBrandBackground() },
        content = content,
    )
}

private fun DrawScope.drawBrandBackground() {
    drawRect(color = BgCanvas)
    val w = size.width
    val h = size.height
    // navy/blue — topo-esquerda
    drawRect(
        brush = Brush.radialGradient(
            colors = listOf(Color(0x47313244), Color.Transparent),
            center = Offset(x = w * 0.15f, y = 0f),
            radius = maxOf(w, h) * 0.6f,
        ),
    )
    // teal — topo-direita
    drawRect(
        brush = Brush.radialGradient(
            colors = listOf(Color(0x2094E2D5), Color.Transparent),
            center = Offset(x = w, y = h * 0.18f),
            radius = maxOf(w, h) * 0.5f,
        ),
    )
    // navy escuro — base-centro
    drawRect(
        brush = Brush.radialGradient(
            colors = listOf(Color(0x3811111B), Color.Transparent),
            center = Offset(x = w * 0.5f, y = h * 1.1f),
            radius = maxOf(w, h) * 0.8f,
        ),
    )
}
