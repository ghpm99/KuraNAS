import { useState, type ReactNode } from 'react';

interface CoverArtProps {
    src: string;
    fallback: ReactNode;
}

const coverImageStyle = {
    width: '100%',
    height: '100%',
    objectFit: 'cover',
    display: 'block',
} as const;

const CoverImage = ({ src, fallback }: CoverArtProps) => {
    const [hasFailedToLoad, setHasFailedToLoad] = useState(false);

    if (hasFailedToLoad) {
        return <>{fallback}</>;
    }
    return (
        <img
            src={src}
            alt=""
            loading="lazy"
            style={coverImageStyle}
            onError={() => setHasFailedToLoad(true)}
        />
    );
};

const CoverArt = ({ src, fallback }: CoverArtProps) => (
    <CoverImage key={src} src={src} fallback={fallback} />
);

export default CoverArt;
