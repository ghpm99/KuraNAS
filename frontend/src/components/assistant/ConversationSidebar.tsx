import { Box, Button, IconButton, List, ListItemButton, Typography } from '@mui/material';
import { Plus, Trash2 } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { Conversation } from '@/service/assistant';
import styles from './ConversationSidebar.module.css';

interface ConversationSidebarProps {
    conversations: Conversation[];
    activeId: number | null;
    onSelect: (id: number) => void;
    onNew: () => void;
    onDelete: (id: number) => void;
}

const ConversationSidebar = ({
    conversations,
    activeId,
    onSelect,
    onNew,
    onDelete,
}: ConversationSidebarProps) => {
    const { t } = useI18n();

    return (
        <Box className={styles.sidebar}>
            <Button
                fullWidth
                variant="outlined"
                startIcon={<Plus size={16} />}
                onClick={onNew}
                className={styles.newButton}
            >
                {t('ASSISTANT_NEW_CONVERSATION')}
            </Button>

            {conversations.length === 0 ? (
                <Typography className={styles.empty}>{t('ASSISTANT_NO_CONVERSATIONS')}</Typography>
            ) : (
                <List className={styles.list} disablePadding>
                    {conversations.map((conversation) => (
                        <ListItemButton
                            key={conversation.id}
                            selected={conversation.id === activeId}
                            onClick={() => onSelect(conversation.id)}
                            className={styles.item}
                        >
                            <span className={styles.title}>
                                {conversation.title || t('ASSISTANT_UNTITLED')}
                            </span>
                            <IconButton
                                size="small"
                                aria-label={t('ASSISTANT_DELETE_CONVERSATION')}
                                onClick={(event) => {
                                    event.stopPropagation();
                                    onDelete(conversation.id);
                                }}
                            >
                                <Trash2 size={15} />
                            </IconButton>
                        </ListItemButton>
                    ))}
                </List>
            )}
        </Box>
    );
};

export default ConversationSidebar;
