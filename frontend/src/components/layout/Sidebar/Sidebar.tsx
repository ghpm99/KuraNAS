"use client"

import { Clock } from "lucide-react"
import Link from "next/link"
import { useUI } from "../../../contexts/UIContext"
import styles from "./Sidebar.module.css"

export default function Sidebar() {
  const { activePage } = useUI()

  return (
    <div className={styles.sidebar}>
      <div className={styles.header}>
        <h1 className={styles.title}>Showpad</h1>
      </div>
      <nav className={styles.nav}>
        <Link href="/" className={`${styles.navItem} ${activePage === "files" ? styles.active : ""}`}>
          <svg className={styles.icon} viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h7" />
          </svg>
          <span>Arquivos</span>
        </Link>
        <Link href="/activity" className={`${styles.navItem} ${activePage === "activity" ? styles.active : ""}`}>
          <Clock className={styles.icon} />
          <span>Diário de Atividades</span>
        </Link>
        <div className={styles.navSection}>
          <div className={styles.sectionTitle}>Categorias</div>
          <div className={styles.folderList}>
            <Link href="#" className={styles.folderItem}>
              <svg className={styles.icon} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                />
              </svg>
              <span>Trabalho</span>
            </Link>
            <Link href="#" className={styles.folderItem}>
              <svg className={styles.icon} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                />
              </svg>
              <span>Estudos</span>
            </Link>
            <Link href="#" className={styles.folderItem}>
              <svg className={styles.icon} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                />
              </svg>
              <span>Pessoal</span>
            </Link>
          </div>
        </div>
      </nav>
    </div>
  )
}
