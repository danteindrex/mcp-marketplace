'use client'

import { LogOut, Settings, User } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'

export interface Profile {
  id: string
  name: string
  email: string
  avatar?: string
}

interface ProfileSelectorProps {
  profile: Profile
  onSettings?: () => void
  onLogout?: () => void
}

export function ProfileSelector({
  profile,
  onSettings,
  onLogout,
}: ProfileSelectorProps) {
  const initials = profile.name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="relative h-8 w-8 rounded-full"
        >
          <Avatar className="h-8 w-8">
            <AvatarImage src={profile.avatar} alt={profile.name} />
            <AvatarFallback>{initials}</AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel className="flex flex-col space-y-1">
          <p className="text-sm font-medium leading-none">{profile.name}</p>
          <p className="text-xs leading-none text-muted-foreground">
            {profile.email}
          </p>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem className="flex items-center gap-2 cursor-pointer">
          <User className="h-4 w-4" />
          <span>Profile</span>
        </DropdownMenuItem>
        {onSettings && (
          <DropdownMenuItem
            onClick={onSettings}
            className="flex items-center gap-2 cursor-pointer"
          >
            <Settings className="h-4 w-4" />
            <span>Settings</span>
          </DropdownMenuItem>
        )}
        <DropdownMenuSeparator />
        {onLogout && (
          <DropdownMenuItem
            onClick={onLogout}
            className="flex items-center gap-2 cursor-pointer text-red-600 dark:text-red-400"
          >
            <LogOut className="h-4 w-4" />
            <span>Log out</span>
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
