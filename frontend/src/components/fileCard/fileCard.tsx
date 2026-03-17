import {
    Box,
    Card,
    CardActionArea,
    CardContent,
    CardMedia,
    IconButton,
    Typography,
} from '@mui/material';
import { Star } from 'lucide-react';

const FileCard = ({
    title,
    metadata,
    thumbnail,
    onClick,
    starred,
    onClickStar,
}: {
    title: string;
    metadata: string;
    thumbnail: string;
    onClick: () => void;
    starred?: boolean;
    onClickStar?: () => void;
}) => {
    return (
        <Card sx={{ position: 'relative' }}>
            <CardActionArea onClick={onClick}>
                <CardMedia
                    component="img"
                    image={thumbnail || '/placeholder.svg'}
                    alt={title}
                    loading="lazy"
                    sx={{ height: 140, objectFit: 'cover' }}
                />
                <CardContent sx={{ py: 1 }}>
                    <Typography variant="body2" fontWeight={500} noWrap>
                        {title}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                        {metadata}
                    </Typography>
                </CardContent>
            </CardActionArea>
            <Box sx={{ position: 'absolute', top: 4, right: 4 }}>
                <IconButton size="small" onClick={onClickStar}>
                    <Star size={16} fill={starred ? 'currentColor' : 'none'} />
                </IconButton>
            </Box>
        </Card>
    );
};

export default FileCard;
