import useI18n from '@/components/i18n/provider/i18nContext';
import { Button } from '@mui/material';

type TrashRestoredToastActionProps = {
    onViewInFolder: () => void;
};

const TrashRestoredToastAction = ({ onViewInFolder }: TrashRestoredToastActionProps) => {
    const { t } = useI18n();

    return (
        <Button size="small" color="inherit" onClick={onViewInFolder}>
            {t('TRASH_VIEW_IN_FOLDER')}
        </Button>
    );
};

export default TrashRestoredToastAction;
