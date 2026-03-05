'use client'

import { useState } from 'react'
import { ChevronDown, Plus } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'

export interface Team {
  id: string
  name: string
  avatar?: string
  role: 'owner' | 'member'
}

interface TeamSelectorProps {
  teams: Team[]
  selectedTeam: Team
  onTeamChange: (team: Team) => void
  onCreateTeam?: () => void
}

export function TeamSelector({
  teams,
  selectedTeam,
  onTeamChange,
  onCreateTeam,
}: TeamSelectorProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="w-full justify-between px-2"
        >
          <div className="flex items-center gap-2">
            <Avatar className="h-6 w-6">
              <AvatarImage src={selectedTeam.avatar} alt={selectedTeam.name} />
              <AvatarFallback>
                {selectedTeam.name
                  .split(' ')
                  .map((n) => n[0])
                  .join('')
                  .toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div className="text-left">
              <p className="text-sm font-medium leading-none">
                {selectedTeam.name}
              </p>
              <p className="text-xs text-muted-foreground capitalize">
                {selectedTeam.role}
              </p>
            </div>
          </div>
          <ChevronDown className="h-4 w-4 opacity-50" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuLabel>Teams</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {teams.map((team) => (
          <DropdownMenuItem
            key={team.id}
            onClick={() => onTeamChange(team)}
            className="flex items-center gap-2 cursor-pointer"
          >
            <Avatar className="h-6 w-6">
              <AvatarImage src={team.avatar} alt={team.name} />
              <AvatarFallback>
                {team.name
                  .split(' ')
                  .map((n) => n[0])
                  .join('')
                  .toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div className="flex-1">
              <p className="text-sm font-medium">{team.name}</p>
              <p className="text-xs text-muted-foreground capitalize">
                {team.role}
              </p>
            </div>
          </DropdownMenuItem>
        ))}
        {onCreateTeam && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={onCreateTeam}
              className="flex items-center gap-2 cursor-pointer"
            >
              <Plus className="h-4 w-4" />
              <span>Create team</span>
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
