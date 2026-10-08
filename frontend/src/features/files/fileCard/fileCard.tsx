import type { MouseEvent } from 'react';
import {
    Box,
    Card,
    CardActionArea,
    CardContent,
    CardMedia,
    Checkbox,
    IconButton,
    Typography,
} from '@mui/material';
import { EllipsisVertical, Star } from 'lucide-react';
import useI18n from '@/components/i18n/provider/i18nContext';
import ColdTierIndicator from '@/features/files/coldTierIndicator/coldTierIndicator';

type FileCardProps = {
    title: string;
    metadata: string;
    secondaryText?: string;
    thumbnail: string;
    href?: string;
    fileId?: number;
    isTabStop?: boolean;
    onFocusItem?: () => void;
    onClick: (event: MouseEvent<HTMLElement>) => void;
    starred?: boolean;
    onClickStar?: () => void;
    isCold?: boolean;
    isSelected?: boolean;
    isSelectionActive?: boolean;
    onToggleSelection?: (event: MouseEvent<HTMLElement>) => void;
    onOpenMenu?: (event: MouseEvent<HTMLElement>) => void;
    onContextMenu?: (event: MouseEvent<HTMLElement>) => void;
};

const selectionControlClassName = 'fileCardSelectionControl';

const FileCard = ({
    title,
    metadata,
    secondaryText,
    thumbnail,
    href,
    fileId,
    isTabStop = true,
    onFocusItem,
    onClick,
    starred,
    onClickStar,
    isCold = false,
    isSelected = false,
    isSelectionActive = false,
    onToggleSelection,
    onOpenMenu,
    onContextMenu,
}: FileCardProps) => {
    const { t } = useI18n();
    const linkProps = href ? { component: 'a', href } : {};
    const isSelectionControlVisible = isSelected || isSelectionActive;

    return (
        <Card
            onContextMenu={onContextMenu}
            sx={{
                position: 'relative',
                outline: isSelected ? '2px solid' : 'none',
                outlineColor: 'primary.main',
                '&:has([data-file-id]:focus-visible)': {
                    boxShadow: (theme) => `0 0 0 3px ${theme.palette.primary.light}`,
                },
                [`& .${selectionControlClassName}`]: {
                    opacity: isSelectionControlVisible ? 1 : 0,
                },
                [`&:hover .${selectionControlClassName}, &:focus-within .${selectionControlClassName}`]:
                    { opacity: 1 },
                '@media (hover: none)': {
                    [`& .${selectionControlClassName}`]: { opacity: 1 },
                },
            }}
        >
            <CardActionArea
                {...linkProps}
                data-file-id={fileId}
                tabIndex={isTabStop ? 0 : -1}
                onFocus={onFocusItem}
                onClick={onClick}
            >
                <CardMedia
                    component="img"
                    image={thumbnail || '/placeholder.svg'}
                    alt=""
                    loading="lazy"
                    sx={{ height: 140, objectFit: 'cover' }}
                />
                <CardContent sx={{ py: 1 }}>
                    <Typography variant="body2" fontWeight={500} noWrap>
                        {title}
                    </Typography>
                    {secondaryText ? (
                        <Typography variant="caption" color="text.secondary" display="block" noWrap>
                            {secondaryText}
                        </Typography>
                    ) : null}
                    <Typography variant="caption" color="text.secondary">
                        {metadata}
                    </Typography>
                </CardContent>
            </CardActionArea>
            {onToggleSelection ? (
                <Box
                    className={selectionControlClassName}
                    sx={{ position: 'absolute', top: 0, left: 0 }}
                >
                    <Checkbox
                        size="small"
                        checked={isSelected}
                        onClick={(event) => {
                            event.stopPropagation();
                            onToggleSelection(event);
                        }}
                        slotProps={{
                            input: {
                                readOnly: true,
                                'aria-label': t('FILES_SELECT_ITEM', { name: title }),
                            },
                        }}
                        sx={{ bgcolor: 'background.paper', borderRadius: 1, m: 0.5, p: 0.5 }}
                    />
                </Box>
            ) : null}
            {isCold ? (
                <Box
                    sx={{
                        position: 'absolute',
                        top: 8,
                        left: onToggleSelection ? 44 : 8,
                        display: 'flex',
                        p: 0.5,
                        borderRadius: 1,
                        bgcolor: 'background.paper',
                    }}
                >
                    <ColdTierIndicator size={16} />
                </Box>
            ) : null}
            <Box sx={{ position: 'absolute', top: 4, right: 4, display: 'flex' }}>
                <IconButton
                    size="small"
                    aria-label={t(starred ? 'FILES_UNFAVORITE_ITEM' : 'FILES_FAVORITE_ITEM', {
                        name: title,
                    })}
                    aria-pressed={Boolean(starred)}
                    onClick={onClickStar}
                >
                    <Star size={16} fill={starred ? 'currentColor' : 'none'} />
                </IconButton>
                {onOpenMenu ? (
                    <IconButton
                        size="small"
                        aria-label={t('FILES_ITEM_MENU', { name: title })}
                        onClick={onOpenMenu}
                    >
                        <EllipsisVertical size={16} />
                    </IconButton>
                ) : null}
            </Box>
        </Card>
    );
};

export default FileCard;
