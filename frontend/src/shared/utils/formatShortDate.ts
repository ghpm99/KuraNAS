export const formatShortDate = (isoTimestamp: string, locale?: string): string => {
    const date = new Date(isoTimestamp);
    if (Number.isNaN(date.getTime())) return '';
    return new Intl.DateTimeFormat(locale, { dateStyle: 'medium' }).format(date);
};
