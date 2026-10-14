import useI18n from '@/components/i18n/provider/i18nContext';
import type { ShownShortcutFeedback } from './usePlayerKeyboardShortcuts';
import styles from './PlayerShortcutFeedback.module.css';

interface PlayerShortcutFeedbackProps {
    feedback?: ShownShortcutFeedback | null;
}

export default function PlayerShortcutFeedback({ feedback = null }: PlayerShortcutFeedbackProps) {
    const { t } = useI18n();

    return (
        <div className={styles.liveRegion} role="status" aria-live="polite">
            {feedback ? (
                <span key={feedback.id} className={styles.bubble}>
                    {t(feedback.messageKey, feedback.params)}
                </span>
            ) : null}
        </div>
    );
}
