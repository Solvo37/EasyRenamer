import {
  ArrowDown,
  ArrowRight,
  ArrowUp,
  CaseUpper,
  Check,
  ChevronDown,
  ChevronRight,
  CircleHelp,
  Copy,
  Eraser,
  Eye,
  File,
  Folder,
  FileImage,
  FilePlus2,
  Film,
  FolderPlus,
  Hash,
  List,
  Minus,
  Moon,
  MoreHorizontal,
  Pencil,
  Play,
  RefreshCcw,
  Replace,
  Save,
  Search,
  Sun,
  Trash2,
  Undo2,
  X,
  Maximize2,
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { CSSProperties, PointerEvent as ReactPointerEvent, RefObject } from 'react'
import { appApi, runtimeApi } from './api'
import { methodCatalog, tagCatalog } from './catalog'
import type {
  BootstrapData,
  ExecuteProgress,
  HistoryEntry,
  PathClassification,
  PreviewItem,
  RenameMethod,
  SortMode,
  ThemeMode,
} from './types'

const categoryOptions = [
  { value: 'All files', key: 'category.all' },
  { value: 'Images', key: 'category.images' },
  { value: 'Videos', key: 'category.videos' },
  { value: 'Audio', key: 'category.audio' },
  { value: 'Documents', key: 'category.documents' },
  { value: 'Archives', key: 'category.archives' },
  { value: 'Custom', key: 'category.custom' },
]

const languageNames: Record<string, string> = {
  en: 'English',
  ru: 'Русский',
  es: 'Español',
  zh: '中文',
}

const presetOptions = [
  { labelKey: 'preset.sequence_original', value: '<Inc NrDir:01>_<Name>' },
  { labelKey: 'preset.original_sequence', value: '<Name>_<Inc:001>' },
  { labelKey: 'preset.parent_sequence', value: '<DirName:1>_<Inc NrDir:01>' },
  { labelKey: 'preset.date_original', value: '<Date:yyyyMMdd>_<Name>' },
]

const functionEntries = [
  ['lower(text)', 'lower(Name)'],
  ['upper(text)', 'upper(Name)'],
  ['trim(text)', 'trim(Name)'],
  ['replace(text, from, to)', "replace(Name, ' ', '-')"],
  ['concat(a, b, ...)', "concat(Name, '-', Index)"],
  ['substr(text, start, length)', 'substr(Name, 0, 8)'],
]

const variableEntries = [
  ['Name', 'File name without extension'],
  ['Ext', 'Extension with dot'],
  ['FullName', 'Full file name'],
  ['Index', 'Global index'],
  ['DirIndex', 'Per-folder index'],
  ['DirName', 'Parent directory'],
  ['UnixTimestamp', 'Batch Unix timestamp'],
  ['ModifiedUnix', 'Modified time Unix timestamp'],
]

const autocompleteTags = tagCatalog.flatMap((category) => category.items)

function methodIcon(kind: string) {
  const cls = 'method-icon-svg'
  switch (kind) {
    case 'replace':
      return <Replace className={cls} />
    case 'renumber':
      return <Hash className={cls} />
    case 'case':
      return <CaseUpper className={cls} />
    case 'remove':
    case 'remove_pattern':
      return <Eraser className={cls} />
    case 'move':
      return <ArrowRight className={cls} />
    case 'list':
    case 'list_replace':
      return <List className={cls} />
    case 'timestamp':
      return <RefreshCcw className={cls} />
    default:
      return <Pencil className={cls} />
  }
}

function fileIcon(type: string) {
  if (['JPG', 'JPEG', 'PNG', 'GIF', 'WEBP', 'BMP'].includes(type)) return <FileImage size={18} />
  if (['MP4', 'MOV', 'MKV', 'AVI', 'WEBM'].includes(type)) return <Film size={18} />
  return <File size={18} />
}

function formatBytes(size: number) {
  if (size >= 1024 ** 3) return `${(size / 1024 ** 3).toFixed(1)} GB`
  if (size >= 1024 ** 2) return `${(size / 1024 ** 2).toFixed(1)} MB`
  if (size >= 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${size} B`
}

function statusKey(status: string) {
  switch (status) {
    case 'OK':
      return 'item.status.ok'
    case 'Unchanged':
      return 'item.status.unchanged'
    case 'Conflict':
      return 'item.status.conflict'
    case 'Invalid':
      return 'item.status.invalid'
    default:
      return status
  }
}

function cloneMethod(method: RenameMethod): RenameMethod {
  return JSON.parse(JSON.stringify(method))
}

interface DropState extends PathClassification {
  open: boolean
}

type FileTableRow =
  | { kind: 'folder'; folder: string; group: PreviewItem[] }
  | { kind: 'file'; item: PreviewItem }

function App() {
  const [bootstrap, setBootstrap] = useState<BootstrapData | null>(null)
  const [strings, setStrings] = useState<Record<string, string>>({})
  const [language, setLanguage] = useState('ru')
  const [theme, setTheme] = useState<ThemeMode>(() => (localStorage.getItem('easyrenamer-theme') as ThemeMode) || 'system')
  const [systemDark, setSystemDark] = useState(window.matchMedia?.('(prefers-color-scheme: dark)').matches ?? true)

  const [sources, setSources] = useState<string[]>([])
  const [recursive, setRecursive] = useState(true)
  const [category, setCategory] = useState('All files')
  const [extensions, setExtensions] = useState('')
  const [methods, setMethods] = useState<RenameMethod[]>([])
  const [selectedMethod, setSelectedMethod] = useState(0)

  const [items, setItems] = useState<PreviewItem[]>([])
  const [checked, setChecked] = useState<Set<string>>(new Set())
  const [selectionTouched, setSelectionTouched] = useState(false)
  const [selectedPath, setSelectedPath] = useState('')
  const [thumbnail, setThumbnail] = useState('')
  const [previewBusy, setPreviewBusy] = useState(false)
  const [toast, setToast] = useState('')
  const [helpOpen, setHelpOpen] = useState(false)
  const [executeConfirm, setExecuteConfirm] = useState(false)
  const [dropState, setDropState] = useState<DropState>({ open: false, files: [], folders: [] })
  const [dropMode, setDropMode] = useState<'both' | 'files' | 'folders'>('both')
  const [compactView, setCompactView] = useState(false)

  const [sortBy, setSortBy] = useState<SortMode>(() => (localStorage.getItem('easyrenamer-sort') as SortMode) || 'name')
  const [sortDescending, setSortDescending] = useState(() => localStorage.getItem('easyrenamer-sort-desc') === '1')
  const [sortPerFolder, setSortPerFolder] = useState(() => localStorage.getItem('easyrenamer-sort-per-folder') !== '0')
  const [groupByFolder, setGroupByFolder] = useState(() => localStorage.getItem('easyrenamer-group-folders') !== '0')
  const [fileSearch, setFileSearch] = useState('')
  const [showErrorsOnly, setShowErrorsOnly] = useState(false)
  const [collapsedFolders, setCollapsedFolders] = useState<Set<string>>(new Set())
  const [methodsHeight, setMethodsHeight] = useState(() => Number(localStorage.getItem('easyrenamer-methods-height')) || 260)
  const [executing, setExecuting] = useState(false)
  const [executeProgress, setExecuteProgress] = useState<ExecuteProgress | null>(null)
  const [historyOpen, setHistoryOpen] = useState(false)
  const [historyItems, setHistoryItems] = useState<HistoryEntry[]>([])
  const [historyBusy, setHistoryBusy] = useState(false)
  const [tableScrollTop, setTableScrollTop] = useState(0)
  const [tableViewportHeight, setTableViewportHeight] = useState(640)

  const [tagTab, setTagTab] = useState<'tags' | 'functions' | 'variables'>('tags')
  const [tagCategory, setTagCategory] = useState(0)
  const [tagSearch, setTagSearch] = useState('')

  const templateInputRef = useRef<HTMLInputElement>(null)
  const previewTimer = useRef<number | null>(null)

  const t = useCallback((key: string, fallback?: string) => strings[key] || fallback || key, [strings])
  const ux = useMemo(() => language === 'ru' ? {
    order: 'Порядок', name: 'Имя', created: 'Дата создания', modified: 'Дата изменения',
    size: 'Размер', extension: 'Расширение', path: 'Путь', added: 'Порядок добавления', manual: 'Ручной порядок',
    perFolder: 'Сортировать отдельно внутри каждой папки', groupFolders: 'Группировка: по папкам',
    noGrouping: 'Группировка: нет', search: 'Имя или путь...', check: 'Проверить',
    errorsOnly: 'Только ошибки', allFiles: 'Все файлы', addMethod: '+ Метод', folders: 'Папок',
    cancelOperation: 'Отмена', preparing: 'Подготовка', renaming: 'Переименование файлов',
    cancelled: 'Операция отменена, исходные имена восстановлены',
    history: 'История', rollback: 'Откатить', rolledBack: 'Откат выполнен', undone: 'Откат выполнен',
    noHistory: 'История операций пока пуста', filesRenamed: 'Переименовано'
  } : {
    order: 'Order', name: 'Name', created: 'Created', modified: 'Modified',
    size: 'Size', extension: 'Extension', path: 'Path', added: 'Added order', manual: 'Manual order',
    perFolder: 'Sort separately inside each folder', groupFolders: 'Grouping: folders',
    noGrouping: 'Grouping: none', search: 'Name or path...', check: 'Check',
    errorsOnly: 'Errors only', allFiles: 'All files', addMethod: '+ Method', folders: 'Folders',
    cancelOperation: 'Cancel', preparing: 'Preparing', renaming: 'Renaming files',
    cancelled: 'Operation cancelled; original names restored',
    history: 'History', rollback: 'Rollback', rolledBack: 'Rollback complete', undone: 'Undone',
    noHistory: 'Operation history is empty', filesRenamed: 'Renamed'
  }, [language])

  const effectiveTheme = theme === 'system' ? (systemDark ? 'dark' : 'light') : theme

  useEffect(() => {
    document.documentElement.dataset.theme = effectiveTheme
    localStorage.setItem('easyrenamer-theme', theme)
  }, [effectiveTheme, theme])

  useEffect(() => {
    localStorage.setItem('easyrenamer-sort', sortBy)
    localStorage.setItem('easyrenamer-sort-desc', sortDescending ? '1' : '0')
    localStorage.setItem('easyrenamer-sort-per-folder', sortPerFolder ? '1' : '0')
    localStorage.setItem('easyrenamer-group-folders', groupByFolder ? '1' : '0')
  }, [sortBy, sortDescending, sortPerFolder, groupByFolder])

  useEffect(() => {
    const media = window.matchMedia?.('(prefers-color-scheme: dark)')
    if (!media) return
    const listener = (event: MediaQueryListEvent) => setSystemDark(event.matches)
    media.addEventListener?.('change', listener)
    return () => media.removeEventListener?.('change', listener)
  }, [])

  useEffect(() => {
    let alive = true
    ;(async () => {
      try {
        const data = await appApi().Bootstrap()
        if (!alive) return
        setBootstrap(data)
        setLanguage(data.language)
        setStrings(data.translations || {})
        const first = await appApi().NewMethod('template')
        if (alive) setMethods([first])
      } catch (err) {
        setToast(String(err))
      }
    })()
    return () => {
      alive = false
    }
  }, [])

  const addSources = useCallback((paths: string[]) => {
    if (!paths.length) return
    setSources((prev) => {
      const map = new Map(prev.map((path) => [path.toLowerCase(), path]))
      paths.forEach((path) => map.set(path.toLowerCase(), path))
      return [...map.values()]
    })
    setSelectionTouched(false)
  }, [])

  useEffect(() => {
    try {
      const runtime = runtimeApi()
      runtime.OnFileDrop(async (_x, _y, paths) => {
        try {
          const classified = await (window.go!.main!.App as any).ClassifyPaths(paths)
          if (classified.folders?.length) {
            setDropState({ open: true, files: classified.files || [], folders: classified.folders || [] })
          } else {
            addSources(classified.files || [])
          }
        } catch (err) {
          setToast(String(err))
        }
      }, true)
      return () => runtime.OnFileDropOff()
    } catch {
      return
    }
  }, [addSources])

  useEffect(() => {
    let off: (() => void) | undefined
    try {
      const runtime = runtimeApi()
      off = runtime.EventsOn?.('rename:progress', (progress: ExecuteProgress) => setExecuteProgress(progress))
    } catch {
      // Runtime is unavailable only during plain browser development.
    }
    return () => off?.()
  }, [])

  const requestPreview = useCallback(async () => {
    if (!sources.length || !methods.length) {
      setItems([])
      setChecked(new Set())
      return
    }
    setPreviewBusy(true)
    try {
      const result = await appApi().Preview(sources, recursive, category, extensions, methods, sortBy, sortDescending, sortPerFolder)
      const nextItems = result.items || []
      setItems(nextItems)
      setChecked((prev) => {
        const next = new Set<string>()
        for (const item of nextItems) {
          if (item.status !== 'OK') continue
          if (!selectionTouched || prev.has(item.sourcePath)) next.add(item.sourcePath)
        }
        return next
      })
      setSelectedPath((prev) => {
        if (nextItems.some((item) => item.sourcePath === prev)) return prev
        return nextItems[0]?.sourcePath || ''
      })
    } catch (err) {
      const message = String(err)
      if (!message.toLowerCase().includes('canceled') && !message.toLowerCase().includes('cancelled')) {
        setToast(message)
      }
    } finally {
      setPreviewBusy(false)
    }
  }, [sources, recursive, category, extensions, methods, selectionTouched, sortBy, sortDescending, sortPerFolder])

  useEffect(() => {
    if (previewTimer.current) window.clearTimeout(previewTimer.current)
    previewTimer.current = window.setTimeout(() => requestPreview(), 260)
    return () => {
      if (previewTimer.current) window.clearTimeout(previewTimer.current)
    }
  }, [requestPreview])

  const selectedItem = useMemo(
    () => items.find((item) => item.sourcePath === selectedPath) || items[0],
    [items, selectedPath]
  )

  useEffect(() => {
    let alive = true
    setThumbnail('')
    if (!selectedItem) return
    appApi()
      .Thumbnail(selectedItem.sourcePath)
      .then((data) => alive && setThumbnail(data || ''))
      .catch(() => {})
    return () => {
      alive = false
    }
  }, [selectedItem?.sourcePath])

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const ctrl = event.ctrlKey || event.metaKey
      const key = event.key.toLowerCase()
      if (ctrl && key === 'z') {
        event.preventDefault()
        handleUndo()
      } else if (ctrl && !event.shiftKey && key === 'o') {
        event.preventDefault()
        addFiles()
      } else if (ctrl && event.shiftKey && key === 'o') {
        event.preventDefault()
        addFolders()
      } else if (ctrl && key === 'a' && items.length) {
        event.preventDefault()
        setSelectionTouched(true)
        setChecked(new Set(validItems.map((item) => item.sourcePath)))
      } else if (ctrl && event.key === 'Enter' && checked.size && !errorCount && !executing) {
        event.preventDefault()
        setExecuteConfirm(true)
      } else if (event.key === 'F5') {
        event.preventDefault()
        requestPreview()
      } else if (event.key === 'Escape' && executing) {
        event.preventDefault()
        cancelExecution()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  })

  const handleLanguage = async (lang: string) => {
    try {
      const table = await appApi().SetLanguage(lang)
      setLanguage(lang)
      setStrings(table)
    } catch (err) {
      setToast(String(err))
    }
  }

  const cycleTheme = () => {
    setTheme((current) => (current === 'system' ? 'dark' : current === 'dark' ? 'light' : 'system'))
  }

  const addFiles = async () => {
    try {
      addSources(await appApi().PickFiles())
    } catch (err) {
      setToast(String(err))
    }
  }

  const addFolders = async () => {
    try {
      addSources(await appApi().PickFolders())
    } catch (err) {
      setToast(String(err))
    }
  }

  const clearAll = () => {
    setSources([])
    setItems([])
    setChecked(new Set())
    setSelectedPath('')
    setSelectionTouched(false)
  }

  const selectedMethodValue = methods[selectedMethod]

  const updateMethod = (patch: Partial<RenameMethod>) => {
    setMethods((prev) => prev.map((method, index) => (index === selectedMethod ? { ...method, ...patch } : method)))
  }

  const toggleMethod = (index: number) => {
    setMethods((prev) =>
      prev.map((method, i) => (i === index ? { ...method, disabled: !method.disabled } : method))
    )
  }

  const addMethod = async (kind: string) => {
    try {
      const method = await appApi().NewMethod(kind)
      setMethods((prev) => [...prev, method])
      setSelectedMethod(methods.length)
    } catch (err) {
      setToast(String(err))
    }
  }

  const duplicateMethod = () => {
    if (!selectedMethodValue) return
    setMethods((prev) => {
      const copy = [...prev]
      copy.splice(selectedMethod + 1, 0, cloneMethod(selectedMethodValue))
      return copy
    })
    setSelectedMethod((index) => index + 1)
  }

  const removeMethod = () => {
    if (!selectedMethodValue) return
    if (methods.length === 1) return
    setMethods((prev) => prev.filter((_, index) => index !== selectedMethod))
    setSelectedMethod((index) => Math.max(0, Math.min(index, methods.length - 2)))
  }

  const moveMethod = (direction: -1 | 1) => {
    const target = selectedMethod + direction
    if (target < 0 || target >= methods.length) return
    setMethods((prev) => {
      const copy = [...prev]
      ;[copy[selectedMethod], copy[target]] = [copy[target], copy[selectedMethod]]
      return copy
    })
    setSelectedMethod(target)
  }

  const saveMethods = async () => {
    try {
      await appApi().SaveMethodSet(methods)
      setToast(t('menu.save_methods'))
    } catch (err) {
      setToast(String(err))
    }
  }

  const loadMethods = async () => {
    try {
      const loaded = await appApi().LoadMethodSet()
      if (loaded?.length) {
        setMethods(loaded)
        setSelectedMethod(0)
      }
    } catch (err) {
      setToast(String(err))
    }
  }

  const toggleChecked = (path: string) => {
    setSelectionTouched(true)
    setChecked((prev) => {
      const next = new Set(prev)
      if (next.has(path)) next.delete(path)
      else next.add(path)
      return next
    })
  }

  const validItems = useMemo(() => items.filter((item) => item.status === 'OK'), [items])
  const errorCount = useMemo(
    () => items.filter((item) => item.status === 'Conflict' || item.status === 'Invalid').length,
    [items]
  )
  const folderCount = useMemo(() => new Set(items.map((item) => item.path.toLowerCase())).size, [items])
  const visibleItems = useMemo(() => {
    const query = fileSearch.trim().toLowerCase()
    return items.filter((item) => {
      if (showErrorsOnly && item.status !== 'Conflict' && item.status !== 'Invalid') return false
      if (!query) return true
      return item.oldName.toLowerCase().includes(query)
        || item.newName.toLowerCase().includes(query)
        || item.sourcePath.toLowerCase().includes(query)
    })
  }, [items, fileSearch, showErrorsOnly])
  const groupedItems = useMemo(() => {
    const groups = new Map<string, PreviewItem[]>()
    for (const item of visibleItems) {
      const group = groups.get(item.path) || []
      group.push(item)
      groups.set(item.path, group)
    }
    return groups
  }, [visibleItems])

  const tableRows = useMemo<FileTableRow[]>(() => {
    if (!groupByFolder) return visibleItems.map((item) => ({ kind: 'file', item }))
    const rows: FileTableRow[] = []
    for (const [folder, group] of groupedItems.entries()) {
      rows.push({ kind: 'folder', folder, group })
      if (!collapsedFolders.has(folder)) {
        for (const item of group) rows.push({ kind: 'file', item })
      }
    }
    return rows
  }, [visibleItems, groupedItems, groupByFolder, collapsedFolders])

  const virtualRowHeight = compactView ? 30 : 36
  const virtualOverscan = 14
  const virtualStart = Math.max(0, Math.floor(tableScrollTop / virtualRowHeight) - virtualOverscan)
  const virtualCount = Math.ceil(tableViewportHeight / virtualRowHeight) + virtualOverscan * 2
  const virtualEnd = Math.min(tableRows.length, virtualStart + virtualCount)
  const virtualRows = tableRows.slice(virtualStart, virtualEnd)
  const virtualTop = virtualStart * virtualRowHeight
  const virtualBottom = Math.max(0, (tableRows.length - virtualEnd) * virtualRowHeight)

  useEffect(() => {
    setTableScrollTop(0)
  }, [fileSearch, showErrorsOnly, groupByFolder, compactView])

  const execute = async () => {
    setExecuteConfirm(false)
    setExecuting(true)
    setExecuteProgress({ phase: 'staging', completed: 0, total: checked.size, current: '' })
    try {
      const result = await appApi().Execute([...checked])
      if (result.cancelled) {
        setToast(ux.cancelled)
        await requestPreview()
        return
      }
      if (result.pairs?.length) {
        setSources((prev) =>
          prev.map((source) => result.pairs!.find((pair) => pair.from.toLowerCase() === source.toLowerCase())?.to || source)
        )
      }
      setToast(t('dialog.renamed').replace('%d', String(result.count)))
      setSelectionTouched(false)
      await requestPreview()
    } catch (err) {
      setToast(String(err))
    } finally {
      setExecuting(false)
      setExecuteProgress(null)
    }
  }

  const cancelExecution = async () => {
    try {
      await appApi().CancelExecute()
    } catch (err) {
      setToast(String(err))
    }
  }

  async function handleUndo() {
    try {
      const result = await appApi().Undo()
      if (result.pairs?.length) {
        setSources((prev) =>
          prev.map((source) => result.pairs!.find((pair) => pair.to.toLowerCase() === source.toLowerCase())?.from || source)
        )
      }
      setToast(t('dialog.undone'))
      setSelectionTouched(false)
      await requestPreview()
    } catch (err) {
      setToast(String(err))
    }
  }

  const loadHistory = async () => {
    setHistoryBusy(true)
    try {
      setHistoryItems(await appApi().History())
    } catch (err) {
      setToast(String(err))
    } finally {
      setHistoryBusy(false)
    }
  }

  const openHistory = async () => {
    setHistoryOpen(true)
    await loadHistory()
  }

  const rollbackHistory = async (id: string) => {
    setHistoryBusy(true)
    try {
      const result = await appApi().UndoHistory(id)
      if (result.pairs?.length) {
        setSources((prev) =>
          prev.map((source) => result.pairs!.find((pair) => pair.to.toLowerCase() === source.toLowerCase())?.from || source)
        )
      }
      setToast(ux.rolledBack)
      setSelectionTouched(false)
      await requestPreview()
      setHistoryItems(await appApi().History())
    } catch (err) {
      setToast(String(err))
    } finally {
      setHistoryBusy(false)
    }
  }

  const insertTag = (token: string) => {
    if (!selectedMethodValue || selectedMethodValue.type !== 'template') return
    const input = templateInputRef.current
    const value = selectedMethodValue.template || ''
    if (!input) {
      updateMethod({ template: value + token })
      return
    }
    const start = input.selectionStart ?? value.length
    const end = input.selectionEnd ?? start
    updateMethod({ template: value.slice(0, start) + token + value.slice(end) })
    requestAnimationFrame(() => {
      input.focus()
      input.setSelectionRange(start + token.length, start + token.length)
    })
  }

  const filteredTags = useMemo(() => {
    const categoryTags = tagCatalog[tagCategory]?.items || []
    const query = tagSearch.trim().toLowerCase()
    if (!query) return categoryTags
    return categoryTags.filter(
      (tag) => tag.token.toLowerCase().includes(query) || t(tag.labelKey).toLowerCase().includes(query)
    )
  }, [tagCategory, tagSearch, t])

  const applyDrop = () => {
    const paths =
      dropMode === 'files'
        ? dropState.files
        : dropMode === 'folders'
          ? dropState.folders
          : [...dropState.files, ...dropState.folders]
    addSources(paths)
    setDropState({ open: false, files: [], folders: [] })
  }

  const startSidebarResize = (event: ReactPointerEvent<HTMLDivElement>) => {
    event.preventDefault()
    const startY = event.clientY
    const startHeight = methodsHeight
    let latestHeight = startHeight
    const move = (moveEvent: PointerEvent) => {
      latestHeight = Math.max(118, Math.min(460, startHeight + moveEvent.clientY - startY))
      setMethodsHeight(latestHeight)
    }
    const stop = () => {
      window.removeEventListener('pointermove', move)
      window.removeEventListener('pointerup', stop)
      localStorage.setItem('easyrenamer-methods-height', String(Math.round(latestHeight)))
    }
    window.addEventListener('pointermove', move)
    window.addEventListener('pointerup', stop)
  }

  const toggleFolder = (folder: string) => {
    setCollapsedFolders((prev) => {
      const next = new Set(prev)
      if (next.has(folder)) next.delete(folder)
      else next.add(folder)
      return next
    })
  }

  const renderFileRow = (item: PreviewItem) => (
    <tr
      key={item.sourcePath}
      className={selectedItem?.sourcePath === item.sourcePath ? 'selected' : ''}
      onClick={() => setSelectedPath(item.sourcePath)}
      onDoubleClick={() => appApi().Reveal(item.sourcePath)}
      title={item.error || item.sourcePath}
    >
      <td className="checkcol" onClick={(e) => e.stopPropagation()}>
        <input
          type="checkbox"
          disabled={item.status !== 'OK' || executing}
          checked={checked.has(item.sourcePath)}
          onChange={() => toggleChecked(item.sourcePath)}
        />
      </td>
      <td>{item.globalIndex}</td>
      <td><span className="file-name">{fileIcon(item.type)}{item.oldName}</span></td>
      <td className="new-name">{item.newName}</td>
      <td className={`path-cell path-column ${groupByFolder ? 'hidden' : ''}`} title={item.path}>{item.path}</td>
      <td>{formatBytes(item.size)}</td>
      <td>{item.type || 'FILE'}</td>
      <td title={item.error || ''}>
        <span className={`status-pill status-${item.status.toLowerCase()}`}>
          <i />{t(statusKey(item.status), item.status)}
        </span>
      </td>
    </tr>
  )

  const renderFolderRow = (folder: string, group: PreviewItem[]) => {
    const folderValid = group.filter((item) => item.status === 'OK')
    const folderChecked = folderValid.length > 0 && folderValid.every((item) => checked.has(item.sourcePath))
    const collapsed = collapsedFolders.has(folder)
    return (
      <tr className="folder-group-row" key={`folder:${folder}`} onClick={() => toggleFolder(folder)}>
        <td className="checkcol" onClick={(e) => e.stopPropagation()}>
          <input
            type="checkbox"
            disabled={!folderValid.length || executing}
            checked={folderChecked}
            onChange={(e) => {
              setSelectionTouched(true)
              setChecked((prev) => {
                const next = new Set(prev)
                for (const item of folderValid) {
                  if (e.target.checked) next.add(item.sourcePath)
                  else next.delete(item.sourcePath)
                }
                return next
              })
            }}
          />
        </td>
        <td colSpan={7}>
          <span className="folder-group-copy" title={folder}>
            <ChevronDown className={collapsed ? 'collapsed' : ''} />
            <Folder />
            <strong>{folder}</strong>
            <em>{group.length}</em>
          </span>
        </td>
      </tr>
    )
  }

  return (
    <div className="app-shell">
      <header className="titlebar" style={{ '--wails-draggable': 'drag' } as CSSProperties}>
        <div className="brand">
          <div className="brand-mark">ER</div>
          <strong>EasyRenamer <span>{bootstrap?.version || 'dev'}</span></strong>
          <span className="brand-subtitle">{t('app.subtitle', 'Batch file renaming')}</span>
        </div>
        <div className="tagline">{t('app.tagline', 'Order in names — more order in work')}</div>
        <div className="window-controls" style={{ '--wails-draggable': 'no-drag' } as CSSProperties}>
          <button onClick={() => runtimeApi().WindowMinimise()}><Minus size={17} /></button>
          <button onClick={() => runtimeApi().WindowToggleMaximise()}><Maximize2 size={15} /></button>
          <button className="close" onClick={() => runtimeApi().Quit()}><X size={18} /></button>
        </div>
      </header>

      <section className="commandbar">
        <div className="command-left">
          <button className={`action ${items.length ? '' : 'primary'}`} disabled={executing} onClick={addFiles} title="Ctrl+O"><FilePlus2 />{t('button.files')}</button>
          <button className="action" disabled={executing} onClick={addFolders} title="Ctrl+Shift+O"><FolderPlus />{t('button.folders')}</button>
          <button className="action danger-soft" disabled={executing} onClick={clearAll}><Trash2 />{t('button.clear')}</button>
          <button className="action" disabled={executing || previewBusy || !sources.length} onClick={requestPreview} title="F5">
            <Eye />{ux.check}
          </button>
        </div>
        <div className="command-right">
          <button
            className={`start-btn ${items.length && !errorCount && checked.size ? 'primary' : ''}`}
            disabled={!checked.size || previewBusy || executing || errorCount > 0}
            onClick={() => setExecuteConfirm(true)}
            title="Ctrl+Enter"
          >
            <Play fill="currentColor" />{t('button.start')}
          </button>
          <span className="source-count">{t('sources.count').replace('%d', String(sources.length))}</span>
          <select value={language} onChange={(e) => handleLanguage(e.target.value)}>
            {Object.entries(languageNames).map(([code, name]) => <option key={code} value={code}>{name}</option>)}
          </select>
          <button className="icon-btn" title={theme} onClick={cycleTheme}>
            {effectiveTheme === 'dark' ? <Sun /> : <Moon />}
          </button>
          <button className="action help-btn" onClick={() => setHelpOpen(true)}><CircleHelp />{t('button.help')}</button>
        </div>
      </section>

      <section className="filterbar">
        <label>{t('filter.label')}</label>
        <select value={category} onChange={(e) => setCategory(e.target.value)}>
          {categoryOptions.map((option) => <option key={option.value} value={option.value}>{t(option.key)}</option>)}
        </select>
        <label className="checkline">
          <input type="checkbox" checked={recursive} onChange={(e) => setRecursive(e.target.checked)} />
          <span>{t('filter.subfolders')}</span>
        </label>
        <label>{t('filter.extensions')}</label>
        <input
          className="extensions"
          value={extensions}
          onChange={(e) => {
            setExtensions(e.target.value)
            if (e.target.value.trim()) setCategory('Custom')
          }}
          placeholder={t('filter.extensions_hint')}
        />
        <label>{t('collision.label')}</label>
        <select className="collision" value="prevent" disabled><option value="prevent">{t('collision.prevent')}</option></select>
        <button className="mini-more" onClick={openHistory} title={ux.history}><MoreHorizontal /></button>
      </section>

      <main className="workspace">
        <aside className="sidebar" style={{ '--methods-height': `${methodsHeight}px` } as CSSProperties}>
          <section className="panel methods-panel">
            <div className="panel-title method-panel-title">
              <span className="panel-title-copy"><b>✦</b>{t('group.methods')}</span>
              <select
                className="method-add-select"
                value=""
                onChange={(e) => {
                  if (e.target.value) addMethod(e.target.value)
                }}
                title={ux.addMethod}
              >
                <option value="">{ux.addMethod}</option>
                {methodCatalog.map((entry) => <option key={entry.type} value={entry.type}>{t(entry.titleKey)}</option>)}
              </select>
            </div>
            <div className="method-list">
              {methods.map((method, index) => {
                const def = methodCatalog.find((entry) => entry.type === method.type)
                return (
                  <button
                    key={index}
                    className={`method-card ${selectedMethod === index ? 'active' : ''} ${method.disabled ? 'disabled' : ''}`}
                    onClick={() => setSelectedMethod(index)}
                  >
                    <span className="method-active-toggle" onClick={(event) => { event.stopPropagation(); toggleMethod(index) }}>
                      {method.disabled ? '' : <Check size={13} />}
                    </span>
                    <span className="method-icon">{methodIcon(method.type)}</span>
                    <span className="method-copy">
                      <strong>{t(def?.titleKey || `method.${method.type}`, method.type)}</strong>
                      <small>{t(def?.descriptionKey || '', '')}</small>
                    </span>
                    <ChevronRight size={18} />
                  </button>
                )
              })}
            </div>
            <div className="method-toolbar">
              <button title={t('toolbar.move_up', 'Переместить выше')} onClick={() => moveMethod(-1)} disabled={selectedMethod === 0}><ArrowUp /></button>
              <button title={t('toolbar.move_down', 'Переместить ниже')} onClick={() => moveMethod(1)} disabled={selectedMethod >= methods.length - 1}><ArrowDown /></button>
              <button title={t('toolbar.duplicate', 'Дублировать метод')} onClick={duplicateMethod}><Copy /></button>
              <button title={t('menu.save_methods')} onClick={saveMethods}><Save /></button>
              <button title={t('menu.load_methods')} onClick={loadMethods}><RefreshCcw /></button>
              <button title={t('toolbar.delete', 'Удалить метод')} onClick={removeMethod} disabled={methods.length === 1}><Trash2 /></button>
            </div>
          </section>

          <div className="sidebar-splitter" onPointerDown={startSidebarResize} title="Drag to resize" />

          <section className="panel settings-panel">
            <h2>{t('group.settings')}: {selectedMethodValue ? t(methodCatalog.find((x) => x.type === selectedMethodValue.type)?.titleKey || '') : ''}</h2>
            {selectedMethodValue && (
              <MethodEditor
                method={selectedMethodValue}
                update={updateMethod}
                t={t}
                templateInputRef={templateInputRef}
                tagTab={tagTab}
                setTagTab={setTagTab}
                tagCategory={tagCategory}
                setTagCategory={setTagCategory}
                tagSearch={tagSearch}
                setTagSearch={setTagSearch}
                filteredTags={filteredTags}
                insertTag={insertTag}
              />
            )}
          </section>
        </aside>

        <section className="files-panel">
          <div className="files-head">
            <div className="files-title-line">
              <h1>{t('files.title_count').replace('%d', String(items.length))}</h1>
              <div className="drop-hint"><FolderPlus size={18} />{t('drop.hint')} · {t('drop.subhint')}</div>
            </div>
            <div className="file-tools">
              <label className="table-search" title={ux.search}>
                <Search size={16} />
                <input value={fileSearch} onChange={(e) => setFileSearch(e.target.value)} placeholder={ux.search} />
              </label>
              <label className="sort-control">
                <span>{ux.order}:</span>
                <select value={sortBy} onChange={(e) => setSortBy(e.target.value as SortMode)}>
                  <option value="name">{ux.name}</option>
                  <option value="created">{ux.created}</option>
                  <option value="modified">{ux.modified}</option>
                  <option value="size">{ux.size}</option>
                  <option value="extension">{ux.extension}</option>
                  <option value="path">{ux.path}</option>
                  <option value="added">{ux.added}</option>
                  <option value="manual">{ux.manual}</option>
                </select>
              </label>
              <button className="sort-direction" onClick={() => setSortDescending((value) => !value)} title={sortDescending ? 'Descending' : 'Ascending'}>
                {sortDescending ? <ArrowDown /> : <ArrowUp />}
              </button>
              <label className="checkline sort-per-folder">
                <input type="checkbox" checked={sortPerFolder} onChange={(e) => setSortPerFolder(e.target.checked)} />
                <span>{ux.perFolder}</span>
              </label>
              <select className="group-select" value={groupByFolder ? 'folder' : 'none'} onChange={(e) => setGroupByFolder(e.target.value === 'folder')}>
                <option value="folder">{ux.groupFolders}</option>
                <option value="none">{ux.noGrouping}</option>
              </select>
              {errorCount > 0 && (
                <button className={`errors-toggle ${showErrorsOnly ? 'active' : ''}`} onClick={() => setShowErrorsOnly((value) => !value)}>
                  {ux.errorsOnly} ({errorCount})
                </button>
              )}
              <div className="view-toggle">
                <button title="Comfortable rows" className={!compactView ? 'active' : ''} onClick={() => setCompactView(false)}><List /></button>
                <button title="Compact rows" className={compactView ? 'active' : ''} onClick={() => setCompactView(true)}>▦</button>
              </div>
            </div>
          </div>

          <div
            className="table-wrap"
            style={{ '--wails-drop-target': 'drop' } as CSSProperties}
            onScroll={(e) => {
              setTableScrollTop(e.currentTarget.scrollTop)
              setTableViewportHeight(e.currentTarget.clientHeight)
            }}
          >
            <table className={compactView ? 'compact' : ''}>
              <thead>
                <tr>
                  <th className="checkcol">
                    <input
                      type="checkbox"
                      checked={validItems.length > 0 && checked.size === validItems.length}
                      onChange={(e) => {
                        setSelectionTouched(true)
                        setChecked(e.target.checked ? new Set(validItems.map((item) => item.sourcePath)) : new Set())
                      }}
                    />
                  </th>
                  <th>#</th>
                  <th>{t('column.filename')}</th>
                  <th>{t('column.new_filename')}</th>
                  <th className={`path-column ${groupByFolder ? 'hidden' : ''}`}>{t('column.path')}</th>
                  <th>{t('column.size')}</th>
                  <th>{t('column.type')}</th>
                  <th>{t('column.status')}</th>
                </tr>
              </thead>
              <tbody>
                {virtualTop > 0 && (
                  <tr className="virtual-spacer"><td colSpan={8} style={{ height: virtualTop }} /></tr>
                )}
                {virtualRows.map((row) =>
                  row.kind === 'folder'
                    ? renderFolderRow(row.folder, row.group)
                    : renderFileRow(row.item)
                )}
                {virtualBottom > 0 && (
                  <tr className="virtual-spacer"><td colSpan={8} style={{ height: virtualBottom }} /></tr>
                )}
                {!visibleItems.length && (
                  <tr className="empty-row">
                    <td colSpan={8}>
                      <div className="empty-state">
                        <FolderPlus />
                        <strong>{items.length ? ux.allFiles : t('drop.hint')}</strong>
                        <span>{items.length ? ux.search : t('drop.subhint')}</span>
                      </div>
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          <section className="file-preview">
            <div className="preview-thumb">
              {thumbnail ? <img src={thumbnail} alt="" /> : selectedItem ? fileIcon(selectedItem.type) : <FileImage />}
            </div>
            <div className="preview-meta">
              <strong>{selectedItem?.oldName || t('preview.no_selection')}</strong>
              {selectedItem && (
                <>
                  <span>{selectedItem.width && selectedItem.height ? `${selectedItem.width} × ${selectedItem.height} · ` : ''}{formatBytes(selectedItem.size)} · {selectedItem.type}</span>
                  <span>{selectedItem.path}</span>
                </>
              )}
            </div>
            <div className="name-compare">
              <div className="compare-values">
                <strong title={selectedItem?.oldName || ''}>{selectedItem?.oldName || '—'}</strong>
                <ArrowRight />
                <strong title={selectedItem?.newName || ''}>{selectedItem?.newName || '—'}</strong>
              </div>
            </div>
          </section>
        </section>
      </main>

      <footer className="statusbar">
        <span className={`state ${executing ? 'busy' : errorCount ? 'error' : ''}`}><i />{executing ? ux.renaming : errorCount ? t('status.errors') : t('status.ready')}</span>
        <div>
          <span>{t('button.select_valid')}: {checked.size} {t('status.of')} {validItems.length}</span>
          <span>{t('group.files')}: {items.length}</span>
          <span>{ux.folders}: {folderCount}</span>
          <span>{t('status.errors')}: {errorCount}</span>
        </div>
      </footer>

      {executing && (
        <div className="operation-progress">
          <div className="operation-progress-head">
            <strong>{executeProgress?.phase === 'renaming' ? ux.renaming : ux.preparing}</strong>
            <span>{executeProgress?.completed || 0} / {executeProgress?.total || checked.size}</span>
          </div>
          <progress max={Math.max(1, executeProgress?.total || checked.size)} value={executeProgress?.completed || 0} />
          <small title={executeProgress?.current || ''}>{executeProgress?.current || '...'}</small>
          <button onClick={cancelExecution}>{ux.cancelOperation}</button>
        </div>
      )}

      {historyOpen && (
        <div className="modal-backdrop" onMouseDown={() => setHistoryOpen(false)}>
          <div className="modal history-modal" onMouseDown={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setHistoryOpen(false)}><X /></button>
            <h2>{ux.history}</h2>
            <div className="history-list">
              {historyBusy && !historyItems.length && <div className="history-empty">...</div>}
              {!historyBusy && !historyItems.length && <div className="history-empty">{ux.noHistory}</div>}
              {historyItems.map((entry) => (
                <div className={`history-entry ${entry.undone ? 'undone' : ''}`} key={entry.id}>
                  <div className="history-entry-main">
                    <strong>{entry.createdAt}</strong>
                    <span>{ux.filesRenamed}: {entry.count}</span>
                    <small title={entry.folder}>{entry.folder || '—'}</small>
                  </div>
                  <button
                    disabled={entry.undone || historyBusy || executing}
                    onClick={() => rollbackHistory(entry.id)}
                  >
                    <Undo2 />{entry.undone ? ux.undone : ux.rollback}
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {helpOpen && (
        <div className="modal-backdrop" onMouseDown={() => setHelpOpen(false)}>
          <div className="modal help-modal" onMouseDown={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setHelpOpen(false)}><X /></button>
            <h2>{t('help.title')}</h2>
            <p className="preline">{t('help.guide_body')}</p>
          </div>
        </div>
      )}

      {executeConfirm && (
        <div className="modal-backdrop">
          <div className="modal confirm-modal">
            <h2>{t('dialog.confirm_rename_title')}</h2>
            <p>{t('dialog.confirm_rename_body').replace('%d', String(checked.size))}</p>
            <div className="modal-actions">
              <button onClick={() => setExecuteConfirm(false)}>{t('button.cancel')}</button>
              <button className="primary" onClick={execute}><Play />{t('button.start')}</button>
            </div>
          </div>
        </div>
      )}

      {dropState.open && (
        <div className="modal-backdrop">
          <div className="modal drop-modal">
            <h2>{t('drop.dialog_title')}</h2>
            <p>{t('drop.dialog_info')}</p>
            <label><input type="radio" name="dropmode" checked={dropMode === 'both'} onChange={() => setDropMode('both')} />{t('drop.mode_all')}</label>
            <label><input type="radio" name="dropmode" checked={dropMode === 'files'} onChange={() => setDropMode('files')} />{t('drop.mode_files')} ({dropState.files.length})</label>
            <label><input type="radio" name="dropmode" checked={dropMode === 'folders'} onChange={() => setDropMode('folders')} />{t('drop.mode_folders')} ({dropState.folders.length})</label>
            <label className="checkline"><input type="checkbox" checked={recursive} onChange={(e) => setRecursive(e.target.checked)} />{t('drop.include_subfolders')}</label>
            <div className="modal-actions">
              <button onClick={() => setDropState({ open: false, files: [], folders: [] })}>{t('button.cancel')}</button>
              <button className="primary" onClick={applyDrop}>{t('button.add')}</button>
            </div>
          </div>
        </div>
      )}

      {toast && <div className="toast" onClick={() => setToast('')}>{toast}</div>}
    </div>
  )
}

interface MethodEditorProps {
  method: RenameMethod
  update: (patch: Partial<RenameMethod>) => void
  t: (key: string, fallback?: string) => string
  templateInputRef: RefObject<HTMLInputElement>
  tagTab: 'tags' | 'functions' | 'variables'
  setTagTab: (tab: 'tags' | 'functions' | 'variables') => void
  tagCategory: number
  setTagCategory: (index: number) => void
  tagSearch: string
  setTagSearch: (value: string) => void
  filteredTags: { token: string; labelKey: string }[]
  insertTag: (token: string) => void
}

function MethodEditor(props: MethodEditorProps) {
  const { method, update, t } = props
  const [autocompleteOpen, setAutocompleteOpen] = useState(false)
  const [autocompleteQuery, setAutocompleteQuery] = useState('')
  const [autocompleteIndex, setAutocompleteIndex] = useState(0)
  const [guideSelected, setGuideSelected] = useState('')

  const autocompleteMatches = useMemo(() => {
    const query = autocompleteQuery.toLowerCase()
    return autocompleteTags
      .filter((entry) => !query || entry.token.toLowerCase().startsWith('<' + query) || entry.token.toLowerCase().includes(query))
      .slice(0, 10)
  }, [autocompleteQuery])

  const updateAutocomplete = (value: string, caret: number | null) => {
    const position = caret ?? value.length
    const before = value.slice(0, position)
    const open = before.lastIndexOf('<')
    if (open < 0 || before.slice(open).includes('>')) {
      setAutocompleteOpen(false)
      return
    }
    const query = before.slice(open + 1)
    setAutocompleteQuery(query)
    setAutocompleteIndex(0)
    setAutocompleteOpen(true)
  }

  const acceptAutocomplete = (token: string) => {
    const input = props.templateInputRef.current
    const value = method.template || ''
    const caret = input?.selectionStart ?? value.length
    const open = value.slice(0, caret).lastIndexOf('<')
    if (open < 0) {
      props.insertTag(token)
      setAutocompleteOpen(false)
      return
    }
    const next = value.slice(0, open) + token + value.slice(caret)
    const nextCaret = open + token.length
    update({ template: next })
    setAutocompleteOpen(false)
    requestAnimationFrame(() => {
      input?.focus()
      input?.setSelectionRange(nextCaret, nextCaret)
    })
  }

  const number = (key: keyof RenameMethod, value = 1, min?: number) => (
    <input type="number" min={min} value={(method[key] as number | undefined) ?? value} onChange={(e) => update({ [key]: Number(e.target.value) } as Partial<RenameMethod>)} />
  )

  if (method.type === 'template') {
    return (
      <div className="editor">
        <div className="field-row">
          <label>{t('label.preset')}</label>
          <select
            value={presetOptions.find((p) => p.value === method.template)?.value || ''}
            onChange={(e) => e.target.value && update({ template: e.target.value })}
          >
            <option value="">{t('preset.sequence_original')}</option>
            {presetOptions.map((preset) => <option key={preset.value} value={preset.value}>{t(preset.labelKey)}</option>)}
          </select>
        </div>
        <div className="field-row template-field">
          <label>{t('label.new_name')}</label>
          <div className="template-input-wrap">
            <input
              ref={props.templateInputRef}
              value={method.template || ''}
              onChange={(e) => {
                update({ template: e.target.value })
                updateAutocomplete(e.target.value, e.target.selectionStart)
              }}
              onKeyDown={(e) => {
                if (!autocompleteOpen || !autocompleteMatches.length) return
                if (e.key === 'ArrowDown') {
                  e.preventDefault()
                  setAutocompleteIndex((index) => (index + 1) % autocompleteMatches.length)
                } else if (e.key === 'ArrowUp') {
                  e.preventDefault()
                  setAutocompleteIndex((index) => (index - 1 + autocompleteMatches.length) % autocompleteMatches.length)
                } else if (e.key === 'Enter') {
                  e.preventDefault()
                  acceptAutocomplete(autocompleteMatches[Math.min(autocompleteIndex, autocompleteMatches.length - 1)].token)
                } else if (e.key === 'Escape') {
                  e.preventDefault()
                  setAutocompleteOpen(false)
                }
              }}
              onBlur={() => window.setTimeout(() => setAutocompleteOpen(false), 120)}
            />
            {autocompleteOpen && autocompleteMatches.length > 0 && (
              <div className="tag-autocomplete">
                {autocompleteMatches.map((entry, index) => (
                  <button
                    type="button"
                    key={entry.token}
                    className={index === autocompleteIndex ? 'active' : ''}
                    onMouseDown={(e) => e.preventDefault()}
                    onClick={() => acceptAutocomplete(entry.token)}
                  >
                    <code>{entry.token}</code>
                    <span>{t(entry.labelKey)}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
        <div className="field-row">
          <label>{t('tag.category')}</label>
          <select value={props.tagCategory} onChange={(e) => props.setTagCategory(Number(e.target.value))}>
            {tagCatalog.map((category, index) => <option value={index} key={category.key}>{t(category.key)}</option>)}
          </select>
        </div>
        <div className="tag-tabs">
          <button className={props.tagTab === 'tags' ? 'active' : ''} onClick={() => props.setTagTab('tags')}>{t('help.tags')}</button>
          <button className={props.tagTab === 'functions' ? 'active' : ''} onClick={() => props.setTagTab('functions')}>{t('label.functions_help').split(':')[0]}</button>
          <button className={props.tagTab === 'variables' ? 'active' : ''} onClick={() => props.setTagTab('variables')}>{t('label.variables_help').split(':')[0]}</button>
        </div>
        <div className="tag-search"><Search /><input value={props.tagSearch} onChange={(e) => props.setTagSearch(e.target.value)} placeholder={t('tag.search')} /></div>
        <div className="tag-grid">
          {props.tagTab === 'tags' && props.filteredTags.map((tag) => (
            <button
              key={tag.token}
              className={guideSelected === tag.token ? 'selected' : ''}
              title={`${tag.token} — ${t(tag.labelKey)}`}
              onClick={() => setGuideSelected(tag.token)}
              onDoubleClick={() => props.insertTag(tag.token)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault()
                  props.insertTag(tag.token)
                }
              }}
            >
              <code>{tag.token}</code><span>{t(tag.labelKey)}</span>
            </button>
          ))}
          {props.tagTab === 'functions' && functionEntries.map(([name, example]) => <button key={name} onClick={() => props.insertTag(example)}><code>{name}</code><span>{example}</span></button>)}
          {props.tagTab === 'variables' && variableEntries.map(([name, description]) => <button key={name} onClick={() => props.insertTag(name)}><code>{name}</code><span>{description}</span></button>)}
        </div>
        <div className="editor-hint">ⓘ {t('tag.hint')}</div>
      </div>
    )
  }

  if (method.type === 'replace') return (
    <div className="editor two-col">
      <label>{t('label.find')}</label><input value={method.find || ''} onChange={(e) => update({ find: e.target.value })} />
      <label>{t('label.replace_with')}</label><input value={method.replace_with || ''} onChange={(e) => update({ replace_with: e.target.value })} />
      <label className="wide checkline"><input type="checkbox" checked={!!method.use_regex} onChange={(e) => update({ use_regex: e.target.checked })} />{t('label.regex')}</label>
    </div>
  )

  if (method.type === 'renumber') return (
    <div className="editor two-col">
      <label>{t('label.start')}</label>{number('renumber_start', 1)}
      <label>{t('label.step')}</label>{number('renumber_step', 1)}
      <label>{t('label.padding')}</label>{number('renumber_padding', 2, 1)}
      <label>{t('label.separator')}</label><input value={method.renumber_separator || ''} onChange={(e) => update({ renumber_separator: e.target.value })} />
      <label>{t('label.position')}</label><select value={method.renumber_position || 'prefix'} onChange={(e) => update({ renumber_position: e.target.value as 'prefix' | 'suffix' })}><option value="prefix">{t('label.prefix_position')}</option><option value="suffix">{t('label.suffix_position')}</option></select>
      <label className="wide checkline"><input type="checkbox" checked={!!method.renumber_per_dir} onChange={(e) => update({ renumber_per_dir: e.target.checked })} />{t('label.per_folder')}</label>
    </div>
  )

  if (method.type === 'case') return (
    <div className="editor two-col">
      <label>{t('label.case_to')}</label>
      <select value={method.case_mode || 'lower'} onChange={(e) => update({ case_mode: e.target.value as 'lower' | 'upper' | 'title' })}>
        <option value="lower">{t('case.lower')}</option><option value="upper">{t('case.upper')}</option><option value="title">{t('case.title')}</option>
      </select>
    </div>
  )

  if (method.type === 'remove') return (
    <div className="editor two-col"><label>{t('label.start')}</label>{number('remove_start', 1, 1)}<label>{t('label.count')}</label>{number('remove_count', 1, 1)}<p className="wide muted">{t('label.remove_help')}</p></div>
  )

  if (method.type === 'remove_pattern') return (
    <div className="editor two-col"><label>{t('label.pattern')}</label><input value={method.remove_pattern || ''} onChange={(e) => update({ remove_pattern: e.target.value })} /><label className="wide checkline"><input type="checkbox" checked={!!method.remove_pattern_regex} onChange={(e) => update({ remove_pattern_regex: e.target.checked })} />{t('label.regex')}</label></div>
  )

  if (method.type === 'prefix_suffix') return (
    <div className="editor two-col"><label>{t('label.prefix')}</label><input value={method.prefix || ''} onChange={(e) => update({ prefix: e.target.value })} /><label>{t('label.suffix')}</label><input value={method.suffix || ''} onChange={(e) => update({ suffix: e.target.value })} /></div>
  )

  if (method.type === 'move') return (
    <div className="editor two-col"><label>{t('label.start')}</label>{number('move_start', 1, 1)}<label>{t('label.count')}</label>{number('move_count', 1, 1)}<label>{t('label.move_to')}</label>{number('move_to', 1, 1)}</div>
  )

  if (method.type === 'swap') return (
    <div className="editor two-col"><label>{t('label.separator')}</label><input value={method.swap_separator || ''} onChange={(e) => update({ swap_separator: e.target.value })} /><label>{t('label.occurrence')}</label>{number('swap_occurrence', 1, 1)}<p className="wide muted">{t('label.swap_example')}</p></div>
  )

  if (method.type === 'trim') return (
    <div className="editor"><label className="checkline"><input type="checkbox" checked={!!method.trim_normalize_spaces} onChange={(e) => update({ trim_normalize_spaces: e.target.checked })} />{t('label.trim_spaces')}</label><p className="muted">{t('label.trim_info')}</p></div>
  )

  if (method.type === 'timestamp') return (
    <div className="editor two-col">
      <label>{t('label.source')}</label><select value={method.timestamp_source || 'modified'} onChange={(e) => update({ timestamp_source: e.target.value as 'modified' | 'batch' })}><option value="modified">{t('label.file_modified_time')}</option><option value="batch">{t('label.batch_time')}</option></select>
      <label>{t('label.format')}</label><input value={method.timestamp_format || ''} onChange={(e) => update({ timestamp_format: e.target.value })} />
      <label>{t('label.position')}</label><select value={method.timestamp_position || 'suffix'} onChange={(e) => update({ timestamp_position: e.target.value as 'prefix' | 'suffix' })}><option value="suffix">{t('label.suffix_position')}</option><option value="prefix">{t('label.prefix_position')}</option></select>
      <label>{t('label.separator')}</label><input value={method.timestamp_separator || ''} onChange={(e) => update({ timestamp_separator: e.target.value })} />
    </div>
  )

  if (method.type === 'list') return (
    <div className="editor"><textarea rows={8} value={method.list_text || ''} onChange={(e) => update({ list_text: e.target.value })} placeholder={t('label.list_info')} /><label className="checkline"><input type="checkbox" checked={!!method.list_include_extension} onChange={(e) => update({ list_include_extension: e.target.checked })} />{t('label.list_ext')}</label></div>
  )

  if (method.type === 'list_replace') return (
    <div className="editor"><textarea rows={8} value={method.list_replace_text || ''} onChange={(e) => update({ list_replace_text: e.target.value })} placeholder={t('label.rules_info')} /><label className="checkline"><input type="checkbox" checked={!!method.list_replace_regex} onChange={(e) => update({ list_replace_regex: e.target.checked })} />{t('label.regex_plural')}</label><label className="checkline"><input type="checkbox" checked={!!method.list_replace_case_sensitive} onChange={(e) => update({ list_replace_case_sensitive: e.target.checked })} />{t('label.case_sensitive')}</label></div>
  )

  if (method.type === 'script') return (
    <div className="editor"><textarea rows={7} value={method.script_expression || ''} onChange={(e) => update({ script_expression: e.target.value })} /><p className="muted">{t('label.variables_help')}</p><p className="muted">{t('label.functions_help')}</p></div>
  )

  return <div className="editor"><p className="muted">{t('method.choose')}</p></div>
}

export default App
