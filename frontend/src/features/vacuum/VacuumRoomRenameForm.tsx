import { Button, DialogActions, DialogContent, DialogTitle, TextField } from '@mui/material';
import { useState } from 'react';
import useI18n from '@/components/i18n/provider/i18nContext';
import type { VacuumRoom } from '@/types/vacuum';

export const vacuumRoomNameMaxLength = 40;

type VacuumRoomRenameFormProps = {
    room: VacuumRoom;
    isSaving: boolean;
    onCancel: () => void;
    onConfirm: (name: string) => void;
};

const VacuumRoomRenameForm = ({ room, isSaving, onCancel, onConfirm }: VacuumRoomRenameFormProps) => {
    const { t } = useI18n();
    const [typedName, setTypedName] = useState(room.name ?? '');

    return (
        <form
            onSubmit={(event) => {
                event.preventDefault();
                onConfirm(typedName.trim());
            }}
        >
            <DialogTitle>{t('VACUUM_ROOM_RENAME_TITLE')}</DialogTitle>
            <DialogContent>
                <TextField
                    autoFocus
                    fullWidth
                    margin="dense"
                    label={t('VACUUM_ROOM_RENAME_FIELD')}
                    helperText={t('VACUUM_ROOM_RENAME_HELP', { number: String(room.id) })}
                    value={typedName}
                    onChange={(event) => setTypedName(event.target.value)}
                    slotProps={{ htmlInput: { maxLength: vacuumRoomNameMaxLength } }}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={onCancel} disabled={isSaving}>
                    {t('ACTION_CANCEL')}
                </Button>
                <Button type="submit" variant="contained" disabled={isSaving}>
                    {t('VACUUM_ROOM_RENAME_SAVE')}
                </Button>
            </DialogActions>
        </form>
    );
};

export default VacuumRoomRenameForm;
