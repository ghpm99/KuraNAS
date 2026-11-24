import { Autocomplete, TextField } from '@mui/material';

export type PlaylistRuleOption = {
    value: string;
    label: string;
};

type PlaylistRuleOptionsFieldProps = {
    label: string;
    selectedValues: string[];
    options: PlaylistRuleOption[];
    isLoading: boolean;
    hasError: boolean;
    onChange: (selectedValues: string[]) => void;
};

export default function PlaylistRuleOptionsField({
    label,
    selectedValues,
    options,
    isLoading,
    hasError,
    onChange,
}: PlaylistRuleOptionsFieldProps) {
    const selectedOptions = selectedValues.map(
        (selectedValue) =>
            options.find((option) => option.value === selectedValue) ?? {
                value: selectedValue,
                label: selectedValue,
            }
    );

    return (
        <Autocomplete
            multiple
            size="small"
            loading={isLoading}
            options={options}
            value={selectedOptions}
            getOptionLabel={(option) => option.label}
            isOptionEqualToValue={(option, selected) => option.value === selected.value}
            onChange={(_event, nextOptions) => onChange(nextOptions.map((option) => option.value))}
            renderInput={(params) => <TextField {...params} label={label} error={hasError} />}
        />
    );
}
