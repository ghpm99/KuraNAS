import { Alert } from '@mui/material';
import useI18n from '@/components/i18n/provider/i18nContext';

type VacuumFaultCodesProps = {
    faultCodes: number[] | undefined;
};

const VacuumFaultCodes = ({ faultCodes }: VacuumFaultCodesProps) => {
    const { t } = useI18n();

    if (!Array.isArray(faultCodes) || faultCodes.length === 0) return null;

    return (
        <>
            {faultCodes.map((faultCode) => (
                <Alert key={faultCode} severity="error" role="alert">
                    {t('VACUUM_FAULT_CODE', { code: String(faultCode) })}
                </Alert>
            ))}
        </>
    );
};

export default VacuumFaultCodes;
