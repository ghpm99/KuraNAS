import Card from '@/components/ui/Card/Card';
import { useState } from 'react';
import styles from './TechicalInfoCard.module.css';
import Button from '@/components/ui/Button/Button';
import { Copy } from 'lucide-react';
import { useAbout } from '@/components/hooks/AboutProvider/AboutContext';
import useI18n from '@/components/i18n/provider/i18nContext';

const TechnicalInfoCard = () => {
	const { commit_hash, gin_mode, gin_version, go_version, node_version } = useAbout();
	const [copied, setCopied] = useState(false);
	const { t } = useI18n();

	const copyCommitHash = async () => {
		try {
			await navigator.clipboard.writeText(commit_hash);
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		} catch (err) {
			console.error(t('COPY_ERROR'), err);
		}
	};

	return (
		<Card title={t('TECHNICAL_INFO_TITLE')}>
			<div className={styles.techInfo}>
				<div className={styles.commitSection}>
					<div className={styles.commitHeader}>
						<span className={styles.label}>{t('COMMIT_HASH')}</span>
						<Button variant='secondary' onClick={copyCommitHash} className={styles.copyButton}>
							<Copy className={styles.copyIcon} />
							{copied ? t('COPIED') : t('COPY')}
						</Button>
					</div>
					<div className={styles.commitHash}>{commit_hash}</div>
					<div className={styles.commitDescription}>{t('COMMIT_DESCRIPTION')}</div>
				</div>

				<div className={styles.buildInfo}>
					<h4 className={styles.sectionTitle}>{t('BUILD_DETAILS_TITLE')}</h4>
					<div className={styles.buildDetails}>
						<div className={styles.buildItem}>
							<span className={styles.buildLabel}>{t('ENVIRONMENT')}</span>
							<span className={styles.buildValue}>{gin_mode}</span>
						</div>
						<div className={styles.buildItem}>
							<span className={styles.buildLabel}>{t('COMPILER')}</span>
							<span className={styles.buildValue}>{go_version}</span>
						</div>
						<div className={styles.buildItem}>
							<span className={styles.buildLabel}>{t('BACKEND')}</span>
							<span className={styles.buildValue}>{gin_version}</span>
						</div>
						<div className={styles.buildItem}>
							<span className={styles.buildLabel}>{t('NODEJS')}</span>
							<span className={styles.buildValue}>{node_version}</span>
						</div>
					</div>
				</div>
			</div>
		</Card>
	);
};

export default TechnicalInfoCard;
