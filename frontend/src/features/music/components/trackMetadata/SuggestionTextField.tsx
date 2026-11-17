import { Autocomplete, TextField } from '@mui/material';

interface SuggestionTextFieldProps {
    label: string;
    value: string;
    suggestions: string[];
    errorMessage?: string;
    isDisabled: boolean;
    onChange: (nextValue: string) => void;
}

export default function SuggestionTextField({
    label,
    value,
    suggestions,
    errorMessage,
    isDisabled,
    onChange,
}: SuggestionTextFieldProps) {
    return (
        <Autocomplete
            freeSolo
            options={suggestions}
            value={value}
            inputValue={value}
            onInputChange={(_event, nextValue) => onChange(nextValue)}
            disabled={isDisabled}
            renderInput={(inputProps) => (
                <TextField
                    {...inputProps}
                    label={label}
                    error={errorMessage !== undefined}
                    helperText={errorMessage}
                />
            )}
        />
    );
}
