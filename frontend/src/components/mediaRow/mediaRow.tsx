"use client"

import { useState, useRef } from "react"

import { ChevronLeft, ChevronRight, Play, Plus, ThumbsUp } from "lucide-react"
import "./mediaRow.css"

interface MediaItem {
  id: number
  title: string
  type: string
  year: number
}

interface MediaRowProps {
  items: MediaItem[]
}

export default function MediaRow({ items }: MediaRowProps) {
  const rowRef = useRef<HTMLDivElement>(null)
  const [showLeftArrow, setShowLeftArrow] = useState(false)
  const [showRightArrow, setShowRightArrow] = useState(true)
  const [hoveredItemId, setHoveredItemId] = useState<number | null>(null)

  const scroll = (direction: "left" | "right") => {
    if (rowRef.current) {
      const { scrollLeft, clientWidth } = rowRef.current
      const scrollTo = direction === "left" ? scrollLeft - clientWidth * 0.75 : scrollLeft + clientWidth * 0.75

      rowRef.current.scrollTo({
        left: scrollTo,
        behavior: "smooth",
      })
    }
  }

  const handleScroll = () => {
    if (rowRef.current) {
      const { scrollLeft, scrollWidth, clientWidth } = rowRef.current
      setShowLeftArrow(scrollLeft > 0)
      setShowRightArrow(scrollLeft < scrollWidth - clientWidth - 10)
    }
  }

  return (
    <div className="media-row-container">
      {showLeftArrow && (
        <button className="scroll-arrow scroll-arrow-left" onClick={() => scroll("left")}>
          <ChevronLeft size={30} />
        </button>
      )}

      <div ref={rowRef} className="media-items-container" onScroll={handleScroll}>
        {items.map((item) => (
          <div
            key={item.id}
            className="media-item-wrapper"
            onMouseEnter={() => setHoveredItemId(item.id)}
            onMouseLeave={() => setHoveredItemId(null)}
          >
            <div className={`media-item ${hoveredItemId === item.id ? "media-item-hovered" : ""}`}>
              <image
                href={`/placeholder.svg?height=160&width=280&text=${item.title}`}

                className="media-item-image"
              />

              {hoveredItemId === item.id && (
                <div className="media-item-overlay">
                  <div className="media-item-title">{item.title}</div>

                  <div className="media-item-details">
                    <div className="media-item-info">
                      <span className="media-item-match">98% Match</span>
                      <span className="media-item-type">{item.type}</span>
                      <span className="media-item-year">{item.year}</span>
                    </div>

                    <div className="media-item-actions">
                      <button className="media-action-button">
                        <Play size={16} />
                      </button>
                      <button className="media-action-button">
                        <Plus size={16} />
                      </button>
                      <button className="media-action-button">
                        <ThumbsUp size={16} />
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      {showRightArrow && (
        <button className="scroll-arrow scroll-arrow-right" onClick={() => scroll("right")}>
          <ChevronRight size={30} />
        </button>
      )}
    </div>
  )
}

