import { findBaseNameLength, findFileNameIssue } from './fileNameValidation';

describe('findFileNameIssue', () => {
    it.each([
        ['', undefined, 'empty'],
        ['   ', undefined, 'empty'],
        ['a/b', undefined, 'invalidCharacters'],
        ['a\\b', undefined, 'invalidCharacters'],
        ['same.txt', 'same.txt', 'unchanged'],
        ['  same.txt ', 'same.txt', 'unchanged'],
        ['other.txt', 'same.txt', null],
        ['new', undefined, null],
    ])('classifies %j against %j as %s', (typedName, currentName, expectedIssue) => {
        expect(findFileNameIssue(typedName, currentName)).toBe(expectedIssue);
    });
});

describe('findBaseNameLength', () => {
    it('excludes the extension for files', () => {
        expect(findBaseNameLength('report.final.pdf', false)).toBe('report.final'.length);
    });

    it('keeps the whole name for files without an extension or dotfiles', () => {
        expect(findBaseNameLength('Makefile', false)).toBe(8);
        expect(findBaseNameLength('.env', false)).toBe(4);
    });

    it('keeps the whole name for folders', () => {
        expect(findBaseNameLength('photos.2024', true)).toBe(11);
    });
});
