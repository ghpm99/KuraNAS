import { CalendarSearch } from 'lucide-react';
import { Drawer, useMediaQuery, useTheme } from '@mui/material';
import { useMemo, useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ImageTimelineBucket } from '@/types/imageLibrary';
import styles from './ImageDateScrubber.module.css';

type ImageDateScrubberProps = {
    buckets: ImageTimelineBucket[];
    onSelectMonth: (year: number, month: number) => void;
};

type YearEntry = { year: number; months: ImageTimelineBucket[] };

const groupBucketsByYear = (buckets: ImageTimelineBucket[]): YearEntry[] => {
    const monthsByYear = new Map<number, ImageTimelineBucket[]>();
    for (const bucket of buckets) {
        monthsByYear.set(bucket.year, [...(monthsByYear.get(bucket.year) ?? []), bucket]);
    }
    return [...monthsByYear.entries()]
        .sort(([leftYear], [rightYear]) => rightYear - leftYear)
        .map(([year, months]) => ({
            year,
            months: [...months].sort((left, right) => right.month - left.month),
        }));
};

type MonthButtonsProps = {
    entry: YearEntry;
    onSelectMonth: (year: number, month: number) => void;
};

function MonthButtons({ entry, onSelectMonth }: MonthButtonsProps) {
    const { t } = useI18n();
    const locale = t('LOCALE');
    const shortMonthFormatter = useMemo(
        () => new Intl.DateTimeFormat(locale, { month: 'short', timeZone: 'UTC' }),
        [locale]
    );
    const longMonthFormatter = useMemo(
        () => new Intl.DateTimeFormat(locale, { month: 'long', year: 'numeric', timeZone: 'UTC' }),
        [locale]
    );

    return (
        <ul className={styles.monthList}>
            {entry.months.map((bucket) => {
                const monthStart = new Date(Date.UTC(bucket.year, bucket.month - 1, 1));
                return (
                    <li key={bucket.month}>
                        <button
                            type="button"
                            className={styles.monthButton}
                            aria-label={t('IMAGES_SCRUBBER_MONTH_ARIA', {
                                month: longMonthFormatter.format(monthStart),
                                count: String(bucket.count),
                            })}
                            onClick={() => onSelectMonth(bucket.year, bucket.month)}
                        >
                            {shortMonthFormatter.format(monthStart)}
                        </button>
                    </li>
                );
            })}
        </ul>
    );
}

export default function ImageDateScrubber({ buckets = [], onSelectMonth }: ImageDateScrubberProps) {
    const { t } = useI18n();
    const theme = useTheme();
    const isDesktop = useMediaQuery(theme.breakpoints.up('md'), { noSsr: true });
    const [expandedYear, setExpandedYear] = useState<number | null>(null);
    const [isSheetOpen, setIsSheetOpen] = useState(false);
    const yearEntries = useMemo(() => groupBucketsByYear(buckets), [buckets]);

    if (yearEntries.length === 0) {
        return null;
    }

    const expandedEntry = yearEntries.find((entry) => entry.year === expandedYear);

    if (isDesktop) {
        return (
            <nav className={styles.rail} aria-label={t('IMAGES_SCRUBBER_ARIA')}>
                {expandedEntry && (
                    <div className={styles.monthPanel}>
                        <MonthButtons
                            entry={expandedEntry}
                            onSelectMonth={(year, month) => {
                                setExpandedYear(null);
                                onSelectMonth(year, month);
                            }}
                        />
                    </div>
                )}
                <ul className={styles.yearList}>
                    {yearEntries.map((entry) => (
                        <li key={entry.year}>
                            <button
                                type="button"
                                className={
                                    entry.year === expandedYear
                                        ? `${styles.yearButton} ${styles.yearButtonActive}`
                                        : styles.yearButton
                                }
                                aria-expanded={entry.year === expandedYear}
                                onClick={() =>
                                    setExpandedYear(entry.year === expandedYear ? null : entry.year)
                                }
                            >
                                {entry.year}
                            </button>
                        </li>
                    ))}
                </ul>
            </nav>
        );
    }

    return (
        <>
            <button
                type="button"
                className={styles.openSheetButton}
                onClick={() => setIsSheetOpen(true)}
            >
                <CalendarSearch size={16} />
                <span>{t('IMAGES_SCRUBBER_OPEN')}</span>
            </button>
            <Drawer
                anchor="bottom"
                open={isSheetOpen}
                onClose={() => setIsSheetOpen(false)}
                slotProps={{
                    paper: { className: styles.sheet, 'aria-label': t('IMAGES_SCRUBBER_ARIA') },
                }}
            >
                {yearEntries.map((entry) => (
                    <section key={entry.year} className={styles.sheetYear}>
                        <h3>{entry.year}</h3>
                        <MonthButtons
                            entry={entry}
                            onSelectMonth={(year, month) => {
                                setIsSheetOpen(false);
                                onSelectMonth(year, month);
                            }}
                        />
                    </section>
                ))}
            </Drawer>
        </>
    );
}
