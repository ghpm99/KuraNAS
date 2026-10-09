import { useActivityDiary } from '@/components/providers/activityDiaryProvider/ActivityDiaryContext';
import { Alert } from '@mui/material';

const ActivityDiaryMessage = () => {
    const { message } = useActivityDiary();

    if (!message) {
        return null;
    }

    return <Alert severity={message.type}>{message.text}</Alert>;
};

export default ActivityDiaryMessage;
