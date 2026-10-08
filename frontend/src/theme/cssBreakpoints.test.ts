import { readdirSync, readFileSync, statSync } from 'fs';
import { join } from 'path';

const allowedMaxWidthValues = ['599.95', '899.95', '1199.95', '1599.95'];
const allowedMinWidthValues = ['600', '900', '1200', '1600'];

const collectStylesheetPaths = (directoryPath: string): string[] =>
    readdirSync(directoryPath).flatMap((entryName) => {
        const entryPath = join(directoryPath, entryName);
        if (statSync(entryPath).isDirectory()) return collectStylesheetPaths(entryPath);
        return entryPath.endsWith('.css') ? [entryPath] : [];
    });

const findMediaWidthConditions = (stylesheetSource: string) =>
    Array.from(stylesheetSource.matchAll(/@media[^{]*/g)).flatMap((mediaPrelude) =>
        Array.from(mediaPrelude[0].matchAll(/\((max|min)-width:\s*([\d.]+)px\)/g)).map(
            (condition) => ({ bound: condition[1] ?? '', widthPx: condition[2] ?? '' })
        )
    );

describe('css media query breakpoints', () => {
    const stylesheetPaths = collectStylesheetPaths(join(__dirname, '..'));

    it('finds the stylesheets to scan', () => {
        expect(stylesheetPaths.length).toBeGreaterThan(10);
    });

    it('only uses the breakpoint token widths in media queries', () => {
        const violations = stylesheetPaths.flatMap((stylesheetPath) =>
            findMediaWidthConditions(readFileSync(stylesheetPath, 'utf8'))
                .filter(({ bound, widthPx }) =>
                    bound === 'max'
                        ? !allowedMaxWidthValues.includes(widthPx)
                        : !allowedMinWidthValues.includes(widthPx)
                )
                .map(({ bound, widthPx }) => `${stylesheetPath}: ${bound}-width ${widthPx}px`)
        );

        expect(violations).toEqual([]);
    });
});
