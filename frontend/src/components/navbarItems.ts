export interface NavbarItem {
  label: string
  to: string
  icon: string
  requiresAuthentication: boolean
  guestOnly?: boolean
}

export const navbarItems: NavbarItem[] = [
  {
    label: 'Log in',
    to: '/login',
    icon: 'i-lucide-log-in',
    requiresAuthentication: false,
    guestOnly: true,
  },
  {
    label: 'Create account',
    to: '/signup',
    icon: 'i-lucide-user-plus',
    requiresAuthentication: false,
    guestOnly: true,
  },
  {
    label: 'Projects',
    to: '/projects',
    icon: 'i-lucide-folder-kanban',
    requiresAuthentication: true,
  },
  {
    label: 'Account',
    to: '/account',
    icon: 'i-lucide-circle-user-round',
    requiresAuthentication: true,
  },
]
