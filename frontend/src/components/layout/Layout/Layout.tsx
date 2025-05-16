"use client"

import type { ReactNode } from "react"
import { useActivity } from "../../../contexts/ActivityContext"
import { useUI } from "../../../contexts/UIContext"
import Header from "../Header/Header"
import Sidebar from "../Sidebar/Sidebar"
import styles from "./Layout.module.css"

interface LayoutProps {
  children: ReactNode
}

export default function Layout({ children }: LayoutProps) {
  const { activePage } = useUI()
  const { currentTime } = useActivity()

  const showClock = activePage === "activity"

  return (
    <div className={styles.layout}>
      <Sidebar />
      <div className={styles.mainContent}>
        <Header showClock={showClock} currentTime={currentTime} />
        <div className={styles.content}>{children}</div>
      </div>
    </div>
  )
}
