import { Dialog } from '@mui/material';
import type { VacuumRoom } from '@/types/vacuum';
import VacuumRoomRenameForm from './VacuumRoomRenameForm';

type VacuumRoomRenameDialogProps = {
    room: VacuumRoom | null;
    isSaving: boolean;
    onCancel: () => void;
    onConfirm: (name: string) => void;
};

const VacuumRoomRenameDialog = ({
    room,
    isSaving,
    onCancel,
    onConfirm,
}: VacuumRoomRenameDialogProps) => (
    <Dialog open={room !== null} onClose={onCancel} maxWidth="xs" fullWidth>
        {room && (
            <VacuumRoomRenameForm
                key={room.id}
                room={room}
                isSaving={isSaving}
                onCancel={onCancel}
                onConfirm={onConfirm}
            />
        )}
    </Dialog>
);

export default VacuumRoomRenameDialog;
