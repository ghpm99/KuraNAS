import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Typography,
} from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { ShortcutDefinition } from './shortcutDefinition';

type GlobalShortcutsDialogProps = {
    isOpen: boolean;
    onClose: () => void;
    globalShortcuts: ShortcutDefinition[];
    pageShortcuts: ShortcutDefinition[];
};

type ShortcutsSectionProps = {
    titleKey: string;
    shortcuts: ShortcutDefinition[];
};

const ShortcutsSection = ({ titleKey, shortcuts }: ShortcutsSectionProps) => {
    const { t } = useI18n();

    return (
        <section>
            <Typography variant="h6" component="h2" gutterBottom>
                {t(titleKey)}
            </Typography>
            <Table size="small" aria-label={t(titleKey)}>
                <TableHead>
                    <TableRow>
                        <TableCell>{t('SHORTCUTS_COLUMN_KEYS')}</TableCell>
                        <TableCell>{t('SHORTCUTS_COLUMN_ACTION')}</TableCell>
                    </TableRow>
                </TableHead>
                <TableBody>
                    {shortcuts.map((shortcut) => (
                        <TableRow key={shortcut.descriptionKey}>
                            <TableCell sx={{ whiteSpace: 'nowrap' }}>
                                {shortcut.keyLabels.map((keyLabel) => (
                                    <kbd key={keyLabel} style={{ marginRight: 6 }}>
                                        {keyLabel}
                                    </kbd>
                                ))}
                            </TableCell>
                            <TableCell>{t(shortcut.descriptionKey)}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </section>
    );
};

const groupShortcutsBySection = (
    shortcuts: ShortcutDefinition[]
): Map<string, ShortcutDefinition[]> => {
    const shortcutsBySectionTitleKey = new Map<string, ShortcutDefinition[]>();
    shortcuts.forEach((shortcut) => {
        const sectionTitleKey = shortcut.sectionTitleKey ?? 'SHORTCUTS_SECTION_CURRENT_PAGE';
        shortcutsBySectionTitleKey.set(sectionTitleKey, [
            ...(shortcutsBySectionTitleKey.get(sectionTitleKey) ?? []),
            shortcut,
        ]);
    });
    return shortcutsBySectionTitleKey;
};

const GlobalShortcutsDialog = ({
    isOpen,
    onClose,
    globalShortcuts,
    pageShortcuts,
}: GlobalShortcutsDialogProps) => {
    const { t } = useI18n();

    return (
        <Dialog open={isOpen} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>{t('SHORTCUTS_DIALOG_TITLE')}</DialogTitle>
            <DialogContent>
                <ShortcutsSection titleKey="SHORTCUTS_SECTION_GLOBAL" shortcuts={globalShortcuts} />
                {Array.from(groupShortcutsBySection(pageShortcuts)).map(
                    ([sectionTitleKey, sectionShortcuts]) => (
                        <ShortcutsSection
                            key={sectionTitleKey}
                            titleKey={sectionTitleKey}
                            shortcuts={sectionShortcuts}
                        />
                    )
                )}
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>{t('CLOSE')}</Button>
            </DialogActions>
        </Dialog>
    );
};

export default GlobalShortcutsDialog;
