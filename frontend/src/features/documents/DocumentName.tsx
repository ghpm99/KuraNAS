import styles from './DocumentName.module.css';

type DocumentNameProps = {
    name: string;
    className?: string;
    onOpen?: () => void;
};

const DocumentName = ({ name, className, onOpen }: DocumentNameProps) => {
    if (!onOpen) {
        return (
            <span className={className} title={name}>
                {name}
            </span>
        );
    }

    return (
        <button
            type="button"
            className={`${className ?? ''} ${styles.openButton}`}
            title={name}
            onClick={onOpen}
        >
            {name}
        </button>
    );
};

export default DocumentName;
