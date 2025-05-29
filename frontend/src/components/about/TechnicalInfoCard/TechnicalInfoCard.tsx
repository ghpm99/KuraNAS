import Card from '@/components/ui/Card/Card';
import { useState } from 'react';
import styles from './TechicalInfoCard.module.css';
import Button from '@/components/ui/Button/Button';
import { Copy } from 'lucide-react';
import { useAbout } from '@/components/hooks/AboutProvider/AboutContext';

const TechnicalInfoCard = () => {
	const { commit_hash } = useAbout();
	const [copied, setCopied] = useState(false);

	const copyCommitHash = async () => {
		try {
			await navigator.clipboard.writeText(commit_hash);
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		} catch (err) {
			console.error('Falha ao copiar:', err);
		}
	};

	return (
		<Card title='Informações Técnicas'>
			<div className={styles.techInfo}>
				<div className={styles.commitSection}>
					<div className={styles.commitHeader}>
						<span className={styles.label}>Hash do Commit</span>
						<Button variant='secondary' onClick={copyCommitHash} className={styles.copyButton}>
							<Copy className={styles.copyIcon} />
							{copied ? 'Copiado!' : 'Copiar'}
						</Button>
					</div>
					<div className={styles.commitHash}>{commit_hash}</div>
					<div className={styles.commitDescription}>Identificador único da versão atual do código</div>
				</div>

				<div className={styles.buildInfo}>
					<h4 className={styles.sectionTitle}>Detalhes da Build</h4>
					<div className={styles.buildDetails}>
						<div className={styles.buildItem}>
							<span className={styles.buildLabel}>Ambiente:</span>
							<span className={styles.buildValue}>Production</span>
						</div>
						<div className={styles.buildItem}>
							<span className={styles.buildLabel}>Compilador:</span>
							<span className={styles.buildValue}>-</span>
						</div>
						<div className={styles.buildItem}>
							<span className={styles.buildLabel}>Framework:</span>
							<span className={styles.buildValue}>-</span>
						</div>
						<div className={styles.buildItem}>
							<span className={styles.buildLabel}>Node.js:</span>
							<span className={styles.buildValue}>-</span>
						</div>
					</div>
				</div>
			</div>
		</Card>
	);
};

export default TechnicalInfoCard;
