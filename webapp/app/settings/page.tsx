import { SettingsBlocks } from '@/components/blocks/application'
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

export default async function SettingsPage() {
  const cookieStore = await cookies()
  const token = cookieStore.get('mcp_access_token')?.value
  const role = cookieStore.get('mcp_active_role')?.value
  const roleAllowed = role === 'buyer' || role === 'merchant' || role === 'admin'
  if (!token || !roleAllowed) {
    redirect('/login?next=/settings')
  }

  return <SettingsBlocks.Page defaultTab="profile" />
}
