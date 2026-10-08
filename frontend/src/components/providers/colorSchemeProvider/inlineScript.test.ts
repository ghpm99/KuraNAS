import { readFileSync } from 'fs';
import { join } from 'path';
import { colorSchemeTokens } from '@/theme/visualTokens';
import { themeModeStorageKey } from './colorScheme';

describe('index.html pre-render theme script', () => {
    const indexHtml = readFileSync(join(__dirname, '../../../../index.html'), 'utf-8');

    it('reads the same storage key used by the provider', () => {
        expect(indexHtml).toContain(`'${themeModeStorageKey}'`);
    });

    it('paints the same root backgrounds as the token table', () => {
        expect(indexHtml).toContain(colorSchemeTokens.dark.colors.backgroundRoot);
        expect(indexHtml).toContain(colorSchemeTokens.light.colors.backgroundRoot);
    });
});
