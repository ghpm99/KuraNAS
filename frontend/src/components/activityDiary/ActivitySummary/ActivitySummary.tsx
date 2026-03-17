import { useActivityDiary } from '@/components/providers/activityDiaryProvider/ActivityDiaryContext';
import { formatDuration } from '@/utils';
import Card from '@/components/ui/Card/Card';
import { Box, Grid, Typography } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ReactNode } from 'react';

function SummaryItem({
    title,
    value,
    subtitle,
}: {
    title: string;
    value: ReactNode;
    subtitle?: string;
}) {
    return (
        <Box>
            <Typography variant="body2" color="text.secondary" gutterBottom>
                {title}
            </Typography>
            <Typography variant="h5">{value}</Typography>
            {subtitle && (
                <Typography variant="caption" color="text.secondary">
                    {subtitle}
                </Typography>
            )}
        </Box>
    );
}

const ActivitySummary = () => {
    const { data } = useActivityDiary();
    const { t } = useI18n();

    return (
        <Card title={t('DAY_SUMMARY_TITLE')}>
            <Grid container spacing={3}>
                <Grid size={{ xs: 12, sm: 4 }}>
                    <SummaryItem
                        title={t('TOTAL_ACTIVITIES')}
                        value={data?.summary?.total_activities}
                    />
                </Grid>
                <Grid size={{ xs: 12, sm: 4 }}>
                    <SummaryItem
                        title={t('TOTAL_WORKED_TIME')}
                        value={formatDuration(data?.summary?.total_time_spent_seconds)}
                    />
                </Grid>
                <Grid size={{ xs: 12, sm: 4 }}>
                    <SummaryItem
                        title={t('LONGEST_ACTIVITY')}
                        value={data?.summary?.longest_activity?.name}
                        subtitle={formatDuration(data?.summary?.longest_activity?.duration_seconds)}
                    />
                </Grid>
            </Grid>
        </Card>
    );
};

export default ActivitySummary;
