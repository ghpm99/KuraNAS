package com.kuranas.android.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.Color

data class SemanticColors(
    val positive: Color,
    val negative: Color,
    val alert: Color,
    val info: Color,
)

private val LocalSemanticColors = staticCompositionLocalOf {
    SemanticColors(
        positive = StatusPositive,
        negative = StatusNegative,
        alert = StatusAlert,
        info = StatusInfo,
    )
}

private val DarkColorScheme = darkColorScheme(
    primary = Blue500,
    onPrimary = BgCanvas,
    primaryContainer = Blue700,
    onPrimaryContainer = Blue400,
    secondary = Teal500,
    onSecondary = BgCanvas,
    secondaryContainer = Navy900,
    onSecondaryContainer = Teal400,
    background = BgCanvas,
    onBackground = TextPrimary,
    surface = SurfaceElevated,
    onSurface = TextPrimary,
    surfaceVariant = SurfaceRaised,
    onSurfaceVariant = TextSecondary,
    outline = GlassStrongBorder,
    outlineVariant = DividerColor,
    error = StatusNegative,
    onError = TextPrimary,
    scrim = ScrimDeep,
)

@Composable
fun KuranasTheme(content: @Composable () -> Unit) {
    CompositionLocalProvider(
        LocalSemanticColors provides SemanticColors(
            positive = StatusPositive,
            negative = StatusNegative,
            alert = StatusAlert,
            info = StatusInfo,
        )
    ) {
        MaterialTheme(
            colorScheme = DarkColorScheme,
            typography = Typography,
            content = content,
        )
    }
}

object KuranasTheme {
    val semantic: SemanticColors
        @Composable get() = LocalSemanticColors.current
}
