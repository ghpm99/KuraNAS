export type ColorScheme = 'dark' | 'light';

type SchemeTokens = {
    colors: {
        backgroundRoot: string;
        backgroundElevated: string;
        surface1: string;
        surface2: string;
        surface3: string;
        borderSubtle: string;
        borderStrong: string;
        textPrimary: string;
        textSecondary: string;
        textMuted: string;
        textDisabled: string;
        primary: string;
        primaryHover: string;
        cyanAccent: string;
        pinkAccent: string;
        success: string;
        warning: string;
        danger: string;
        link: string;
        textOnMedia: string;
        overlaySoft: string;
        overlayStrong: string;
    };
    channels: {
        ink: string;
        primary: string;
        backgroundElevated: string;
        surface1: string;
        surface2: string;
        surface3: string;
    };
    backgrounds: {
        app: string;
        panel: string;
        panelElevated: string;
    };
    shadow: {
        card: string;
        floating: string;
        activeGlowPrimary: string;
        layered: string;
    };
    glass: {
        medium: { background: string; backdropFilter: string };
        strong: { background: string; backdropFilter: string };
    };
    scrollbar: {
        track: string;
        thumb: string;
    };
    metaThemeColor: string;
};

const darkSchemeTokens: SchemeTokens = {
    colors: {
        backgroundRoot: '#070A10',
        backgroundElevated: '#0A1018',
        surface1: '#0F1623',
        surface2: '#121A29',
        surface3: '#182234',
        borderSubtle: '#1B2637',
        borderStrong: '#263349',
        textPrimary: '#F3F7FF',
        textSecondary: '#B8C2D9',
        textMuted: '#7E8AA3',
        textDisabled: '#5B6579',
        primary: '#6D5DF6',
        primaryHover: '#7C70FF',
        cyanAccent: '#06B6D4',
        pinkAccent: '#E11D48',
        success: '#22C55E',
        warning: '#F59E0B',
        danger: '#EF4444',
        link: '#7DD3FC',
        textOnMedia: '#F8FAFC',
        overlaySoft: 'rgba(7, 10, 16, 0.62)',
        overlayStrong: 'rgba(7, 10, 16, 0.82)',
    },
    channels: {
        ink: '255, 255, 255',
        primary: '109, 93, 246',
        backgroundElevated: '10, 16, 24',
        surface1: '15, 22, 35',
        surface2: '18, 26, 41',
        surface3: '24, 34, 52',
    },
    backgrounds: {
        app: 'radial-gradient(1200px 400px at 20% -10%, rgba(225, 29, 72, 0.18), transparent 60%), radial-gradient(900px 320px at 95% 2%, rgba(14, 116, 144, 0.22), transparent 62%), linear-gradient(180deg, #070a10 0%, #0a1018 100%)',
        panel: 'linear-gradient(180deg, rgba(18, 26, 41, 0.94) 0%, rgba(15, 22, 35, 0.98) 100%)',
        panelElevated:
            'linear-gradient(180deg, rgba(24, 34, 52, 0.96) 0%, rgba(18, 26, 41, 0.94) 100%)',
    },
    shadow: {
        card: '0 8px 30px rgba(0, 0, 0, 0.24)',
        floating: '0 12px 40px rgba(0, 0, 0, 0.35)',
        activeGlowPrimary: '0 0 0 1px rgba(109, 93, 246, 0.6), 0 8px 30px rgba(109, 93, 246, 0.18)',
        layered:
            '0 1px 2px rgba(0,0,0,0.1), 0 4px 8px rgba(0,0,0,0.1), 0 16px 32px rgba(0,0,0,0.1)',
    },
    glass: {
        medium: {
            background: 'rgba(18, 26, 41, 0.7)',
            backdropFilter: 'blur(12px)',
        },
        strong: {
            background: 'rgba(15, 22, 35, 0.85)',
            backdropFilter: 'blur(16px)',
        },
    },
    scrollbar: {
        track: 'rgba(7, 10, 16, 0.26)',
        thumb: 'rgba(126, 138, 163, 0.38)',
    },
    metaThemeColor: '#070A10',
};

