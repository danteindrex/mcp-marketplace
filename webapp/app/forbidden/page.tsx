import { Forbidden } from '@/components/blocks/application/error-pages'

export default function ForbiddenPage() {
  return <Forbidden homeHref="/" loginHref="/login" />
}
