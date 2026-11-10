export const formatDateTime = (isoTimestamp: string | undefined, languageTag?: string): string => {
    if (!isoTimestamp) return '';
    const moment = new Date(isoTimestamp);
    if (Number.isNaN(moment.getTime())) return '';
    return new Intl.DateTimeFormat(languageTag, { dateStyle: 'short', timeStyle: 'short' }).format(
        moment
    );
};
