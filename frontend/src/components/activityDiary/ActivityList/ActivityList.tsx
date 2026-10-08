import { useActivityDiary } from '@/components/providers/activityDiaryProvider/ActivityDiaryContext';
import { formatDate, formatDuration } from '@/utils';
import Card from '@/components/ui/Card/Card';
import {
    CircularProgress,
    IconButton,
    Table,
    TableContainer,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Typography,
} from '@mui/material';
import { Copy } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import ErrorState from '@/components/errorState/errorState';
import LoadMoreSentinel from '@/components/loadMoreSentinel/loadMoreSentinel';

const ActivityList = () => {
    const {
        data,
        getCurrentDuration,
        copyActivity,
        isEntriesLoading,
        hasEntriesError,
        entriesErrorMessage,
        hasMoreEntries,
        isFetchingMoreEntries,
        loadMoreEntries,
        reloadEntries,
    } = useActivityDiary();
    const { t } = useI18n();

    const renderBody = () => {
        if (isEntriesLoading) {
            return <CircularProgress size={24} role="progressbar" aria-label={t('LOADING')} />;
        }
        if (hasEntriesError) {
            return (
                <ErrorState
                    title={t('ACTIVITY_DIARY_LOAD_ERROR')}
                    backendMessage={entriesErrorMessage}
                    onRetry={reloadEntries}
                />
            );
        }
        if (!data?.entries?.items?.length) {
            return <Typography>{t('NO_ACTIVITIES')}</Typography>;
        }
        return (
            <>
                <TableContainer>
                    <Table>
                        <TableHead>
                            <TableRow>
                                <TableCell>{t('NAME')}</TableCell>
                                <TableCell>{t('DESCRIPTION')}</TableCell>
                                <TableCell>{t('START')}</TableCell>
                                <TableCell>{t('END')}</TableCell>
                                <TableCell>{t('DURATION')}</TableCell>
                                <TableCell>{t('ACTION')}</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {data?.entries?.items.map((activity) => (
                                <TableRow
                                    key={activity.id}
                                    sx={
                                        activity.end_time === null
                                            ? { bgcolor: 'action.hover' }
                                            : undefined
                                    }
                                >
                                    <TableCell>{activity.name}</TableCell>
                                    <TableCell>{activity.description || '-'}</TableCell>
                                    <TableCell>{formatDate(activity.start_time)}</TableCell>
                                    <TableCell>
                                        {activity.end_time.HasValue
                                            ? formatDate(activity.end_time.Value)
                                            : t('IN_PROGRESS')}
                                    </TableCell>
                                    <TableCell>
                                        {activity.end_time.HasValue
                                            ? formatDuration(activity.duration)
                                            : formatDuration(
                                                  getCurrentDuration(activity.start_time)
                                              )}
                                    </TableCell>
                                    <TableCell>
                                        <IconButton
                                            size="small"
                                            onClick={() => copyActivity(activity)}
                                        >
                                            <Copy size={16} />
                                        </IconButton>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </TableContainer>
                <LoadMoreSentinel
                    hasNextPage={hasMoreEntries}
                    isFetchingNextPage={isFetchingMoreEntries}
                    fetchNextPage={loadMoreEntries}
                />
            </>
        );
    };

    return <Card title={t('REGISTERED_ACTIVITIES_TITLE')}>{renderBody()}</Card>;
};

export default ActivityList;
