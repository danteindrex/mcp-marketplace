'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTheme } from 'next-themes'
import { Bell, CreditCard, Lock, Settings2, UserCircle2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  changeUserPassword,
  fetchBilling,
  fetchUserNotificationSettings,
  fetchUserPreferencesSettings,
  fetchUserProfileSettings,
  updateUserNotificationSettings,
  updateUserPreferencesSettings,
  updateUserProfileSettings,
  type Billing,
  type UserNotificationSettings,
  type UserPreferencesSettings,
  type UserProfileSettings,
} from '@/lib/api-client'
import { getActiveRole } from '@/lib/api'

type SettingsTab = 'profile' | 'security' | 'preferences' | 'notifications' | 'billing'

function SettingsPage({ defaultTab = 'profile' }: { defaultTab?: SettingsTab }) {
  const router = useRouter()
  const { setTheme } = useTheme()
  const [activeTab, setActiveTab] = useState<SettingsTab>(defaultTab)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const [profile, setProfile] = useState<UserProfileSettings | null>(null)
  const [preferences, setPreferences] = useState<UserPreferencesSettings | null>(null)
  const [notifications, setNotifications] = useState<UserNotificationSettings | null>(null)
  const [billing, setBilling] = useState<Billing | null>(null)

  const [passwordCurrent, setPasswordCurrent] = useState('')
  const [passwordNew, setPasswordNew] = useState('')
  const [passwordConfirm, setPasswordConfirm] = useState('')

  const role = useMemo(() => getActiveRole(), [])

  useEffect(() => {
    if (!role) {
      router.push('/login?next=/settings')
      return
    }
    let mounted = true
    setLoading(true)
    Promise.all([
      fetchUserProfileSettings(),
      fetchUserPreferencesSettings(),
      fetchUserNotificationSettings(),
      role === 'buyer' ? fetchBilling().catch(() => null) : Promise.resolve(null),
    ])
      .then(([profileData, preferenceData, notificationData, billingData]) => {
        if (!mounted) return
        setProfile(profileData)
        setPreferences(preferenceData)
        setNotifications(notificationData)
        setBilling(billingData)
        setTheme(preferenceData.theme)
      })
      .catch((e) => {
        if (!mounted) return
        setError(e instanceof Error ? e.message : 'Failed to load settings')
      })
      .finally(() => {
        if (mounted) setLoading(false)
      })

    return () => {
      mounted = false
    }
  }, [role, router, setTheme])

  async function saveProfile() {
    if (!profile) return
    setSaving(true)
    setError('')
    setSuccess('')
    try {
      const updated = await updateUserProfileSettings({
        name: profile.name,
        email: profile.email,
        phone: profile.phone || '',
        avatarUrl: profile.avatarUrl || '',
        locale: profile.locale || 'en-US',
        timezone: profile.timezone || 'America/Los_Angeles',
      })
      setProfile(updated)
      setSuccess('Profile saved')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to save profile')
    } finally {
      setSaving(false)
    }
  }

  async function savePreferences() {
    if (!preferences) return
    setSaving(true)
    setError('')
    setSuccess('')
    try {
      const updated = await updateUserPreferencesSettings(preferences)
      setPreferences(updated)
      setTheme(updated.theme)
      setSuccess('Preferences saved')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to save preferences')
    } finally {
      setSaving(false)
    }
  }

  async function saveNotifications() {
    if (!notifications) return
    setSaving(true)
    setError('')
    setSuccess('')
    try {
      const updated = await updateUserNotificationSettings(notifications)
      setNotifications(updated)
      setSuccess('Notification settings saved')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to save notifications')
    } finally {
      setSaving(false)
    }
  }

  async function savePassword() {
    if (!passwordCurrent || !passwordNew) {
      setError('Please enter current and new password')
      return
    }
    setSaving(true)
    setError('')
    setSuccess('')
    try {
      await changeUserPassword({
        currentPassword: passwordCurrent,
        newPassword: passwordNew,
        confirmPassword: passwordConfirm,
      })
      setPasswordCurrent('')
      setPasswordNew('')
      setPasswordConfirm('')
      setSuccess('Password changed')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to change password')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <main className="min-h-screen bg-background px-4 py-10">
        <div className="mx-auto max-w-5xl">
          <Card>
            <CardHeader>
              <CardTitle>Settings</CardTitle>
              <CardDescription>Loading your settings...</CardDescription>
            </CardHeader>
          </Card>
        </div>
      </main>
    )
  }

  return (
    <main className="min-h-screen bg-background px-4 py-10">
      <div className="mx-auto max-w-5xl space-y-6">
        <Card className="border-[3px] border-foreground shadow-[4px_4px_0px_hsl(var(--shadow-color))]">
          <CardHeader>
            <CardTitle className="text-3xl font-black uppercase tracking-wide">User Settings</CardTitle>
            <CardDescription>Manage your account profile, access, preferences, and notifications.</CardDescription>
          </CardHeader>
          <CardContent>
            {error ? <p className="mb-3 text-sm text-destructive">{error}</p> : null}
            {success ? <p className="mb-3 text-sm text-success">{success}</p> : null}

            <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as SettingsTab)}>
              <TabsList className="grid w-full grid-cols-5">
                <TabsTrigger value="profile" className="gap-1"><UserCircle2 className="h-4 w-4" />Profile</TabsTrigger>
                <TabsTrigger value="security" className="gap-1"><Lock className="h-4 w-4" />Security</TabsTrigger>
                <TabsTrigger value="preferences" className="gap-1"><Settings2 className="h-4 w-4" />Preferences</TabsTrigger>
                <TabsTrigger value="notifications" className="gap-1"><Bell className="h-4 w-4" />Notifications</TabsTrigger>
                <TabsTrigger value="billing" className="gap-1"><CreditCard className="h-4 w-4" />Billing</TabsTrigger>
              </TabsList>

              <TabsContent value="profile" className="space-y-4 pt-4">
                <div className="grid gap-4 md:grid-cols-2">
                  <div className="space-y-2">
                    <Label>Name</Label>
                    <Input value={profile?.name || ''} onChange={(e) => setProfile((p) => (p ? { ...p, name: e.target.value } : p))} />
                  </div>
                  <div className="space-y-2">
                    <Label>Email</Label>
                    <Input type="email" value={profile?.email || ''} onChange={(e) => setProfile((p) => (p ? { ...p, email: e.target.value } : p))} />
                  </div>
                  <div className="space-y-2">
                    <Label>Phone</Label>
                    <Input value={profile?.phone || ''} onChange={(e) => setProfile((p) => (p ? { ...p, phone: e.target.value } : p))} />
                  </div>
                  <div className="space-y-2">
                    <Label>Avatar URL</Label>
                    <Input value={profile?.avatarUrl || ''} onChange={(e) => setProfile((p) => (p ? { ...p, avatarUrl: e.target.value } : p))} />
                  </div>
                  <div className="space-y-2">
                    <Label>Locale</Label>
                    <Input value={profile?.locale || ''} onChange={(e) => setProfile((p) => (p ? { ...p, locale: e.target.value } : p))} />
                  </div>
                  <div className="space-y-2">
                    <Label>Timezone</Label>
                    <Input value={profile?.timezone || ''} onChange={(e) => setProfile((p) => (p ? { ...p, timezone: e.target.value } : p))} />
                  </div>
                </div>
                <Button onClick={saveProfile} disabled={saving}>Save Profile</Button>
              </TabsContent>

              <TabsContent value="security" className="space-y-4 pt-4">
                <div className="space-y-2">
                  <Label>Current Password</Label>
                  <Input type="password" value={passwordCurrent} onChange={(e) => setPasswordCurrent(e.target.value)} />
                </div>
                <div className="space-y-2">
                  <Label>New Password</Label>
                  <Input type="password" value={passwordNew} onChange={(e) => setPasswordNew(e.target.value)} />
                </div>
                <div className="space-y-2">
                  <Label>Confirm Password</Label>
                  <Input type="password" value={passwordConfirm} onChange={(e) => setPasswordConfirm(e.target.value)} />
                </div>
                <Button onClick={savePassword} disabled={saving}>Change Password</Button>
              </TabsContent>

              <TabsContent value="preferences" className="space-y-4 pt-4">
                <div className="grid gap-4 md:grid-cols-2">
                  <div className="space-y-2">
                    <Label>Theme (`light`, `dark`, or `system`)</Label>
                    <Input
                      value={preferences?.theme || 'system'}
                      onChange={(e) =>
                        setPreferences((p) =>
                          p ? { ...p, theme: (e.target.value || 'system') as UserPreferencesSettings['theme'] } : p,
                        )
                      }
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Language</Label>
                    <Input value={preferences?.language || 'en'} onChange={(e) => setPreferences((p) => (p ? { ...p, language: e.target.value } : p))} />
                  </div>
                  <div className="space-y-2">
                    <Label>Timezone</Label>
                    <Input value={preferences?.timezone || ''} onChange={(e) => setPreferences((p) => (p ? { ...p, timezone: e.target.value } : p))} />
                  </div>
                  <div className="space-y-2">
                    <Label>Default Landing Path</Label>
                    <Input value={preferences?.defaultLanding || '/'} onChange={(e) => setPreferences((p) => (p ? { ...p, defaultLanding: e.target.value } : p))} />
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Switch
                    checked={preferences?.compactMode || false}
                    onCheckedChange={(checked) => setPreferences((p) => (p ? { ...p, compactMode: checked } : p))}
                  />
                  <Label>Compact mode</Label>
                </div>
                <Button onClick={savePreferences} disabled={saving}>Save Preferences</Button>
              </TabsContent>

              <TabsContent value="notifications" className="space-y-4 pt-4">
                <div className="grid gap-4 md:grid-cols-2">
                  {([
                    ['productUpdates', 'Product updates'],
                    ['securityAlerts', 'Security alerts'],
                    ['billingAlerts', 'Billing alerts'],
                    ['marketingEmail', 'Marketing emails'],
                    ['weeklyDigest', 'Weekly digest'],
                  ] as const).map(([key, label]) => (
                    <div key={key} className="flex items-center justify-between rounded-md border p-3">
                      <Label>{label}</Label>
                      <Switch
                        checked={Boolean(notifications?.[key])}
                        onCheckedChange={(checked) =>
                          setNotifications((n) => (n ? { ...n, [key]: checked } : n))
                        }
                      />
                    </div>
                  ))}
                </div>
                <Button onClick={saveNotifications} disabled={saving}>Save Notifications</Button>
              </TabsContent>

              <TabsContent value="billing" className="space-y-4 pt-4">
                {billing ? (
                  <div className="grid gap-4 md:grid-cols-2">
                    <Card>
                      <CardHeader>
                        <CardTitle>Plan</CardTitle>
                        <CardDescription>{billing.plan}</CardDescription>
                      </CardHeader>
                    </Card>
                    <Card>
                      <CardHeader>
                        <CardTitle>Current Balance</CardTitle>
                        <CardDescription>${billing.currentBalance.toFixed(2)}</CardDescription>
                      </CardHeader>
                    </Card>
                    <Card>
                      <CardHeader>
                        <CardTitle>Monthly Spend</CardTitle>
                        <CardDescription>${billing.monthlySpend.toFixed(2)}</CardDescription>
                      </CardHeader>
                    </Card>
                    <Card>
                      <CardHeader>
                        <CardTitle>Next Billing</CardTitle>
                        <CardDescription>{new Date(billing.nextBillingDate).toLocaleDateString()}</CardDescription>
                      </CardHeader>
                    </Card>
                  </div>
                ) : (
                  <p className="text-sm text-muted-foreground">
                    Billing details are available for buyer accounts.
                  </p>
                )}
              </TabsContent>
            </Tabs>
          </CardContent>
        </Card>
      </div>
    </main>
  )
}

export const SettingsBlocks = {
  Page: SettingsPage,
}
