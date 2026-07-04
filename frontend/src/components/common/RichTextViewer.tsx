import { cn } from '@/lib/utils'

interface RichTextViewerProps {
  content: string
  className?: string
}

export default function RichTextViewer({ content, className }: RichTextViewerProps) {
  if (!content) return null

  return (
    <div
      className={cn('prose prose-sm max-w-none prose-headings:text-gray-800 prose-p:text-gray-600 prose-strong:text-gray-800 prose-ol:text-gray-600 prose-ul:text-gray-600 prose-li:marker:text-gray-400', className)}
      dangerouslySetInnerHTML={{ __html: content }}
    />
  )
}