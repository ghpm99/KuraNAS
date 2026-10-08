import { useActivityDiary } from '@/components/providers/activityDiaryProvider/ActivityDiaryContext';
import useI18n from '@/components/i18n/provider/i18nContext';
import { Alert, Box, Typography } from '@mui/material';

const ActivityDiaryActionBar = () => {
    const { message } = useActivityDiary();
    const { t } = useI18n();

    return (
        <Box sx={{ mb: 2 }}>
            <Typography variant="h4" gutterBottom>
                {t('ACTIVITY_DIARY_TITLE')}
            </Typography>
            {message && <Alert severity={message.type}>{message.text}</Alert>}
        </Box>
    );
};

export default ActivityDiaryActionBar;
