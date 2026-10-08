import { FileType } from '@/utils';
import type { FileData } from '@/features/files/providers/fileProvider/fileContext';
import { formatModifiedDate, formatSizeColumn, formatTypeColumn } from './fileListFormatting';

const translate = (key: string) => key;

const buildFile = (overrides: Partial<FileData>): FileData =>
    ({ type: FileType.File, format: '.mp3', size: 2048, directory_content_count: 0, ...overrides }) as FileData;

describe('fileListFormatting', () => {
    it('formats the modified date with the requested locale', () => {
        const formattedDate = formatModifiedDate('2026-03-10T10:00:00Z', 'en-US');

        expect(formattedDate).toContain('2026');
        expect(formattedDate).toContain('Mar');
    });

    it('returns an empty string for missing or invalid dates', () => {
        expect(formatModifiedDate('')).toBe('');
        expect(formatModifiedDate('not-a-date')).toBe('');
    });

    it('shows byte size for files and item count for directories', () => {
        expect(formatSizeColumn(buildFile({}), translate)).toBe('2.00 KB');
        expect(
            formatSizeColumn(buildFile({ type: FileType.Directory, directory_content_count: 1 }), translate)
        ).toBe('1 ITEM');
        expect(
            formatSizeColumn(buildFile({ type: FileType.Directory, directory_content_count: 3 }), translate)
        ).toBe('3 ITENS');
        expect(
            formatSizeColumn(
                buildFile({ type: FileType.Directory, directory_content_count: undefined as never }),
                translate
            )
        ).toBe('0 ITENS');
    });

    it('shows the extension for files and the folder label for directories', () => {
        expect(formatTypeColumn(buildFile({}), translate)).toBe('MP3');
        expect(formatTypeColumn(buildFile({ format: undefined as never }), translate)).toBe('');
        expect(formatTypeColumn(buildFile({ type: FileType.Directory }), translate)).toBe('FOLDER');
    });
});