const lightSchemeTokens: SchemeTokens = {
    colors: {
        backgroundRoot: '#F3F5FA',
        backgroundElevated: '#E9EDF5',
        surface1: '#FFFFFF',
        surface2: '#FBFCFE',
        surface3: '#EFF2F9',
        borderSubtle: '#DCE2EE',
        borderStrong: '#C3CCDE',
        textPrimary: '#0F172A',
        textSecondary: '#475569',
        textMuted: '#64748B',
        textDisabled: '#94A3B8',
        primary: '#6D5DF6',
        primaryHover: '#5848DE',
        cyanAccent: '#0891B2',
        pinkAccent: '#E11D48',
        success: '#16A34A',
        warning: '#D97706',
        danger: '#DC2626',
        link: '#0369A1',
        textOnMedia: '#F8FAFC',
        overlaySoft: 'rgba(243, 245, 250, 0.7)',
        overlayStrong: 'rgba(243, 245, 250, 0.9)',
    },
    channels: {
        ink: '15, 23, 42',
        primary: '109, 93, 246',
        backgroundElevated: '233, 237, 245',
        surface1: '255, 255, 255',
        surface2: '251, 252, 254',
        surface3: '239, 242, 249',
    },
    backgrounds: {
        app: 'radial-gradient(1200px 400px at 20% -10%, rgba(225, 29, 72, 0.08), transparent 60%), radial-gradient(900px 320px at 95% 2%, rgba(14, 116, 144, 0.1), transparent 62%), linear-gradient(180deg, #f3f5fa 0%, #e9edf5 100%)',
        panel: 'linear-gradient(180deg, rgba(255, 255, 255, 0.96) 0%, rgba(251, 252, 254, 0.98) 100%)',
        panelElevated:
            'linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(239, 242, 249, 0.96) 100%)',
    },
    shadow: {
        card: '0 8px 30px rgba(15, 23, 42, 0.08)',
        floating: '0 12px 40px rgba(15, 23, 42, 0.16)',
        activeGlowPrimary: '0 0 0 1px rgba(109, 93, 246, 0.5), 0 8px 30px rgba(109, 93, 246, 0.12)',
        layered:
            '0 1px 2px rgba(15,23,42,0.06), 0 4px 8px rgba(15,23,42,0.06), 0 16px 32px rgba(15,23,42,0.06)',
    },
    glass: {
        medium: {
            background: 'rgba(255, 255, 255, 0.78)',
            backdropFilter: 'blur(12px)',
        },
        strong: {
            background: 'rgba(255, 255, 255, 0.9)',
            backdropFilter: 'blur(16px)',
        },
    },
    scrollbar: {
        track: 'rgba(15, 23, 42, 0.05)',
        thumb: 'rgba(71, 85, 105, 0.35)',
    },
    metaThemeColor: '#F3F5FA',
};

export const colorSchemeTokens: Record<ColorScheme, SchemeTokens> = {
    dark: darkSchemeTokens,
    light: lightSchemeTokens,
};

export const breakpointsPx = {
    phone: 600,
    tablet: 900,
    desktop: 1200,
    wide: 1600,
} as const;

const toBelowBreakpointPx = (breakpointPx: number) => (breakpointPx * 100 - 5) / 100;

export const belowBreakpointMediaQuery = (breakpointPx: number) =>
    `(max-width: ${toBelowBreakpointPx(breakpointPx)}px)`;

export const atLeastBreakpointMediaQuery = (breakpointPx: number) =>
    `(min-width: ${breakpointPx}px)`;

export const viewportMediaQueries = {
    belowPhone: belowBreakpointMediaQuery(breakpointsPx.phone),
    belowTablet: belowBreakpointMediaQuery(breakpointsPx.tablet),
    belowDesktop: belowBreakpointMediaQuery(breakpointsPx.desktop),
    compactDesktop: `${atLeastBreakpointMediaQuery(breakpointsPx.tablet)} and ${belowBreakpointMediaQuery(breakpointsPx.desktop)}`,
} as const;

export const visualTokens = {
    colors: darkSchemeTokens.colors,
    backgrounds: darkSchemeTokens.backgrounds,
    spacing: {
        1: '4px',
        2: '8px',
        3: '12px',
        4: '16px',
        5: '20px',
        6: '24px',
        8: '32px',
        10: '40px',
        12: '48px',
        16: '64px',
    },
    radius: {
        sm: '10px',
        md: '14px',
        lg: '18px',
        xl: '24px',
        pill: '999px',
    },
    typography: {
        pageTitle: '1.5rem',
        pageSubtitle: '0.875rem',
        sectionTitle: '1.125rem',
    },
    layout: {
        sidebarWidth: '240px',
        sidebarWidthCollapsed: '80px',
        headerHeight: '64px',
        headerHeightMobile: '56px',
        bottomNavHeight: '64px',
        playerHeight: '88px',
        playerHeightMobile: '64px',
        toastGap: '8px',
        contentMaxWidth: '1440px',
    },
    intrinsicSize: {
        fileCardHeight: '220px',
        fileRowHeight: '56px',
        imageTileHeight: '180px',
        trackRowHeight: '56px',
        videoCardHeight: '240px',
    },
    motion: {
        fast: '160ms',
        base: '220ms',
        slow: '300ms',
    },
} as const;

