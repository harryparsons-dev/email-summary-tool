export interface EmailAttachment {
  filename: string
  content_type: string
  size: number
}

export interface Email {
  id: string
  project_id: string
  envelope_from: string
  envelope_to: string
  message_id: string
  from_address: string
  from_name: string
  to_addresses: string
  subject: string
  sent_at: string | null
  text_body: string
  html_body: string
  attachments: EmailAttachment[]
  received_at: string
}
