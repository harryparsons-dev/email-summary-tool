export type AuthCredentials = {
  email: string
  password: string
}

async function getErrorMessage(response: Response): Promise<string> {
  const contentType = response.headers.get('content-type') ?? ''

  if (contentType.includes('application/json')) {
    const body = (await response.json()) as { error?: string }
    return body.error ?? 'Something went wrong. Please try again.'
  }

  return (await response.text()) || 'Something went wrong. Please try again.'
}

export async function authenticate(path: '/login' | '/signup', credentials: AuthCredentials) {
  const response = await fetch(`/api${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(credentials),
  })

  if (!response.ok) {
    throw new Error(await getErrorMessage(response))
  }
}
