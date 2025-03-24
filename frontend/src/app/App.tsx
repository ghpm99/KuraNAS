
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ChevronRight, Info, Play } from "lucide-react"
import MediaRow from "../components/mediaRow"
import "./App.css"


export type FileData = {
  id: number
  name: string
  format: string
  size: number
  updateAt: string
  createdAt: string
  lastInteraction: string
  lastBackup: string
}

export type Pagination = {
  hasNext: boolean
  hasPrevious: boolean
  page: number
  pageSize: number
}

export type PaginationResponse = {
  items: FileData[]
  pagination: Pagination
}

function usePosts() {
  return useQuery({
    queryKey: ['files'],
    queryFn: async (): Promise<PaginationResponse> => {
      const response = await fetch('http://localhost:8080/api/v1/files/')
      return await response.json()
    },
  })
}

export default function NetflixStyleGallery() {
  const queryClient = useQueryClient()
  const {status, data,error,isFetching} = usePosts()
  // Sample categories and media items
  console.log(data)
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

          <div className="category-container">
            <div className="category-header">
              <h2 className="category-title">Imagens</h2>
              <button className="category-see-all">
                <span>See all</span>
                <ChevronRight size={16} />
              </button>
            </div>
            {
              status === "success" && (

                <MediaRow items={data.items} />
              )
            }
          </div>

      </div>
    </div>
  )
}

