import type { ReactNode } from "react"
import styles from "./Card.module.css"

interface CardProps {
  title: string
  children: ReactNode
  className?: string
}

export default function Card({ title, children, className = "" }: CardProps) {
  return (
    <div className={`${styles.card} ${className}`}>
      <h2 className={styles.title}>{title}</h2>
      <div className={styles.content}>{children}</div>
    </div>
  )
}
