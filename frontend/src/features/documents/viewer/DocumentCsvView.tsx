import { Alert, Box, Table, TableBody, TableCell, TableHead, TableRow } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import { parseCsv } from './parseCsv';

export const maxCsvRowsShown = 500;

const DocumentCsvView = ({ source }: { source: string }) => {
    const { t } = useI18n();
    const rows = parseCsv(source ?? '');
    const [headerRow = [], ...bodyRows] = rows;
    const shownRows = bodyRows.slice(0, maxCsvRowsShown);
    const isTruncated = bodyRows.length > maxCsvRowsShown;

    if (rows.length === 0) return <Alert severity="info">{t('TEXT_VIEWER_EMPTY')}</Alert>;

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1, minWidth: 0 }}>
            {isTruncated ? (
                <Alert severity="info">
                    {t('DOCUMENTS_VIEWER_CSV_TRUNCATED', { count: String(maxCsvRowsShown) })}
                </Alert>
            ) : null}
            <Box sx={{ overflow: 'auto', maxHeight: '70vh' }}>
                <Table size="small" stickyHeader>
                    <TableHead>
                        <TableRow>
                            {headerRow.map((headerCell, columnPosition) => (
                                <TableCell key={columnPosition}>{headerCell}</TableCell>
                            ))}
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {shownRows.map((row, rowPosition) => (
                            <TableRow key={rowPosition}>
                                {row.map((cell, columnPosition) => (
                                    <TableCell key={columnPosition}>{cell}</TableCell>
                                ))}
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </Box>
        </Box>
    );
};

export default DocumentCsvView;
