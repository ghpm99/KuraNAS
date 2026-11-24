import { parseCsv } from './parseCsv';

describe('parseCsv', () => {
    it('parses comma separated rows with quotes, escaped quotes and embedded newlines', () => {
        expect(parseCsv('a,b\n"x, y","he said ""hi"""\n"linha1\nlinha2",z')).toEqual([
            ['a', 'b'],
            ['x, y', 'he said "hi"'],
            ['linha1\nlinha2', 'z'],
        ]);
    });

    it('detects semicolon and tab delimiters and handles CRLF and BOM', () => {
        expect(parseCsv('﻿a;b\r\n1;2\r\n')).toEqual([
            ['a', 'b'],
            ['1', '2'],
        ]);
        expect(parseCsv('a\tb\n1\t2')).toEqual([
            ['a', 'b'],
            ['1', '2'],
        ]);
    });

    it('drops blank lines and returns nothing for an empty source', () => {
        expect(parseCsv('a,b\n\n1,2')).toEqual([
            ['a', 'b'],
            ['1', '2'],
        ]);
        expect(parseCsv('')).toEqual([]);
    });
});
