import type { ReactNode } from 'react';
import type { SettingsSectionId } from './settingsSections';
import styles from './SettingsScreen.module.css';

type SettingsSectionAnchorProps = {
    sectionId: SettingsSectionId;
    children: ReactNode;
};

const SettingsSectionAnchor = ({ sectionId, children }: SettingsSectionAnchorProps) => (
    <div id={sectionId} className={styles.anchor}>
        {children}
    </div>
);

export default SettingsSectionAnchor;
