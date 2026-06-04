import type { KeyboardEvent } from 'react';
import { Box, CircularProgress, IconButton, TextField, Typography } from '@mui/material';
import { Send } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import useAssistantChat from './useAssistantChat';
import styles from './ChatScreen.module.css';

const ChatScreen = () => {
    const { t } = useI18n();
    const { messages, input, isLoading, hasError, setInput, send } = useAssistantChat();

    const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            void send();
        }
    };

    const isEmpty = messages.length === 0 && !isLoading;
    const canSend = !isLoading && input.trim() !== '';

    return (
        <Box className={styles.container}>
            <Box className={styles.header}>
                <Typography variant="h5">{t('ASSISTANT_TITLE')}</Typography>
                <Typography variant="body2" color="text.secondary">
                    {t('ASSISTANT_SUBTITLE')}
                </Typography>
            </Box>

            <Box className={styles.messages} data-testid="assistant-messages">
                {isEmpty ? (
                    <Typography className={styles.empty}>{t('ASSISTANT_EMPTY')}</Typography>
                ) : (
                    messages.map((message, index) => (
                        <Box
                            key={`${index}-${message.role}`}
                            className={message.role === 'user' ? styles.userRow : styles.assistantRow}
                        >
                            <Box className={styles.bubble}>{message.content}</Box>
                        </Box>
                    ))
                )}

                {isLoading && (
                    <Box className={styles.assistantRow}>
                        <Box className={`${styles.bubble} ${styles.thinking}`}>
                            <CircularProgress size={14} />
                            <span>{t('ASSISTANT_THINKING')}</span>
                        </Box>
                    </Box>
                )}
            </Box>

            {hasError && (
                <Typography role="alert" className={styles.error}>
                    {t('ASSISTANT_ERROR')}
                </Typography>
            )}

            <Box className={styles.inputRow}>
                <TextField
                    fullWidth
                    multiline
                    maxRows={4}
                    value={input}
                    onChange={(event) => setInput(event.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder={t('ASSISTANT_PLACEHOLDER')}
                    inputProps={{ 'aria-label': t('ASSISTANT_PLACEHOLDER') }}
                />
                <IconButton
                    aria-label={t('ASSISTANT_SEND')}
                    color="primary"
                    onClick={() => void send()}
                    disabled={!canSend}
                >
                    <Send size={20} />
                </IconButton>
            </Box>
        </Box>
    );
};

export default ChatScreen;
