import { getNextThemeMode } from './useThemeModeCycle';

describe('colorSchemeProvider/useThemeModeCycle', () => {
    it('cycles dark, light and system', () => {
        expect(getNextThemeMode('dark')).toBe('light');
        expect(getNextThemeMode('light')).toBe('system');
        expect(getNextThemeMode('system')).toBe('dark');
    });
});
