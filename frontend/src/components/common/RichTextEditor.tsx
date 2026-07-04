import { EditorContent, useEditor } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import { Bold, Italic, List, ListOrdered } from 'lucide-react'
import { cn } from '@/lib/utils'

interface RichTextEditorProps {
  value: string
  onChange: (content: string) => void
}

export default function RichTextEditor({ value, onChange }: RichTextEditorProps) {
  const editor = useEditor({
    extensions: [StarterKit],
    content: value,
    immediatelyRender: false,
    onUpdate: ({ editor }) => {
      onChange(editor.getHTML())
    },
    editorProps: {
      attributes: {
        class: 'w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none min-h-32 focus-within:ring-3 focus-within:ring-ring/50',
      },
    },
  })

  if (!editor) return null

  return (
    <div>
      <div className="flex items-center gap-1 rounded-lg border border-input bg-gray-50/30 p-1.5 mb-2">
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleBold().run()}
          disabled={!editor.can().toggleBold()}
          className={cn('p-1.5 rounded hover:bg-gray-200 transition-colors', editor.isActive('bold') ? 'bg-gray-200' : '')}
        >
          <Bold className="w-4 h-4 text-gray-600" />
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleItalic().run()}
          disabled={!editor.can().toggleItalic()}
          className={cn('p-1.5 rounded hover:bg-gray-200 transition-colors', editor.isActive('italic') ? 'bg-gray-200' : '')}
        >
          <Italic className="w-4 h-4 text-gray-600" />
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleBulletList().run()}
          className={cn('p-1.5 rounded hover:bg-gray-200 transition-colors', editor.isActive('bulletList') ? 'bg-gray-200' : '')}
        >
          <List className="w-4 h-4 text-gray-600" />
        </button>
        <button
          type="button"
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
          className={cn('p-1.5 rounded hover:bg-gray-200 transition-colors', editor.isActive('orderedList') ? 'bg-gray-200' : '')}
        >
          <ListOrdered className="w-4 h-4 text-gray-600" />
        </button>
      </div>
      <EditorContent editor={editor} />
    </div>
  )
}