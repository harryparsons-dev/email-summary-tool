type ApiErrorResponse = {
  error: string
}

export type ApiRequestOptions<TBody = never> = Omit<RequestInit, 'body'> & {
  body?: TBody
}

const defaultErrorMessage = 'Something went wrong. Please try again.'

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
  const { body, headers: initialHeaders, ...init } = options
  const headers = new Headers(initialHeaders)

  if (body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(`/api${endpoint}`, {
    ...init,
    credentials: 'include',
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  const responseText = await response.text()
  const responseBody = parseResponseBody(
    responseText,
    response.headers.get('content-type') ?? '',
  )

  if (!response.ok) {
    const message = isApiErrorResponse(responseBody)
      ? responseBody.error
      : responseText || defaultErrorMessage

    throw new Error(message)
  }

  return responseBody as TResponse
}
