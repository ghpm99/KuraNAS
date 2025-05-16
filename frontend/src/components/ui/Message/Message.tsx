import styles from "./Message.module.css"

interface MessageProps {
  text: string
  type: "success" | "error" | "info"
}

export default function Message({ text, type }: MessageProps) {
  return <div className={`${styles.message} ${styles[type]}`}>{text}</div>
}
