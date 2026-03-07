import { NotFound as NotFoundScreen } from '@/components/blocks/application/error-pages'

export default function NotFound() {
  return (
    <NotFoundScreen
      homeHref="/"
      backHref="/marketplace"
      showSearch
    />
  )
}
