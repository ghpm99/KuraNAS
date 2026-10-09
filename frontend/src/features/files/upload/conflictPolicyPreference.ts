import type { UploadConflictPolicy } from '@/service/files';

export const defaultConflictPolicy: UploadConflictPolicy = 'rename';

export const conflictPolicyStorageKey = 'kuranas.files.upload.conflictPolicy';

const validPolicies: UploadConflictPolicy[] = ['rename', 'replace', 'skip'];

export const loadConflictPolicy = (): UploadConflictPolicy => {
    try {
        const storedPolicy = window.localStorage.getItem(conflictPolicyStorageKey);
        return validPolicies.includes(storedPolicy as UploadConflictPolicy)
            ? (storedPolicy as UploadConflictPolicy)
            : defaultConflictPolicy;
    } catch {
        return defaultConflictPolicy;
    }
};

export const saveConflictPolicy = (policy: UploadConflictPolicy): void => {
    try {
        window.localStorage.setItem(conflictPolicyStorageKey, policy);
    } catch {
        return;
    }
};
