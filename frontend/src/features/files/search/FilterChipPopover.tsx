import { useId, useState, type ReactNode } from 'react';
import { Chip, Popover } from '@mui/material';
import { ChevronDown } from 'lucide-react';
import styles from './fileSearch.module.css';

interface FilterChipPopoverProps {
    label: string;
    isActive?: boolean;
    children?: ReactNode;
}

const FilterChipPopover = ({ label, isActive = false, children }: FilterChipPopoverProps) => {
    const popoverId = useId();
    const [anchorElement, setAnchorElement] = useState<HTMLElement | null>(null);
    const isOpen = anchorElement !== null;

    return (
        <>
            <Chip
                label={label}
                size="small"
                color={isActive ? 'primary' : 'default'}
                variant={isActive ? 'filled' : 'outlined'}
                icon={<ChevronDown size={14} />}
                onClick={(event) => setAnchorElement(event.currentTarget)}
                aria-haspopup="true"
                aria-expanded={isOpen}
                aria-controls={isOpen ? popoverId : undefined}
            />
            <Popover
                id={popoverId}
                open={isOpen}
                anchorEl={anchorElement}
                onClose={() => setAnchorElement(null)}
                anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
            >
                <div className={styles.filterPopover}>{children}</div>
            </Popover>
        </>
    );
};

export default FilterChipPopover;
