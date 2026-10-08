import { coarsePointerMediaQuery, createAppTheme, minimumTouchTargetPx } from './appTheme';

describe('createAppTheme touch targets', () => {
    it.each(['dark', 'light'] as const)(
        'enlarges small icon buttons on coarse pointers only (%s)',
        (colorScheme) => {
            const iconButtonOverrides = createAppTheme(colorScheme).components?.MuiIconButton
                ?.styleOverrides as Record<string, Record<string, unknown>>;

            expect(iconButtonOverrides.sizeSmall?.[coarsePointerMediaQuery]).toEqual({
                minWidth: minimumTouchTargetPx,
                minHeight: minimumTouchTargetPx,
            });
            expect(minimumTouchTargetPx).toBeGreaterThanOrEqual(44);
        }
    );
});
