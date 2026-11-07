type ApiErrorResponse = {
  error: string
}

export type ApiRequestOptions<TBody = never> = Omit<RequestInit, 'body'> & {
  body?: TBody
  // Set for the auth endpoints, where a 401 means bad credentials or an
  // expired session rather than an access token that needs renewing.
  skipSessionRefresh?: boolean
}

export type SessionHandler = {
  // Renews the access token. Rejects when the session can't be renewed.
  refresh: () => Promise<unknown>
  // Called when the session can't be renewed, for example to go to login.
  expired: () => void
}

const defaultErrorMessage = 'Something went wrong. Please try again.'

let sessionHandler: SessionHandler | undefined
let sessionRefresh: Promise<boolean> | undefined

// Lets requests renew an expired access token. main.ts sets this, so this
// file doesn't need to import the auth service.
export function setSessionHandler(handler: SessionHandler) {
  sessionHandler = handler
}

// Several requests can fail at once when the access token expires; they all
// wait on the same refresh.
function refreshSession(handler: SessionHandler): Promise<boolean> {
  sessionRefresh ??= handler
    .refresh()
    .then(
      () => true,
      () => false,
    )
    .finally(() => {
      sessionRefresh = undefined
    })

  return sessionRefresh
}

function isApiErrorResponse(value: unknown): value is ApiErrorResponse {
  return (
    typeof value === 'object' &&
    value !== null &&
    'error' in value &&
    typeof value.error === 'string'
  )
}

function parseResponseBody(body: string, contentType: string): unknown {
  if (!contentType.includes('application/json') || !body) {
    return body
  }

  try {
    return JSON.parse(body) as unknown
  } catch {
    return body
  }
}

export async function apiRequest<TResponse, TBody = never>(
  endpoint: string,
  options: ApiRequestOptions<TBody> = {},
): Promise<TResponse> {
  const { body, headers: initialHeaders, skipSessionRefresh, ...init } = options
  const headers = new Headers(initialHeaders)

  if (body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const send = () =>
    fetch(`/api${endpoint}`, {
      ...init,
      credentials: 'include',
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    })

  let response = await send()

  // The access token has probably expired: renew it and try once more.
  if (response.status === 401 && !skipSessionRefresh && sessionHandler) {
    if (await refreshSession(sessionHandler)) {
      response = await send()
    } else {
      sessionHandler.expired()
    }
  }

  const responseText = await response.text()
  const responseBody = parseResponseBody(
    responseText,
    response.headers.get('content-type') ?? '',
  )

  if (!response.ok) {
    const message = isApiErrorResponse(responseBody)
      ? responseBody.error
      : typeof responseBody === 'string' && responseBody
        ? responseBody
        : responseText || defaultErrorMessage

    throw new Error(message)
  }

  return responseBody as TResponse
}
