
import { Play, Info, ChevronRight } from "lucide-react"
import MediaRow from "../components/mediaRow"
import "./App.css"

export default function NetflixStyleGallery() {
  // Sample categories and media items
  const categories = [
    {
      name: "Popular Now",
      items: [
        { id: 1, title: "Movie Title 1", type: "Movie", year: 2023 },
        { id: 2, title: "Series Title 1", type: "Series", year: 2022 },
        { id: 3, title: "Movie Title 2", type: "Movie", year: 2023 },
        { id: 4, title: "Series Title 2", type: "Series", year: 2021 },
        { id: 5, title: "Movie Title 3", type: "Movie", year: 2022 },
        { id: 6, title: "Series Title 3", type: "Series", year: 2023 },
      ],
    },
    {
      name: "Trending",
      items: [
        { id: 7, title: "Trending Movie 1", type: "Movie", year: 2023 },
        { id: 8, title: "Trending Series 1", type: "Series", year: 2022 },
        { id: 9, title: "Trending Movie 2", type: "Movie", year: 2023 },
        { id: 10, title: "Trending Series 2", type: "Series", year: 2021 },
        { id: 11, title: "Trending Movie 3", type: "Movie", year: 2022 },
        { id: 12, title: "Trending Series 3", type: "Series", year: 2023 },
      ],
    },
    {
      name: "New Releases",
      items: [
        { id: 13, title: "New Movie 1", type: "Movie", year: 2023 },
        { id: 14, title: "New Series 1", type: "Series", year: 2023 },
        { id: 15, title: "New Movie 2", type: "Movie", year: 2023 },
        { id: 16, title: "New Series 2", type: "Series", year: 2023 },
        { id: 17, title: "New Movie 3", type: "Movie", year: 2023 },
        { id: 18, title: "New Series 3", type: "Series", year: 2023 },
      ],
    },
  ]

  return (
    <div className="netflix-container">
      {/* Hero Section */}
      <div className="hero-section">
        <image
          href="/placeholder.svg?height=1080&width=1920"

          className="hero-image"
        />
        <div className="hero-gradient"></div>
        <div className="hero-content">
          <h1 className="hero-title">Featured Title</h1>
          <p className="hero-description">
            Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et
            dolore magna aliqua.
          </p>
          <div className="hero-buttons">
            <button className="btn btn-play">
              <Play size={20} />
              <span>Play</span>
            </button>
            <button className="btn btn-info">
              <Info size={20} />
              <span>More Info</span>
            </button>
          </div>
        </div>
      </div>

      {/* Content Rows */}
      <div className="content-rows">
        {categories.map((category) => (
          <div key={category.name} className="category-container">
            <div className="category-header">
              <h2 className="category-title">{category.name}</h2>
              <button className="category-see-all">
                <span>See all</span>
                <ChevronRight size={16} />
              </button>
            </div>
            <MediaRow items={category.items} />
          </div>
        ))}
      </div>
    </div>
  )
}

