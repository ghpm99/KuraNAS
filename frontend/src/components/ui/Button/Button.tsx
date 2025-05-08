import type { LucideIcon } from "lucide-react"
import type { ButtonHTMLAttributes, ReactNode } from "react"
import styles from "./Button.module.css"

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode
  variant?: "primary" | "secondary"
  icon?: LucideIcon
}

export default function Button({ children, variant = "primary", icon: Icon, ...props }: ButtonProps) {
  return (
    <button className={`${styles.button} ${styles[variant]}`} {...props}>
      {Icon && <Icon className={styles.icon} />}
      {children}
    </button>
  )
}
