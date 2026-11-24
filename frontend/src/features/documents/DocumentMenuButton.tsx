import type { MouseEvent } from 'react';
import { IconButton } from '@mui/material';
import { EllipsisVertical } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';

type DocumentMenuButtonProps = {
    documentName: string;
    onOpen: (event: MouseEvent<HTMLElement>) => void;
};

const DocumentMenuButton = ({ documentName, onOpen }: DocumentMenuButtonProps) => {
    const { t } = useI18n();

    return (
        <IconButton
            size="small"
            aria-label={t('FILES_ITEM_MENU', { name: documentName })}
            onClick={onOpen}
        >
            <EllipsisVertical size={16} />
        </IconButton>
    );
};

export default DocumentMenuButton;
