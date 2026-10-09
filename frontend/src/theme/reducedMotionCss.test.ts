import { readFileSync } from 'fs';
import { join } from 'path';

describe('global reduced motion stylesheet', () => {
    const globalStylesheet = readFileSync(join(__dirname, '..', 'index.css'), 'utf8');

    it('honors the operating system reduced motion preference', () => {
        expect(globalStylesheet).toContain('@media (prefers-reduced-motion: reduce)');
    });

    it('lets the user explicitly opt back into motion', () => {
        expect(globalStylesheet).toContain("html:not([data-app-motion='full'])");
    });

    it('still honors the reduce motion setting', () => {
        expect(globalStylesheet).toContain("html[data-app-motion='reduced']");
    });
});