export const structuralCssVariables = {
    '--app-spacing-1': visualTokens.spacing[1],
    '--app-spacing-2': visualTokens.spacing[2],
    '--app-spacing-3': visualTokens.spacing[3],
    '--app-spacing-4': visualTokens.spacing[4],
    '--app-spacing-5': visualTokens.spacing[5],
    '--app-spacing-6': visualTokens.spacing[6],
    '--app-spacing-8': visualTokens.spacing[8],
    '--app-spacing-10': visualTokens.spacing[10],
    '--app-spacing-12': visualTokens.spacing[12],
    '--app-spacing-16': visualTokens.spacing[16],
    '--app-radius-sm': visualTokens.radius.sm,
    '--app-radius-md': visualTokens.radius.md,
    '--app-radius-lg': visualTokens.radius.lg,
    '--app-radius-xl': visualTokens.radius.xl,
    '--app-radius-pill': visualTokens.radius.pill,
    '--app-font-page-title': visualTokens.typography.pageTitle,
    '--app-font-page-subtitle': visualTokens.typography.pageSubtitle,
    '--app-font-section-title': visualTokens.typography.sectionTitle,
    '--app-shell-sidebar-width': visualTokens.layout.sidebarWidth,
    '--app-shell-sidebar-width-collapsed': visualTokens.layout.sidebarWidthCollapsed,
    '--app-shell-header-height': visualTokens.layout.headerHeight,
    '--app-shell-header-height-mobile': visualTokens.layout.headerHeightMobile,
    '--app-shell-bottom-nav-height': visualTokens.layout.bottomNavHeight,
    '--app-shell-player-height': visualTokens.layout.playerHeight,
    '--app-shell-player-height-mobile': visualTokens.layout.playerHeightMobile,
    '--app-toast-gap': visualTokens.layout.toastGap,
    '--app-safe-area-top': 'env(safe-area-inset-top, 0px)',
    '--app-safe-area-bottom': 'env(safe-area-inset-bottom, 0px)',
    '--app-content-max-width': visualTokens.layout.contentMaxWidth,
    '--app-intrinsic-file-card-height': visualTokens.intrinsicSize.fileCardHeight,
    '--app-intrinsic-file-row-height': visualTokens.intrinsicSize.fileRowHeight,
    '--app-intrinsic-image-tile-height': visualTokens.intrinsicSize.imageTileHeight,
    '--app-intrinsic-track-row-height': visualTokens.intrinsicSize.trackRowHeight,
    '--app-intrinsic-video-card-height': visualTokens.intrinsicSize.videoCardHeight,
    '--app-motion-fast': visualTokens.motion.fast,
    '--app-motion-base': visualTokens.motion.base,
    '--app-motion-slow': visualTokens.motion.slow,
} as const;

export const buildColorSchemeCssVariables = (scheme: ColorScheme) => {
    const tokens = colorSchemeTokens[scheme];
    return {
        '--app-color-background-root': tokens.colors.backgroundRoot,
        '--app-color-background-elevated': tokens.colors.backgroundElevated,
        '--app-color-surface-1': tokens.colors.surface1,
        '--app-color-surface-2': tokens.colors.surface2,
        '--app-color-surface-3': tokens.colors.surface3,
        '--app-color-border-subtle': tokens.colors.borderSubtle,
        '--app-color-border-strong': tokens.colors.borderStrong,
        '--app-color-text-primary': tokens.colors.textPrimary,
        '--app-color-text-secondary': tokens.colors.textSecondary,
        '--app-color-text-muted': tokens.colors.textMuted,
        '--app-color-text-disabled': tokens.colors.textDisabled,
        '--app-color-primary': tokens.colors.primary,
        '--app-color-primary-hover': tokens.colors.primaryHover,
        '--app-color-cyan-accent': tokens.colors.cyanAccent,
        '--app-color-pink-accent': tokens.colors.pinkAccent,
        '--app-color-success': tokens.colors.success,
        '--app-color-warning': tokens.colors.warning,
        '--app-color-danger': tokens.colors.danger,
        '--app-color-link': tokens.colors.link,
        '--app-color-text-on-media': tokens.colors.textOnMedia,
        '--app-color-overlay-soft': tokens.colors.overlaySoft,
        '--app-color-overlay-strong': tokens.colors.overlayStrong,
        '--app-color-ink-rgb': tokens.channels.ink,
        '--app-color-primary-rgb': tokens.channels.primary,
        '--app-color-background-elevated-rgb': tokens.channels.backgroundElevated,
        '--app-color-surface-1-rgb': tokens.channels.surface1,
        '--app-color-surface-2-rgb': tokens.channels.surface2,
        '--app-color-surface-3-rgb': tokens.channels.surface3,
        '--app-background-app': tokens.backgrounds.app,
        '--app-background-panel': tokens.backgrounds.panel,
        '--app-background-panel-elevated': tokens.backgrounds.panelElevated,
        '--app-shadow-card': tokens.shadow.card,
        '--app-shadow-floating': tokens.shadow.floating,
        '--app-shadow-active-primary': tokens.shadow.activeGlowPrimary,
        '--app-shadow-layered': tokens.shadow.layered,
        '--app-glass-medium-background': tokens.glass.medium.background,
        '--app-glass-medium-backdrop': tokens.glass.medium.backdropFilter,
        '--app-glass-strong-background': tokens.glass.strong.background,
        '--app-glass-strong-backdrop': tokens.glass.strong.backdropFilter,
    };
};
